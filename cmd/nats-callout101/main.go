package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/snapp-incubator/nats-callout101/internal/authenticator"
)

const (
	natsURL  = "nats://auth:auth@127.0.0.1:4222"
	nkeySeed = "SAANDLKMXL6CUS3CP52WIXBEDN6YJ545GDKC65U5JZPPV6WH6ESWUA6YAI"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	auth, err := authenticator.New(logger, authenticator.Config{
		URL:      natsURL,
		NkeySeed: nkeySeed,
		Users: map[string]authenticator.User{
			"admin": {
				Password: "admin",
				Account:  "APP",
			},
		},
	})
	if err != nil {
		log.Fatalf("failed to create authenticator service %s", err)
	}

	if err := auth.Start(); err != nil {
		log.Fatalf("authenticator service subscription failed %s", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Received shutdown signal, exiting...")

	if err := auth.Stop(); err != nil {
		log.Fatalf("authenticator service unsubscription failed %s", err)
	}

	log.Println("Bye bye!")
}
