package grpc

import (
	"fmt"
	"io"
	"log"
	coordinatorredis "github.com/sanjay-rajjan/network-ids/coordinator/redis"
	pb "github.com/sanjay-rajjan/network-ids/proto/ids"
)

type IDSServer struct {
	pb.UnimplementedIDSServiceServer
	publisher *coordinatorredis.Publisher
}

func NewIDSServer(publisher *coordinatorredis.Publisher) *IDSServer {
	return &IDSServer{publisher: publisher}
}

func (s *IDSServer) ReportThreat(stream pb.IDSService_ReportThreatServer) error {
	for {
		event, err := stream.Recv()

		if err == io.EOF {
			return nil;
		}
		if err != nil {
			return fmt.Errorf("error receiving event: %w", err)
		}

		log.Printf("[Coordinator] Threat Received | IP: %s | Type: %s | Node: %s", event.SourceIp, event.Type, event.NodeId)

		
		if err := s.publisher.PublishBlock(stream.Context(), event.SourceIp); err != nil {
			log.Printf("[Coordinator] Failed to publish block: %v", err)
		}
	}

}


