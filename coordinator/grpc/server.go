package grpc

import (
	"fmt"
	"io"
	"log"
	coordinatorredis "github.com/sanjay-rajjan/network-ids/coordinator/redis"
	pb "github.com/sanjay-rajjan/network-ids/proto/ids"
	"github.com/sanjay-rajjan/network-ids/coordinator/metrics"
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

		metrics.ThreatsReceived.WithLabelValues(event.NodeId, event.Type).Inc()
		log.Printf("[Coordinator] Threat Received | IP: %s | Type: %s | Node: %s", event.SourceIp, event.Type, event.NodeId)
		
		if err := s.publisher.PublishBlock(stream.Context(), event.SourceIp, event.Timestamp); err != nil {
			log.Printf("[Coordinator] Failed to publish block: %v", err)
		}
	}

}


