package gobayeux_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sigmavirus24/gobayeux/v2"
	"github.com/sigmavirus24/gobayeux/v2/internal/gobayeuxtest"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name          string
		serverAddress string
		shouldErr     bool
	}{
		{"valid url for server address", "https://example.com", false},
		{"invalid url for server address", "http://192.168.0.%31/", true},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			_, err := gobayeux.NewClient(tc.serverAddress)
			if err != nil && !tc.shouldErr {
				t.Errorf("expected NewClient() to not return an err but it did, %q", err)
			} else if tc.shouldErr && err == nil {
				t.Error("expected NewClient() to err but it didn't")
			}
		})
	}
}

func TestSubscribe(t *testing.T) {
	client, err := gobayeux.NewClient("https://example.com", nil)
	if err != nil {
		t.Fatalf("expected a working client but got an err %q", err)
	}
	client.Subscribe("/foo/bar", nil)
}

func TestUnsubscribe(t *testing.T) {
	server := gobayeuxtest.NewServer(t)
	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server (%v)", err)
	}

	client, err := gobayeux.NewClient(
		"https://example.com",
		gobayeux.WithHTTPTransport(server),
		gobayeux.WithIgnoreError(func(err error) bool { return true }),
	)

	if err != nil {
		t.Fatalf("failed to create client (%v)", err)
	}

	done := make(chan error)
	recv := make(chan interface{})
	msgs := make(chan []gobayeux.Message)
	errs := client.Start(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done)

		count := 0
		for {
			select {
			case ms := <-msgs:
				if count == 0 {
					close(recv)
				}

				count += len(ms)
			case err := <-errs:
				done <- err
				return
			case <-time.After(2 * time.Second):
				// After Unsubscribe we shouldn't receive any more so a timeout
				// with a count greater than zero is success
				if count == 0 {
					done <- fmt.Errorf("timeout with no messages received")
				}

				return
			}
		}
	}()

	client.Subscribe("/foo/bar", msgs)

	// Wait until at least one message is received or fail on timeout
	select {
	case <-recv:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout on recv")
	}

	client.Unsubscribe("/foo/bar")

	// Wait until the worker timeout and done closes without error or timeout
	// and fail
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout on recv")
	}

	if err := client.Disconnect(context.Background()); err != nil {
		t.Fatalf("failed to disconnect (%v)", err)
	}

	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("failed to stop test server (%v)", err)
	}

	wg.Wait()
}

func TestCanDoubleSubscribe(t *testing.T) {
	server := gobayeuxtest.NewServer(t)
	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server (%v)", err)
	}

	client, err := gobayeux.NewClient(
		"https://example.com",
		gobayeux.WithHTTPTransport(server),
		gobayeux.WithIgnoreError(func(err error) bool { return true }),
	)

	if err != nil {
		t.Fatalf("failed to create client (%v)", err)
	}

	done := make(chan error)
	msgs := make(chan []gobayeux.Message)
	errs := client.Start(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done)

		count := 0
		for count < 100 {
			select {
			case ms := <-msgs:
				count += len(ms)
			case err := <-errs:
				if !strings.Contains(err.Error(), "already subscribed") {
					done <- err
					return
				}
			}
		}
	}()

	client.Subscribe("/foo/bar", msgs)
	client.Subscribe("/foo/bar", msgs)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error from client (%v)", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("test timed out")
	}

	if err := client.Disconnect(context.Background()); err != nil {
		t.Fatalf("failed to disconnect (%v)", err)
	}

	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("failed to stop test server (%v)", err)
	}

	wg.Wait()
}

func TestSubscribeWithContext(t *testing.T) {
	t.Run("succeeds with active context", func(t *testing.T) {
		client, err := gobayeux.NewClient("https://example.com", nil)
		if err != nil {
			t.Fatalf("expected a working client but got an err %q", err)
		}

		ctx := context.Background()
		err = client.SubscribeWithContext(ctx, "/foo/bar", nil)
		if err != nil {
			t.Errorf("expected SubscribeWithContext to succeed but got error: %v", err)
		}
	})

	t.Run("respects timeout context", func(t *testing.T) {
		client, err := gobayeux.NewClient("https://example.com", nil)
		if err != nil {
			t.Fatalf("expected a working client but got an err %q", err)
		}

		// Fill the channel to capacity so that the next SubscribewithContext blocks.
		for i := range 10 {
			client.Subscribe(gobayeux.Channel(fmt.Sprintf("/fill/%d", i)), nil)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		time.Sleep(2 * time.Millisecond) // Ensure timeout fires

		err = client.SubscribeWithContext(ctx, "/foo/bar", nil)
		if err == nil {
			t.Error("expected SubscribeWithContext to return an error with expired context")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("expected context.DeadlineExceeded error, got: %v", err)
		}
	})
}

func TestUnsubscribeWithContext(t *testing.T) {
	t.Run("succeeds with active context", func(t *testing.T) {
		client, err := gobayeux.NewClient("https://example.com", nil)
		if err != nil {
			t.Fatalf("expected a working client but got an err %q", err)
		}

		ctx := context.Background()
		err = client.UnsubscribeWithContext(ctx, "/foo/bar")
		if err != nil {
			t.Errorf("expected UnsubscribeWithContext to succeed but got error: %v", err)
		}
	})

	t.Run("respects timeout context", func(t *testing.T) {
		client, err := gobayeux.NewClient("https://example.com", nil)
		if err != nil {
			t.Fatalf("expected a working client but got an err %q", err)
		}

		// Fill the channel to capacity so that the next UnsubscribeWithContext blocks.
		for i := range 10 {
			client.Unsubscribe(gobayeux.Channel(fmt.Sprintf("/fill/%d", i)))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		time.Sleep(2 * time.Millisecond) // Ensure timeout fires

		err = client.UnsubscribeWithContext(ctx, "/foo/bar")
		if err == nil {
			t.Error("expected UnsubscribeWithContext to return an error with expired context")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("expected context.DeadlineExceeded error, got: %v", err)
		}
	})
}

func TestErrorParsing(t *testing.T) {
	server := gobayeuxtest.NewServer(t, gobayeuxtest.WithHandshakeError(true))
	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server (%v)", err)
	}

	client, err := gobayeux.NewClient(
		"https://example.com",
		gobayeux.WithHTTPTransport(server),
	)
	if err != nil {
		t.Fatalf("failed to create client (%v)", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errs := client.Start(ctx)
	err = <-errs
	if err == nil {
		t.Fatal("expected an error when connecting")
	}

	if err.Error() != `expected 200 response from bayeux server, got 400 with status 'Bad Request' and body '"{\"error\":\"Invalid request\"}"'` {
		t.Errorf("expected different error when connecting; got %v", err)
	}

	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("failed to stop test server (%v)", err)
	}
}
