package grpc

import (
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/sanjay-rajjan/network-ids/proto/ids"
)

type IDSServer struct {
	pb.UnimplementedIDSServiceServer
}

func NewIDSServer() *IDSServer {
	return &IDSServer{}
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

		cmd := &pb.BlockCommand{
			SourceIp: event.SourceIp, 
			IssuedAt: time.Now().UnixNano(),
		}
		if err := stream.Send(cmd); err != nil {
			return fmt.Errorf("Error sending block command: %w", err)
		}
	}

}


