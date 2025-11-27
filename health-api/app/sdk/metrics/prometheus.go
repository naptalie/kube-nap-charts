package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for the health API
var (
	// HTTP request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_api_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "health_api_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Application metrics
	HTTPErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_api_http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "path", "code"},
	)

	PanicsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "health_api_panics_total",
			Help: "Total number of panics recovered",
		},
	)

	// Business metrics
	HealthCheckQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_api_health_check_queries_total",
			Help: "Total number of health check queries",
		},
		[]string{"status"},
	)

	GrafanaRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_api_grafana_requests_total",
			Help: "Total number of requests to Grafana API",
		},
		[]string{"endpoint", "status"},
	)

	GrafanaRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "health_api_grafana_request_duration_seconds",
			Help:    "Grafana API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)
