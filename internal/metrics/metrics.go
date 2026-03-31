package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "url_monitor"

type Metrics struct {
	ChecksTotal        *prometheus.CounterVec
	CheckDuration      *prometheus.HistogramVec
	InflightChecks     prometheus.Gauge
	QueueSize          prometheus.Gauge
	RequestErrorsTotal *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		ChecksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "checks_total",
				Help:      "Total number of completed monitor checks.",
			},
			[]string{"status"},
		),
		CheckDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "check_duration_seconds",
				Help:      "Duration of monitor checks.",
				Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
			},
			[]string{"status"},
		),
		InflightChecks: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "inflight_checks",
				Help:      "Number of inflight checks.",
			},
		),
		QueueSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "job_queue_size",
				Help:      "Current size of the job queue.",
			},
		),
		RequestErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "request_errors_total",
				Help:      "Total number of request errors by kind",
			},
			[]string{"kind"},
		),
	}

	reg.MustRegister(
		m.ChecksTotal,
		m.CheckDuration,
		m.InflightChecks,
		m.QueueSize,
		m.RequestErrorsTotal,
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

func (m *Metrics) ObserveCheck(status string, d time.Duration) {
	m.ChecksTotal.WithLabelValues(status).Inc()
	m.CheckDuration.WithLabelValues(status).Observe(d.Seconds())
}

func (m *Metrics) IncInflight() {
	m.InflightChecks.Inc()
}

func (m *Metrics) DecInflight() {
	m.InflightChecks.Dec()
}

func (m *Metrics) SetQueueSize(size int) {
	m.QueueSize.Set(float64(size))
}

func (m *Metrics) IncRequestErrors(kind string) {
	m.RequestErrorsTotal.WithLabelValues(kind).Inc()
}
