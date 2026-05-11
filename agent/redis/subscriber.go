package redis

import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	"time"
	"github.com/redis/go-redis/v9"
	agentmetrics "github.com/sanjay-rajjan/network-ids/agent/metrics"
)

type BlockMessage struct {
	SourceIP string `json:"source_ip"`
	DetectedAt int64 `json:"detected_at"`
}

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
			var blockMsg BlockMessage
			if err := json.Unmarshal([]byte(msg.Payload), &blockMsg); err != nil {
				log.Printf("[Agent-%s] Failed to parse block message: %v", s.nodeID, err)
				continue
			}

			now := time.Now().UnixNano()
			latencySeconds := float64(now - blockMsg.DetectedAt) / float64(time.Second)
			agentmetrics.PropagationLatency.Observe(latencySeconds)

			log.Printf("[Agent-%s] Block command received from Redis | IP: %s | Latency: %.2fms", s.nodeID, blockMsg.SourceIP, latencySeconds*1000)
			onBlock(blockMsg.SourceIP)

		case <-ctx.Done():
			log.Printf("[Agent-%s] Subscriber shutting down", s.nodeID)
			return nil
		}
	}
}