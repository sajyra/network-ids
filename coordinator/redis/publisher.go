package redis

import (
	"context"
	"fmt"
	"log"
	"github.com/redis/go-redis/v9"
)

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

func (p *Publisher) PublishBlock(ctx context.Context, sourceIP string) error {
	err := p.client.Publish(ctx, "blocked-ips", sourceIP).Err()
	if err != nil {
		return fmt.Errorf("failed to publish block command: %w", err)
	}

	log.Printf("[Publisher] Published block for IP: %s", sourceIP)
	return nil
}