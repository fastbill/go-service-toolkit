package observance

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

// TestLogger is an extended Logger interface for testing.
// It allows to check the logs that were recorded.
type TestLogger interface {
	Logger
	LastEntry() TestLogEntry
	Entries() []TestLogEntry
	Reset()
}

// TestLogEntry represents one recorded log entry in the test logger.
type TestLogEntry struct {
	Level   string
	Message string
	Data    map[string]interface{}
}

// LogrusTestLogger implements TestLogger.
type LogrusTestLogger struct {
	LogrusLogger
	result *test.Hook
}

// LastEntry returns the last recorded log entry.
func (l *LogrusTestLogger) LastEntry() TestLogEntry {
	return TestLogEntry{
		Level:   l.result.LastEntry().Level.String(),
		Message: l.result.LastEntry().Message,
		Data:    l.result.LastEntry().Data,
	}
}

// Entries returns all recorded log entries.
func (l *LogrusTestLogger) Entries() []TestLogEntry {
	entries := []TestLogEntry{}
	for _, entry := range l.result.AllEntries() {
		newEntry := TestLogEntry{
			Level:   entry.Level.String(),
			Message: entry.Message,
			Data:    entry.Data,
		}
		entries = append(entries, newEntry)
	}
	return entries
}

// Reset clears the recorded logs to start fresh.
func (l *LogrusTestLogger) Reset() {
	l.result.Reset()
}

// NewTestLogger creates a new TestLogger that can be used to create a test observance instance.
func NewTestLogger() TestLogger {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	return &LogrusTestLogger{
		LogrusLogger: LogrusLogger{
			basicLogger: logger,
			logger:      logger.WithFields(nil),
		},
		result: hook,
	}
}
