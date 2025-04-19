package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

const (
	natsURL  = "nats://auth:auth@127.0.0.1:4222"
	nkeySeed = "SAANDLKMXL6CUS3CP52WIXBEDN6YJ545GDKC65U5JZPPV6WH6ESWUA6YAI"
)

func main() {
	if _, err := nc.Subscribe("$SYS.REQ.USER.AUTH", handler); err != nil {
		log.Fatalf("Error subscribing to authentication subjec: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("Received shutdown signal, exiting...")
}
