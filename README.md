# Network Intrusion Detection System

A distributed network intrusion detection system using eBPF/XDP for kernel-level packet inspection, gRPC for threat streaming, and Redis pub/sub for cluster-wide IP blocking.

## Architecture

```
[ Any Node: eBPF/XDP ] ──gRPC──▶ [ Coordinator ] ──Redis pub/sub──▶ [ All Nodes: XDP Block ]
```

Each agent runs an eBPF/XDP program inside the Linux kernel that inspects every incoming packet before the network stack processes it. Detected threats stream to a central coordinator over gRPC. The coordinator publishes blocked IPs to Redis, which fans out to all agents simultaneously for cluster-wide enforcement.

## Performance

- **483K+ packets/sec** XDP kernel hook throughput, validated via pktgen load testing on AWS EC2 c6g instances
- **1.16ms block propagation latency** from eBPF kernel detection through gRPC, coordinator, Redis pub/sub, to XDP enforcement
- Cross-node blocking confirmed across separate AWS EC2 nodes

## Detection Algorithms

- **SYN Flood** — tracks SYN packets per source IP via BPF hash maps, threshold-based flagging
- **Port Scan** — counts distinct destination ports per source IP
- **SSH Brute Force** — monitors connection attempts to port 22

## Tech Stack

| Component | Technology |
|-----------|------------|
| Packet inspection | eBPF/XDP, C |
| Agent & Coordinator | Go |
| Agent-Coordinator RPC | gRPC + Protocol Buffers |
| Cluster propagation | Redis pub/sub |
| Metrics | Prometheus, Grafana |
| Orchestration | Docker Compose |
| eBPF Go bindings | cilium/ebpf-go |
| Load testing | pktgen, hping3 |
| Deployment | AWS EC2 (c6g.large, ARM64) |

## Project Structure

```
agent/
  ebpf/        # eBPF sensor — loads XDP program, reads BPF maps
  grpc/        # gRPC client — streams threat events to coordinator
  redis/       # Redis subscriber — receives block commands
coordinator/
  grpc/        # gRPC server — receives threat streams from agents
  redis/       # Redis publisher — publishes block commands
  metrics/     # Prometheus metrics
proto/         # Protocol Buffer definitions
docker/        # Docker Compose, Prometheus config, Dockerfile
```

## Running Locally

### Prerequisites

- Docker Desktop
- Lima (Linux VM for eBPF on Mac)
- Go 1.22+

### Start infrastructure (Mac)

```bash
cd docker
docker compose up -d
```

Starts coordinator, Redis, Prometheus, and Grafana via Docker Compose.

### Run agent (requires Linux kernel for eBPF)

```bash
limactl shell ebpf-dev
cd /path/to/network-ids
go build -o /tmp/agent ./agent/
sudo NODE_ID=node-1 IFACE=eth0 \
    COORDINATOR_ADDR=192.168.5.2:50051 \
    REDIS_ADDR=192.168.5.2:6379 \
    METRICS_PORT=9091 \
    /tmp/agent
```

### Regenerate eBPF bindings (after modifying sensor.c)

Run inside Lima:

```bash
cd agent/ebpf
GOPACKAGE=ebpf bpf2go -target arm64 -output-dir . sensor \
    bpf/sensor.c -- -O2 -g -D__TARGET_ARCH_arm64 \
    -I/usr/include/aarch64-linux-gnu
```

### Observability

| Service | URL |
|---------|-----|
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9095 |
| Coordinator metrics | http://localhost:9090/metrics |
| Agent metrics | http://localhost:9091/metrics |

## AWS Deployment

Deployed on AWS EC2 c6g.large instances (ARM64, Ubuntu 24.04).

Agent nodes require Linux kernel 5.4+ with eBPF support. XDP runs in generic mode on the AWS ENA driver — native XDP requires bare metal with supported NIC drivers (Intel i40e, Mellanox ConnectX).

### Throughput Testing

Load tested with the Linux kernel pktgen module generating 483K+ pps at 60-byte packet size between two c6g.large instances in the same AWS VPC. Propagation latency measured via Prometheus histograms with nanosecond-precision eBPF kernel timestamps.

### Deployment Architecture

```
                    AWS us-east-1
    c6g.large (Node 1)          c6g.large (Node 2)
  ┌──────────────────┐        ┌──────────────────┐
  │   eBPF/XDP       │        │   eBPF/XDP       │
  │   Agent          │        │   Agent          │
  └────────┬─────────┘        └────────┬─────────┘
           │ gRPC                      │ gRPC
           └──────────┬────────────────┘
                      ▼
             ┌─────────────────┐
             │   Coordinator   │
             │   + Redis       │
             │   + Prometheus  │
             │   + Grafana     │
             └─────────────────┘
```

