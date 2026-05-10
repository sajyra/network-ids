package grpc

import (
	"context"
	"fmt"
	"log"

	pb "github.com/sanjay-rajjan/network-ids/proto/ids"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/sanjay-rajjan/network-ids/pkg/types"
)

type IDSClient struct {
	conn   *grpc.ClientConn
	client pb.IDSServiceClient
	nodeID string
}

func NewIDSClient(coordinatorAddr string, nodeID string) (*IDSClient, error) {
	conn, err := grpc.NewClient(coordinatorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	client := pb.NewIDSServiceClient(conn)
	return &IDSClient{
		conn: conn,
		client: client,
		nodeID: nodeID,
	}, nil
}

func (c *IDSClient) StartStreaming(ctx context.Context, events <-chan types.ThreatEvent) error {
	stream, err := c.client.ReportThreat(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return fmt.Errorf("event channel closed")
			}
			pbEvent := &pb.ThreatEvent{
				SourceIp: event.SourceIP,
				Type: string(event.Type),
				Timestamp: event.Timestamp,
				NodeId: c.nodeID,
			}

			if err := stream.Send(pbEvent); err != nil {
				return fmt.Errorf("failed to send event: %w", err)
			}
			log.Printf("[Agent-%s] Sent threat event | IP: %s | Type: %s",
				c.nodeID,
				event.SourceIP,
				string(event.Type),
			)
		
		case <-ctx.Done():
			return nil;
		}
		
	}
}

func (c *IDSClient) Close() {
	c.conn.Close()
}
