package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"google.golang.org/grpc"
	coordinatorgrpc "github.com/sanjay-rajjan/network-ids/coordinator/grpc"
	coordinatorredis "github.com/sanjay-rajjan/network-ids/coordinator/redis"
	pb "github.com/sanjay-rajjan/network-ids/proto/ids"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	publisher, err := coordinatorredis.NewPublisher(redisAddr)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	log.Println("[Coordinator] Connected to Redis")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer();
	idsServer := coordinatorgrpc.NewIDSServer(publisher)
	pb.RegisterIDSServiceServer(grpcServer, idsServer)
	log.Println("[Coordinator] Listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
