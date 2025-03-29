package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsMiddleware collects Prometheus metrics for HTTP requests
func (a *App) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		
		c.Next()		
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method
		
		// Record metrics
		a.Metrics.RequestDuration.WithLabelValues(method, path).Observe(duration)
		a.Metrics.RequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
	}
}


type PrometheusMetrics struct {
	// General metrics
	UpSince         prometheus.Gauge
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	
	// Participant metrics
	ParticipantsActive    prometheus.Gauge
	ParticipantsCheckedIn *prometheus.CounterVec
	ParticipantsLeft      *prometheus.CounterVec
	CheckpointCrossings   *prometheus.CounterVec
	
	// Error metrics
	QRFailures      *prometheus.CounterVec
	AuthFailures    *prometheus.CounterVec
	DatabaseErrors  *prometheus.CounterVec
}

// all Prometheus metrics
func prom() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		// General application metrics
		UpSince: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "eventloop_up_since_seconds",
				Help: "The timestamp when the application started",
			},
		),
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventloop_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "eventloop_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		
		// Business metrics
		ParticipantsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "eventloop_participants_active",
				Help: "Current number of active participants (checked in but not checked out)",
			},
		),
		ParticipantsCheckedIn: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventloop_participants_checkin_total",
				Help: "Total number of participant check-ins",
			},
			[]string{"team"},
		),
		ParticipantsLeft: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventloop_participants_checkout_total",
				Help: "Total number of participant check-outs",
			},
			[]string{"team"},
		),
		CheckpointCrossings: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventloop_checkpoint_crossings_total",
				Help: "Total number of checkpoint crossings",
			},
			[]string{"checkpoint"},
		),
		
		// Error metrics
		QRFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventloop_qr_failures_total",
				Help: "Total number of QR failure occurrences",
			},
			[]string{"reason"},
		),
		AuthFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventloop_auth_failures_total",
				Help: "Total number of authentication failures",
			},
			[]string{"role", "reason"},
		),
		DatabaseErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventloop_db_errors_total",
				Help: "Total number of database errors",
			},
			[]string{"operation", "error_type"},
		),
	}

	metrics.UpSince.SetToCurrentTime()

	// Register all metrics
	prometheus.MustRegister(
		metrics.UpSince,
		metrics.RequestsTotal,
		metrics.RequestDuration,
		metrics.ParticipantsActive,
		metrics.ParticipantsCheckedIn,
		metrics.ParticipantsLeft,
		metrics.CheckpointCrossings,
		metrics.QRFailures,
		metrics.AuthFailures,
		metrics.DatabaseErrors,
	)

	return metrics
}

func (a *App) InitializeMetrics() {
	a.Metrics.UpSince.Set(float64(a.AppStart.Unix()))
}
