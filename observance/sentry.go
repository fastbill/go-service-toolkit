package observance

import (
	"errors"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

// TODO Once an official Logrus Sentry integration is provided (https://github.com/getsentry/sentry-go/issues/43)
// or the existing logrus Sentry integration is updated to the official client, we should use that instead of the code in this file.
// The code here is a modified version of https://github.com/onrik/logrus/blob/d75d1852818a603e95398104fe3a6dfc96421dd3/sentry/sentry.go

var (
	logrusLevelsToSentryLevels = map[logrus.Level]sentry.Level{
		logrus.PanicLevel: sentry.LevelFatal,
		logrus.FatalLevel: sentry.LevelFatal,
		logrus.ErrorLevel: sentry.LevelError,
		logrus.WarnLevel:  sentry.LevelWarning,
		logrus.InfoLevel:  sentry.LevelInfo,
		logrus.DebugLevel: sentry.LevelDebug,
		logrus.TraceLevel: sentry.LevelDebug,
	}
)

type sentryOptions sentry.ClientOptions

type sentryHook struct {
	client       *sentry.Client
	levels       []logrus.Level
	tags         map[string]string
	release      string
	environment  string
	prefix       string
	flushTimeout time.Duration
}

func (hook *sentryHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *sentryHook) Fire(entry *logrus.Entry) error {
	exceptions := []sentry.Exception{}

	err, ok := entry.Data[logrus.ErrorKey].(error)
	if !ok && entry.Message != "" {
		// This allows to have a stack trace, even though there was only a message provided.
		err = errors.New(entry.Message)
	}

	if err == nil {
		return nil
	}

	stacktrace := sentry.ExtractStacktrace(err)
	if stacktrace == nil {
		stacktrace = sentry.NewStacktrace()
	}
	exceptions = append(exceptions, sentry.Exception{
		Type:       entry.Message,
		Value:      err.Error(),
		Stacktrace: stacktrace,
	})

	event := sentry.Event{
		Level:       logrusLevelsToSentryLevels[entry.Level],
		Message:     hook.prefix + entry.Message,
		Extra:       map[string]interface{}(entry.Data),
		Tags:        hook.tags,
		Environment: hook.environment,
		Release:     hook.release,
		Exception:   exceptions,
	}

	hub := sentry.CurrentHub()
	hook.client.CaptureEvent(&event, nil, hub.Scope())

	return nil
}

func (hook *sentryHook) SetPrefix(prefix string) {
	hook.prefix = prefix
}

func (hook *sentryHook) SetTags(tags map[string]string) {
	hook.tags = tags
}

func (hook *sentryHook) AddTag(key, value string) {
	hook.tags[key] = value
}

func (hook *sentryHook) SetRelease(release string) {
	hook.release = release
}

func (hook *sentryHook) SetEnvironment(environment string) {
	hook.environment = environment
}

func (hook *sentryHook) SetFlushTimeout(timeout time.Duration) {
	hook.flushTimeout = timeout
}

func (hook *sentryHook) Flush() {
	hook.client.Flush(hook.flushTimeout)
}

func newSentryHook(options sentryOptions, levels []logrus.Level) (*sentryHook, error) {
	client, err := sentry.NewClient(sentry.ClientOptions(options))
	if err != nil {
		return nil, err
	}

	hook := sentryHook{
		client:       client,
		levels:       levels,
		tags:         map[string]string{},
		flushTimeout: 5 * time.Second,
	}

	return &hook, nil
}
