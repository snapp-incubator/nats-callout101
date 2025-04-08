package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

const natsURL = "nats://auth:auth@127.0.0.1:4222"

func handler(msg *nats.Msg) {
	log.Printf("Received authentication request on subject: %s, reply: %s", msg.Subject, msg.Reply)

	log.Println(string(msg.Data))
}

func main() {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}

	if _, err := nc.Subscribe("$SYS.REQ.USER.AUTH", handler); err != nil {
		log.Fatalf("Error subscribing to authentication subjec: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("Received shutdown signal, exiting...")
}
