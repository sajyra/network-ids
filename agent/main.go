//go:build linux && arm64

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	agentredis "github.com/sanjay-rajjan/network-ids/agent/redis"
	agentgrpc "github.com/sanjay-rajjan/network-ids/agent/grpc"
	agentebpf "github.com/sanjay-rajjan/network-ids/agent/ebpf"
	"github.com/sanjay-rajjan/network-ids/pkg/types"
)

func main() {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node-1"
	}

	ifaceName := os.Getenv("IFACE")
	if ifaceName == "" {
		ifaceName = "eth0"
	}

	coordinatorAddr := os.Getenv("COORDINATOR_ADDR")
	if coordinatorAddr == "" {
		coordinatorAddr = "localhost:50051"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	log.Printf("[Agent-%s] Starting up on interface %s", nodeID, ifaceName)

	sensor, err := agentebpf.NewSensor(ifaceName)
	if err != nil {
		log.Fatalf("[Agent-%s] Failed to initialize eBPF sensor: %v", nodeID, err)
	}
	defer sensor.Close()
	log.Printf("[Agent-%s] eBPF sensor initialized", nodeID)

	events := make(chan types.ThreatEvent, 100)
	sensor.ReadEvents(events)

	subscriber, err := agentredis.NewSubscriber(redisAddr, nodeID)
	if err != nil {
		log.Fatalf("[Agent-%s] Failed to connect to Redis: %v", nodeID, err)
	}

	client, err := agentgrpc.NewIDSClient(coordinatorAddr, nodeID)
	if err != nil {
		log.Fatalf("[Agent-%s] Failed to connect to coordinator: %v", nodeID, err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("[Agent-%s] Received signal %v, shutting down", nodeID, sig)
		cancel()
	}()

	go func() {
		err := subscriber.Subscribe(ctx, func(ip string) {
			log.Printf("[Agent-%s] Blocking IP: %s", nodeID, ip)
			if err := sensor.BlockIP(ip); err != nil {
				log.Printf("[Agent-%s] Failed to block IP %s: %v", nodeID, ip, err)
			}
		})
		if err != nil {
			log.Printf("[Agent-%s] Subscriber error: %v", nodeID, err)
		}
	}()

	log.Printf("[Agent-%s] Connected to coordinator, starting stream", nodeID)

	if err := client.StartStreaming(ctx, events); err != nil {
		log.Printf("[Agent-%s] Streaming stopped: %v", nodeID, err)
		os.Exit(1)
	}
}

