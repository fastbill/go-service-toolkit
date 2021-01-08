package observance

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

// Logger is a general interface to be implemented for multiple loggers.
type Logger interface {
	Level() string
	Trace(msg interface{})
	Debug(msg interface{})
	Info(msg interface{})
	Warn(msg interface{})
	Error(msg interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
	SetOutput(w io.Writer)
}

// Fields is a type alias to ease reading.
type Fields = map[string]interface{}

// LogrusLogger wraps Logrus to provide an implementation of the Logger interface.
type LogrusLogger struct {
	basicLogger *logrus.Logger
	logger      *logrus.Entry
}

// Level returns the log level that was set for the logger.
// Only entries with that level or above with be logged.
func (l *LogrusLogger) Level() string {
	return l.basicLogger.Level.String()
}

// Trace writes a log entry with level "trace".
func (l *LogrusLogger) Trace(msg interface{}) {
	l.logger.Trace(msg)
}

// Debug writes a log entry with level "debug".
func (l *LogrusLogger) Debug(msg interface{}) {
	l.logger.Debug(msg)
}

// Info writes a log entry with level "info".
func (l *LogrusLogger) Info(msg interface{}) {
	l.logger.Info(msg)
}

// Warn writes a log entry with level "warning".
func (l *LogrusLogger) Warn(msg interface{}) {
	l.logger.Warn(msg)
}

// Error writes a log entry with level "error".
func (l *LogrusLogger) Error(msg interface{}) {
	l.logger.Error(msg)
}

// WithField adds an additional field for logging.
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{
		basicLogger: l.basicLogger,
		logger:      l.logger.WithField(key, value),
	}
}

// WithFields allows to add multiple additional fields to the logging.
// The argument needs to be of type Fields (map[string]interface{}).
func (l *LogrusLogger) WithFields(fields Fields) Logger {
	return &LogrusLogger{
		basicLogger: l.basicLogger,
		logger:      l.logger.WithFields(logrus.Fields(fields)),
	}
}

// WithError adds an error for logging.
func (l *LogrusLogger) WithError(err error) Logger {
	return &LogrusLogger{
		basicLogger: l.basicLogger,
		logger:      l.logger.WithError(err),
	}
}

// SetOutput changes where the logs are written to. The default is Stdout.
func (l *LogrusLogger) SetOutput(w io.Writer) {
	l.basicLogger.SetOutput(w)
}

// NewLogrus creates a Logrus logger that fulfils the Logger interface with Sentry integration.
// All log messages will contain app name, pid and hostname/containerID.
func NewLogrus(logLevel string, appName string, sentryURL string, version string) (Logger, error) {
	logrusLogLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	basicLogger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrusLogLevel,
	}

	if sentryURL != "" {
		levelsToSendToSentry := []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		}

		sentryOpts := sentryOptions{
			Dsn:              sentryURL,
			AttachStacktrace: true,
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				for i := range event.Exception {
					event.Exception[i].Stacktrace.Frames = filterVendorFrames(event.Exception[i].Stacktrace.Frames)
				}
				// Remove the list of all packages of the service. It just spams Sentry.
				event.Modules = make(map[string]string)
				return event
			},
		}

		hook, err := newSentryHook(sentryOpts, levelsToSendToSentry)
		if err != nil {
			return nil, err
		}

		if version != "" {
			hook.SetRelease(version)
		}

		basicLogger.Hooks.Add(hook)
	}

	logger := basicLogger.WithFields(logrus.Fields{
		"name":     appName,
		"pid":      os.Getpid(),
		"hostname": hostname,
	})

	return &LogrusLogger{
		basicLogger: basicLogger,
		logger:      logger,
	}, nil
}

// filterVendorFrames removes frames that belong to the vendor folder from the stack trace.
// That way, the Sentry GUI only shows the responsible line in the actual application code.
// Our Sentry instance runs an older version of Sentry that does not yet provide the option
// to enter stack trace filters in the settings.
func filterVendorFrames(frames []sentry.Frame) []sentry.Frame {
	filteredFrames := make([]sentry.Frame, 0, len(frames))
	for _, frame := range frames {
		if strings.Contains(frame.AbsPath, "vendor") {
			continue
		}
		filteredFrames = append(filteredFrames, frame)
	}
	return filteredFrames
}
