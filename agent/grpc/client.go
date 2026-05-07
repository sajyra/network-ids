package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/sanjay-rajjan/network-ids/proto/ids"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (c *IDSClient) StartStreaming(ctx context.Context) error {
	stream, err := c.client.ReportThreat(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				log.Printf("[Agent-%s] Stream closed: %v", c.nodeID, err)
				return
			}
			log.Printf("[Agent-%s] Block command received | IP: %s | IssuedAt: %d",
				c.nodeID,
				cmd.SourceIp,
				cmd.IssuedAt,
			)
		}
	}()

	for {
		event := &pb.ThreatEvent{
			SourceIp:  "192.168.1.100",
			Type:      "SYN_FLOOD",
			Timestamp: time.Now().UnixNano(),
			NodeId:    c.nodeID,
		}

		if err := stream.Send(event); err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}

		log.Printf("[AGENT-%s] Sent threat event | IP: %s | Type: %s",
			c.nodeID,
			event.SourceIp,
			event.Type,
		)

		time.Sleep(2 * time.Second)
	}
}

func (c *IDSClient) Close() {
	c.conn.Close()
}
