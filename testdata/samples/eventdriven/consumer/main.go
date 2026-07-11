package main

import (
	"os"

	"github.com/example/queue"
)

func main() {
	brokerURL := os.Getenv("BROKER_URL")
	queue.Connect(brokerURL)

	queue.Consume("orders.created", handleOrderCreated)
}

func handleOrderCreated(payload string) {
	processOrder(payload)
}

func processOrder(payload string) {
	// business logic placeholder
}