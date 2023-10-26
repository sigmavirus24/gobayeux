//go:build go1.21
// +build go1.21

package gobayeux

import "log/slog"

type wrappedSlog struct {
	*slog.Logger
}

func (w *wrappedSlog) WithError(err error) Logger {
	return w.WithField("error", err)
}

func (w *wrappedSlog) WithField(key string, value any) Logger {
	return &wrappedSlog{w.With(slog.Any(key, value))}
}

func WithSlogLogger(logger *slog.Logger) Option {
	return func(options *Options) {
		options.Logger = &wrappedSlog{logger}
	}
}
