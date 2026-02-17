package services

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// --- HTTP Infrastructure Metrics ---

	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by status code and method.",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response latency (seconds) of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Current number of active connections.",
		},
	)

	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Histogram of database query duration (seconds).",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// --- Business-Domain Metrics (TASK-178) ---

	QueryExecutionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "query_executions_total",
			Help: "Total SQL query executions by connection type and result status.",
		},
		[]string{"connection_type", "status", "query_type"},
	)

	QueryExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "query_execution_duration_seconds",
			Help:    "Histogram of user query execution time in seconds.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"connection_type", "query_type"},
	)

	DashboardViewsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dashboard_views_total",
			Help: "Total dashboard view events by dashboard ID.",
		},
		[]string{"dashboard_id"},
	)

	AIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_requests_total",
			Help: "Total AI service requests by type and status.",
		},
		[]string{"type", "status"},
	)

	AIRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_request_duration_seconds",
			Help:    "Histogram of AI request latency in seconds.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
		},
		[]string{"type"},
	)

	AlertEvaluationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alert_evaluations_total",
			Help: "Total alert evaluations by result (triggered, ok, error).",
		},
		[]string{"result"},
	)

	ConnectionPoolActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connection_pool_active",
			Help: "Number of active connections in each database pool.",
		},
		[]string{"connection_id", "connection_type"},
	)

	ConnectionPoolIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connection_pool_idle",
			Help: "Number of idle connections in each database pool.",
		},
		[]string{"connection_id", "connection_type"},
	)

	WebSocketConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Current number of active WebSocket connections.",
		},
	)

	CacheHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total cache hits by cache type.",
		},
		[]string{"cache_type"},
	)

	CacheMissesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total cache misses by cache type.",
		},
		[]string{"cache_type"},
	)

	AuthAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total authentication attempts by method and result.",
		},
		[]string{"method", "result"},
	)

	ScheduledJobExecutionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduled_job_executions_total",
			Help: "Total scheduled job executions by job type and result.",
		},
		[]string{"job_type", "result"},
	)

	ScheduledJobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scheduled_job_duration_seconds",
			Help:    "Histogram of scheduled job execution duration.",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300},
		},
		[]string{"job_type"},
	)
)

// InitMetrics registers all Prometheus metrics collectors with the default registry.
// Must be called once during application startup before any metrics are recorded.
func InitMetrics() {
	// Infrastructure metrics
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(ActiveConnections)
	prometheus.MustRegister(DBQueryDuration)

	// Business-domain metrics (TASK-178)
	prometheus.MustRegister(QueryExecutionsTotal)
	prometheus.MustRegister(QueryExecutionDuration)
	prometheus.MustRegister(DashboardViewsTotal)
	prometheus.MustRegister(AIRequestsTotal)
	prometheus.MustRegister(AIRequestDuration)
	prometheus.MustRegister(AlertEvaluationsTotal)
	prometheus.MustRegister(ConnectionPoolActive)
	prometheus.MustRegister(ConnectionPoolIdle)
	prometheus.MustRegister(WebSocketConnectionsActive)
	prometheus.MustRegister(CacheHitsTotal)
	prometheus.MustRegister(CacheMissesTotal)
	prometheus.MustRegister(AuthAttemptsTotal)
	prometheus.MustRegister(ScheduledJobExecutionsTotal)
	prometheus.MustRegister(ScheduledJobDuration)
}

// RecordQueryExecution records a user query execution metric.
func RecordQueryExecution(connectionType, status, queryType string, durationSeconds float64) {
	QueryExecutionsTotal.WithLabelValues(connectionType, status, queryType).Inc()
	QueryExecutionDuration.WithLabelValues(connectionType, queryType).Observe(durationSeconds)
}

// RecordDashboardView increments the dashboard view counter.
func RecordDashboardView(dashboardID string) {
	DashboardViewsTotal.WithLabelValues(dashboardID).Inc()
}

// RecordAIRequest records an AI service request metric.
func RecordAIRequest(requestType, status string, durationSeconds float64) {
	AIRequestsTotal.WithLabelValues(requestType, status).Inc()
	AIRequestDuration.WithLabelValues(requestType).Observe(durationSeconds)
}

// RecordAlertEvaluation records an alert evaluation result.
func RecordAlertEvaluation(result string) {
	AlertEvaluationsTotal.WithLabelValues(result).Inc()
}

// RecordAuthAttempt records an authentication attempt.
func RecordAuthAttempt(method, result string) {
	AuthAttemptsTotal.WithLabelValues(method, result).Inc()
}

// RecordCacheHit increments the cache hit counter for a given cache type.
func RecordCacheHit(cacheType string) {
	CacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss increments the cache miss counter for a given cache type.
func RecordCacheMiss(cacheType string) {
	CacheMissesTotal.WithLabelValues(cacheType).Inc()
}

// RecordScheduledJobExecution records a scheduled job execution metric.
func RecordScheduledJobExecution(jobType, result string, durationSeconds float64) {
	ScheduledJobExecutionsTotal.WithLabelValues(jobType, result).Inc()
	ScheduledJobDuration.WithLabelValues(jobType).Observe(durationSeconds)
}

// UpdateConnectionPoolMetrics updates the connection pool gauge metrics.
func UpdateConnectionPoolMetrics(connectionID, connectionType string, active, idle int) {
	ConnectionPoolActive.WithLabelValues(connectionID, connectionType).Set(float64(active))
	ConnectionPoolIdle.WithLabelValues(connectionID, connectionType).Set(float64(idle))
}
