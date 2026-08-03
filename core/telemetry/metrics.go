package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests by method, path, status"},
		[]string{"method", "path", "status_code"},
	)
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request latency"},
		[]string{"method", "path"},
	)
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{Name: "db_query_duration_seconds", Help: "DB query latency"},
		[]string{"operation", "table"},
	)
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{Name: "active_connections", Help: "Current active connections"},
	)
)
