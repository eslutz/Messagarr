package api

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricsOnce   sync.Once
	globalMetrics *prometheusMetrics
)

type prometheusMetrics struct {
	// HTTP metrics
	requestDuration *prometheus.HistogramVec
	requestTotal    *prometheus.CounterVec

	// Notification metrics
	notificationsTotal     prometheus.Counter
	notificationsFailed    prometheus.Counter
	notificationsByChannel *prometheus.CounterVec
	notificationDuration   *prometheus.HistogramVec

	// System metrics
	uptime prometheus.Gauge
}

func newPrometheusMetrics() *prometheusMetrics {
	metricsOnce.Do(func() {
		globalMetrics = &prometheusMetrics{
			requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "messagarr_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			}, []string{"path", "method", "code"}),

			requestTotal: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "messagarr_http_requests_total",
				Help: "Total HTTP requests processed.",
			}, []string{"path", "method", "code"}),

			notificationsTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "messagarr_notifications_total",
				Help: "Total number of notifications sent.",
			}),

			notificationsFailed: promauto.NewCounter(prometheus.CounterOpts{
				Name: "messagarr_notifications_failed_total",
				Help: "Total number of failed notifications.",
			}),

			notificationsByChannel: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "messagarr_notifications_by_channel_total",
				Help: "Total notifications sent per channel.",
			}, []string{"channel", "success"}),

			notificationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "messagarr_notification_duration_seconds",
				Help:    "Notification processing duration in seconds.",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			}, []string{"priority"}),

			uptime: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "messagarr_uptime_seconds",
				Help: "Service uptime in seconds.",
			}),
		}
	})
	return globalMetrics
}

func (m *prometheusMetrics) observeRequest(path, method string, code int, duration time.Duration) {
	codeStr := strconv.Itoa(code)
	m.requestDuration.WithLabelValues(path, method, codeStr).Observe(duration.Seconds())
	m.requestTotal.WithLabelValues(path, method, codeStr).Inc()
}

func (m *prometheusMetrics) observeNotification(priority string, duration time.Duration, success bool, channelResults map[string]bool) {
	m.notificationsTotal.Inc()

	if !success {
		m.notificationsFailed.Inc()
	}

	for channel, channelSuccess := range channelResults {
		successStr := "true"
		if !channelSuccess {
			successStr = "false"
		}
		m.notificationsByChannel.WithLabelValues(channel, successStr).Inc()
	}

	m.notificationDuration.WithLabelValues(priority).Observe(duration.Seconds())
}

func (m *prometheusMetrics) updateUptime(startTime time.Time) {
	m.uptime.Set(time.Since(startTime).Seconds())
}
