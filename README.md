# Network Intrusion Detection System

Network IDS using eBPF for kernel-level packet inspection and gRPC for threat reporting.

### Tech Stack
Go, gRPC, eBPF, Redis, Prometheus, Grafana

### Architecture
Each agent runs an eBPF/XDP program inside the Linux kernel that inspects every incoming packet before the network stack processes it. 
Detected threats stream to a central coordinator over gRPC. The coordinator publishes blocked IPs to Redis, which propagates to all agents simultaneously for cluster wide enforcement.


### Detection Algorithms
SYN Flood - tracks SYN packets per Source IP via BPF maps, with threshold based flagging   
Port Scan - counts distinct destination ports per Source IP     
SSH Brute Force - monitors connection attempts to Port 22


