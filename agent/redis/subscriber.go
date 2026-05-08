package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	client *redis.Client
	nodeID string
}

func NewSubscriber(addr string, nodeID string) (*Subscriber, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	log.Printf("[Agent-%s] Subscriber connected to Redis at %s", nodeID, addr)

	return &Subscriber{
		client: client,
		nodeID: nodeID,
	}, nil
}

func (s *Subscriber) Subscribe(ctx context.Context, onBlock func(ip string)) error {
	pubsub := s.client.Subscribe(ctx, "blocked-ips")
	defer pubsub.Close()

	log.Printf("[Agent-%s] Subscribed to blocked-ips channel", s.nodeID)

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			log.Printf("[Agent-%s] Block command received from Redis | IP: %s", s.nodeID, msg.Payload)
			onBlock(msg.Payload)
		case <-ctx.Done():
			log.Printf("[Agent-%s] Subscriber shutting down", s.nodeID)
			return nil
		}
	}
}