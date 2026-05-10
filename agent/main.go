package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	agentredis "github.com/sanjay-rajjan/network-ids/agent/redis"
	agentgrpc "github.com/sanjay-rajjan/network-ids/agent/grpc"
)

func main() {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node-1"
	}

	log.Printf("[Agent-%s] Starting up", nodeID)

	subscriber, err := agentredis.NewSubscriber("localhost:6379", nodeID)
	if err != nil {
		log.Fatalf("[Agent-%s] Failed to connect to Redis: %v", nodeID, err)
	}

	client, err := agentgrpc.NewIDSClient("localhost:50051", nodeID)
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
		})
		if err != nil {
			log.Printf("[Agent-%s] Subscriber error: %v", nodeID, err)
		}
	}()

	log.Printf("[Agent-%s] Connected to coordinator, starting stream", nodeID)

	if err := client.StartStreaming(ctx); err != nil {
		log.Printf("[Agent-%s] Streaming stopped: %v", nodeID, err)
		os.Exit(1)
	}
}

