package gobayeux

import "github.com/sirupsen/logrus"

// Logger defines the logging interface gobayeux leverages
type Logger interface {
	// Debug takes a message and any number of arguments and logs them at the
	// debug level
	Debug(msg string, args ...any)

	// Info takes a message and any number of arguments and logs them at the
	// info level
	Info(msg string, args ...any)

	// Warn takes a message and any number of arguments and logs them at the
	// warn level
	Warn(msg string, args ...any)

	// Error takes a message and any number of arguments and logs them at the
	// error level
	Error(msg string, args ...any)

	// WithError returns a new Logger that addes the given error to any log
	// messages emitted
	WithError(error) Logger

	// WithField returns a new Logger that adds the given key/value to any
	// log messages emitted
	WithField(key string, value any) Logger
}

type nullLogger struct {
}

func (*nullLogger) Debug(msg string, args ...any) {
}

func (*nullLogger) Info(msg string, args ...any) {
}

func (*nullLogger) Warn(msg string, args ...any) {
}

func (*nullLogger) Error(msg string, args ...any) {
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

func (w *wrappedFieldLogger) Debug(msg string, args ...any) {
	w.FieldLogger.Debug(append([]any{msg}, args...))
}

func (w *wrappedFieldLogger) Info(msg string, args ...any) {
	w.FieldLogger.Info(append([]any{msg}, args...))
}

func (w *wrappedFieldLogger) Warn(msg string, args ...any) {
	w.FieldLogger.Warn(append([]any{msg}, args...))
}

func (w *wrappedFieldLogger) Error(msg string, args ...any) {
	w.FieldLogger.Error(append([]any{msg}, args...))
}

func (w *wrappedFieldLogger) WithError(err error) Logger {
	return &wrappedFieldLogger{w.FieldLogger.WithError(err)}
}

func (w *wrappedFieldLogger) WithField(key string, value any) Logger {
	return &wrappedFieldLogger{w.FieldLogger.WithField(key, value)}
}
