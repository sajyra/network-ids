package main

import (
	"log"
	"net"
	"google.golang.org/grpc"
	coordinatorgrpc "github.com/sanjay-rajjan/network-ids/coordinator/grpc"
	pb "github.com/sanjay-rajjan/network-ids/proto/ids"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer();
	idsServer := coordinatorgrpc.NewIDSServer()
	pb.RegisterIDSServiceServer(grpcServer, idsServer)
	log.Println("[COORDINATOR] Listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
