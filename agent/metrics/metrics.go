package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ThreatsDetected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ids_threats_detected_total",
			Help: "Total number of threats detected by the eBPF sensor",
		},
		[]string{"threat_type"},
	)

	PropagationLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "ids_propagation_latency_seconds",
			Help: "Time between threat detection and block command received",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		},
	)
)