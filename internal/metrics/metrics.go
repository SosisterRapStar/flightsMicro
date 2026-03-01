package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	EventsProccessed    prometheus.Counter
	EventProcessingTime *prometheus.HistogramVec
}

func RegisterMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		EventsProccessed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "app_processed_events_total",
			Help: "The total number of processed events",
		}),
		EventProcessingTime: promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
			Name:    "app_kafka_message_processing_duration_seconds",
			Help:    "Time spent processing Kafka messages in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
		},
			[]string{"topic"},
		),
	}
	return m
}
