package observance

import (
	"net/http"
	"time"
)

// Config contains all config variables for setting up observability (logging, metrics).
type Config struct {
	AppName              string
	LogLevel             string
	SentryURL            string
	Version              string
	MetricsURL           string
	MetricsFlushInterval time.Duration
}

// Obs is a wrapper for all things that helps to observe the operation of
// the service: logging, monitoring, tracing
type Obs struct {
	Logger  Logger
	Metrics Measurer
}

// NewObs creates a new observance instance for logging.
// Optional: If a Sentry URL was provided logs with level error will be sent to Sentry.
// Optional: If a metrics URL was provided a Prometheus Pushgateway metrics can be captured.
func NewObs(config Config) (*Obs, error) {
	log, err := NewLogrus(config.LogLevel, config.AppName, config.SentryURL, config.Version)
	if err != nil {
		return nil, err
	}

	if config.MetricsURL == "" {
		return &Obs{Logger: log}, nil
	}

	metrics := NewPrometheusMetrics(config.MetricsURL, config.AppName, config.MetricsFlushInterval, log)
	return &Obs{
		Logger:  log,
		Metrics: metrics,
	}, nil
}

// CopyWithRequest creates a new observance and adds request-specific fields to
// the logger (and maybe at some point to the other parts of observance, too)
// Logs url, method, requestId and accountId (if present)
func (o *Obs) CopyWithRequest(r *http.Request) *Obs {
	obCopy := *o
	obs := &obCopy

	obs.Logger = obs.Logger.WithFields(Fields{
		"url":       r.RequestURI,
		"method":    r.Method,
		"requestId": r.Header.Get("Fastbill-Outer-RequestId"),
	})

	accountID := r.Header.Get("Fastbill-AccountId")
	if accountID != "" {
		obs.Logger = obs.Logger.WithField("accountId", accountID)
	}

	return obs
}
