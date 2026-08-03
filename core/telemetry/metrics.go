package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
		[]string{"method", "path", "status"},
	)
	HTTPDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request latency"},
		[]string{"method", "path"},
	)
	DBDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{Name: "db_query_duration_seconds", Help: "DB query latency"},
		[]string{"operation", "table"},
	)
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{Name: "active_connections", Help: "Active connections"},
	)
)
