package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ThreatsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ids_threats_received_total",
			Help: "Total number of threat events received from agents",
		},
		[]string{"node_id", "threat_type"},
	)

	BlocksPublished = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ids_blocks_published_total",
			Help: "Total number of block commands published to Redis",
		},
	)
)