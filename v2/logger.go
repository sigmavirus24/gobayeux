package gobayeux

import "github.com/sirupsen/logrus"

// Logger defines the logging interface gobayeux leverages
type Logger interface {
	// Debug takes any number of arguments and logs them at the debug level
	Debug(args ...any)

	// Info takes any number of arguments and logs them at the info level
	Info(args ...any)

	// Warn takes any number of arguments and logs them at the info level
	Warn(args ...any)

	// Error takes any number of arguments and logs them at the info level
	Error(args ...any)

	// WithError returns a new Logger that addes the given error to any log
	// messages emitted
	WithError(error) Logger

	// WithField returns a new Logger that adds the given key/value to any
	// log messages emitted
	WithField(key string, value any) Logger
}

type nullLogger struct {
}

func (*nullLogger) Debug(args ...any) {
}

func (*nullLogger) Info(args ...any) {
}

func (*nullLogger) Warn(args ...any) {
}

func (*nullLogger) Error(args ...any) {
}

func (l *nullLogger) WithError(err error) Logger {
	return l
}

func (l *nullLogger) WithField(key string, value any) Logger {
	return l
}

func newNullLogger() *nullLogger {
	return &nullLogger{}
}

type wrappedFieldLogger struct {
	logrus.FieldLogger
}

func (w *wrappedFieldLogger) WithError(err error) Logger {
	return &wrappedFieldLogger{w.FieldLogger.WithError(err)}
}

func (w *wrappedFieldLogger) WithField(key string, value any) Logger {
	return &wrappedFieldLogger{w.FieldLogger.WithField(key, value)}
}
