package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	TodosCompletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todos_completed_total",
			Help: "Total number of todos completed",
		},
		[]string{"user_id"},
	)

	SyncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sync_duration_seconds",
			Help:    "Sync operation duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "status"},
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)
)

func RecordHTTPRequest(method, path string, status int, duration float64) {
	statusStr := ""
	switch {
	case status >= 200 && status < 300:
		statusStr = "2xx"
	case status >= 300 && status < 400:
		statusStr = "3xx"
	case status >= 400 && status < 500:
		statusStr = "4xx"
	case status >= 500:
		statusStr = "5xx"
	default:
		statusStr = "unknown"
	}
	HTTPRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration)
}

func RecordTodoCompleted(userID string) {
	TodosCompletedTotal.WithLabelValues(userID).Inc()
}

func RecordSyncDuration(operation, status string, duration float64) {
	SyncDuration.WithLabelValues(operation, status).Observe(duration)
}

func SetActiveUsers(count float64) {
	ActiveUsers.Set(count)
}
