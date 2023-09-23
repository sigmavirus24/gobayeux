package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"

	gobayeux "github.com/sigmavirus24/gobayeux/v2"
)

type config struct {
	Hostname    string
	Port        uint
	EventBuffer uint
	Protocol    string
	Path        string
	LogLevel    string
	AccessToken string
}

func main() {
	var level logrus.Level
	var cfg config
	flags := flag.NewFlagSet("t", flag.ExitOnError)
	flags.StringVar(&cfg.Protocol, "protocol", "https", "the protocol to use (http or https)")
	flags.UintVar(&cfg.Port, "port", 80, "the port used to connect to the Bayeux server")
	flags.UintVar(&cfg.EventBuffer, "buffer", 100, "the number of events to buffer")
	flags.StringVar(&cfg.Hostname, "hostname", "", "the hostname to connect to")
	flags.StringVar(&cfg.Path, "path", "", "the path used to connect to bayeux")
	flags.StringVar(&cfg.LogLevel, "loglevel", "error", "the level to log at")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("error parsing flags: %q\n", err)
		os.Exit(1)
	}
	channelNames := flags.Args()
	output := make(chan []gobayeux.Message, cfg.EventBuffer)
	logger := logrus.New()

	switch cfg.LogLevel {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	case "fatal":
	default:
		// Let's just skip panic as an option here
		level = logrus.PanicLevel
	}
	logger.SetLevel(level)

	u := url.URL{Scheme: cfg.Protocol, Host: fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port), Path: cfg.Path}
	client, err := gobayeux.NewClient(u.String(), gobayeux.WithLogger(logger))
	if err != nil {
		fmt.Printf("error initializing client: %q\n", err)
		os.Exit(1)
	}
	logger.Debug("got client")

	ctx := context.Background()
	//ctx, cancel := context.WithDeadline(ctx, time.Now().Add(60*time.Second))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errc := client.Start(ctx)

	for _, name := range channelNames {
		client.Subscribe(gobayeux.Channel(name), output)
	}

	for {
		select {
		case err := <-errc:
			fmt.Printf("error in bayeux client: %q\n", err)
			os.Exit(2)
		case ms := <-output:
			for _, m := range ms {
				logger.WithFields(logrus.Fields{
					"channel": m.Channel,
					"data":    string(m.Data),
				}).Info()
			}
		// nolint:staticcheck
		default:
		}
	}
}
