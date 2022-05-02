package observance

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// Config contains all config variables for setting up observability (logging, metrics).
type Config struct {
	AppName              string
	LogLevel             string
	SentryURL            string
	Version              string
	Environment          string
	MetricsURL           string
	MetricsFlushInterval time.Duration
	// LoggedHeaders is map of header names and log field names. If those headers are present in the request,
	// the method CopyWithRequest will add them to the logger with the given field name.
	// E.g. map[string]string{"FastBill-RequestId": "requestId"} means that if the header "FastBill-RequestId" was found
	// in the request headers the value will be added to the logger under the name "requestId".
	LoggedHeaders map[string]string
}

// Obs is a wrapper for all things that helps to observe the operation of
// the service: logging, monitoring, tracing
type Obs struct {
	Logger        Logger
	Metrics       Measurer
	loggedHeaders map[string]string
}

// NewObs creates a new observance instance for logging.
// Optional: If a Sentry URL was provided logs with level error will be sent to Sentry.
// Optional: If a metrics URL was provided a Prometheus Pushgateway metrics can be captured.
func NewObs(config Config) (*Obs, error) {
	log, err := NewLogrus(config)
	if err != nil {
		return nil, err
	}

	obs := &Obs{
		Logger:        log,
		loggedHeaders: config.LoggedHeaders,
	}

	if config.MetricsURL == "" {
		return obs, nil
	}

	metrics := NewPrometheusMetrics(config.MetricsURL, config.AppName, config.MetricsFlushInterval, log)
	obs.Metrics = metrics
	return obs, nil
}

// CopyWithRequest creates a new observance and adds request-specific fields to
// the logger (and maybe at some point to the other parts of observance, too).
// The headers specified in the config (LoggedHeaders) will be added as log fields with their specified field names.
func (o *Obs) CopyWithRequest(r *http.Request) *Obs {
	obCopy := *o
	obs := &obCopy

	obs.Logger = obs.Logger.WithFields(Fields{
		"url":    r.RequestURI,
		"method": r.Method,
	})

	for headerName, fieldName := range o.loggedHeaders {
		headerValue := r.Header.Get(headerName)
		if headerValue != "" {
			obs.Logger = obs.Logger.WithField(fieldName, headerValue)
		}
	}

	return obs
}

// PanicRecover can be used to recover panics in the main thread and log the messages.
func (o *Obs) PanicRecover() {
	if r := recover(); r != nil {
		// According to Russ Cox (leader of the Go team) capturing the stack trace here works:
		// https://groups.google.com/d/msg/golang-nuts/MB8GyW5j2UY/m_YYy7mGYbIJ .
		o.Logger.WithField("stack", string(debug.Stack())).Error(fmt.Sprintf("%v", r))
	}
}
