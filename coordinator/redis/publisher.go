package redis

import (
	"context"
	"fmt"
	"encoding/json"
	"log"
	"github.com/redis/go-redis/v9"
	"github.com/sanjay-rajjan/network-ids/coordinator/metrics"
)

type BlockMessage struct {
	SourceIP string `json:"source_ip"`
	DetectedAt int64 `json:"detected_at"`
}

type Publisher struct {
	client *redis.Client
}

func NewPublisher(addr string) (*Publisher, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	log.Printf("[Publisher] Connected to Redis at %s", addr)
	return &Publisher{client: client}, nil
}

func (p *Publisher) PublishBlock(ctx context.Context, sourceIP string, detectedAt int64) error {
	msg := BlockMessage{
		SourceIP: sourceIP,
		DetectedAt: detectedAt,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal block message: %w", err)
	}
	
	err = p.client.Publish(ctx, "blocked-ips", payload).Err()
	if err != nil {
		return fmt.Errorf("failed to publish block command: %w", err)
	}

	metrics.BlocksPublished.Inc()
	log.Printf("[Publisher] Published block for IP: %s", sourceIP)

	return nil
}