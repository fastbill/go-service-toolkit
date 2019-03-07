package observance

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// Measurer defines the generic interface capturing metrics.
type Measurer interface {
	Increment(name string)
	SetGauge(name string, value float64)
	SetGaugeInt64(name string, value int64)
	DurationSince(name string, start time.Time)
}

// PrometheusMetrics is an implementation of Measurer.
type PrometheusMetrics struct {
	registry *prometheus.Registry
	pusher   *push.Pusher
	gauges   map[string]prometheus.Gauge
	counters map[string]prometheus.Counter
	logger   Logger
}

// NewPrometheusMetrics creates a new metrics instance to collect metrics.
func NewPrometheusMetrics(url, appName string, flushInterval time.Duration, logger Logger) *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	pusher := push.New(url, appName).
		Grouping("instance", hostName()).
		Gatherer(registry)

	go continuouslyPush(pusher, flushInterval, logger)

	return &PrometheusMetrics{
		registry: registry,
		pusher:   pusher,
		gauges:   make(map[string]prometheus.Gauge),
		counters: make(map[string]prometheus.Counter),
		logger:   logger,
	}
}

// Increment is used to count occurances. It can only be used for values that never decrease.
func (m *PrometheusMetrics) Increment(name string) {
	counter, ok := m.counters[name]
	if !ok {
		counter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
		})
		m.counters[name] = counter
		m.register(name, counter)
	}
	counter.Inc()
}

// SetGauge is used to track a float64 value over time.
func (m *PrometheusMetrics) SetGauge(name string, value float64) {
	gauge, ok := m.gauges[name]
	if !ok {
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
		})
		m.gauges[name] = gauge
		m.register(name, gauge)
	}
	gauge.Set(value)
}

// SetGaugeInt64 is used to track an int64 value over time.
// The integer value will be converted to a float to fit the prometheus API.
func (m *PrometheusMetrics) SetGaugeInt64(name string, value int64) {
	gauge, ok := m.gauges[name]
	if !ok {
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
		})
		m.gauges[name] = gauge
		m.register(name, gauge)
	}
	gauge.Set(float64(value))
}

// DurationSince is a utility method that accepts a metrics name and start time.
// It then calculates the duration between the start time and now.
// The result is converted to milliseconds and then tracked using SetGauge.
func (m *PrometheusMetrics) DurationSince(name string, start time.Time) {
	durationInMs := float64(time.Since(start).Round(time.Millisecond) / time.Millisecond)
	m.SetGauge(name, durationInMs)
}

func (m *PrometheusMetrics) register(name string, collector prometheus.Collector) {
	if err := m.registry.Register(collector); err != nil {
		m.logger.WithField("metric", name).WithError(err).Error("failed to register metric")
	}
}

// continuouslyPush calls the Add method of pusher periodically so the metrics get pushed to Prometheus.
func continuouslyPush(pusher *push.Pusher, flushInterval time.Duration, logger Logger) {
	for range time.Tick(flushInterval) {
		err := pusher.Add()
		if err != nil {
			logger.WithError(err).Error("failed to push metrics")
		}
	}
}

func hostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostName
}
