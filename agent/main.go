package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	agentgrpc "github.com/sanjay-rajjan/network-ids/agent/grpc"
)

func main() {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node-1"
	}

	log.Printf("[AGENT-%s] Starting up", nodeID)

	client, err := agentgrpc.NewIDSClient("localhost:50051", nodeID)
	if err != nil {
		log.Fatalf("[AGENT-%s] Failed to connect to coordinator: %v", nodeID, err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("[AGENT-%s] Received signal %v, shutting down", nodeID, sig)
		cancel()
	}()

	log.Printf("[AGENT-%s] Connected to coordinator, starting stream", nodeID)
	if err := client.StartStreaming(ctx); err != nil {
		log.Printf("[AGENT-%s] Streaming stopped: %v", nodeID, err)
		os.Exit(1)
	}
}

