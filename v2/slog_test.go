//go:build go1.21
// +build go1.21

package gobayeux_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/sigmavirus24/gobayeux/v2"
)

func ExampleWithSlogLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))
	client, err := gobayeux.NewClient("http://localhost:9876", gobayeux.WithSlogLogger(logger))
	if err != nil {
		panic(err)
	}

	errs := client.Start(context.Background())
	err = <-errs
	// Output:
	// level=DEBUG msg=starting at=handshake
	// level=DEBUG msg="error during request" at=handshake error="Post \"http://localhost:9876\": dial tcp 127.0.0.1:9876: connect: connection refused"
}
