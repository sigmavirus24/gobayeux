//go:build go1.21
// +build go1.21

package gobayeux_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/sigmavirus24/gobayeux/v2"
)

type roundTripFn func(*http.Request) (*http.Response, error)

func (fn roundTripFn) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

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

	handler := roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
		}, nil
	})

	client, err := gobayeux.NewClient("http://127.0.0.1:9876",
		gobayeux.WithSlogLogger(logger),
		gobayeux.WithHTTPTransport(handler),
	)
	if err != nil {
		panic(err)
	}

	errs := client.Start(context.Background())
	err = <-errs
	if err == nil {
		panic("expected an error when connecting")
	}
	// Output:
	// level=DEBUG msg=starting at=handshake
	// level=DEBUG msg="error parsing response" at=handshake error=EOF
}
