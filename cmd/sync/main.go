package main

import (
	"log"

	"github.com/csmistry/cointracker/pkg/db"
	"github.com/csmistry/cointracker/pkg/queue"
)

func main() {
	// connect to DB
	dbClient, err := db.InitDB()
	if err != nil {
		log.Fatalf("failed to create db client: %v", err)
	}

	err = dbClient.Ping()
	if err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("Connected to db")

	// connect to rabbitMQ
	queueClient, err := queue.InitQueue()
	if err != nil {
		log.Fatal("failed to connect to queue", err)
	}

	err = queueClient.Ping()
	if err != nil {
		log.Fatalf("failed to ping queue: %v", err)
	}
	log.Println("Connected to queue")

	conn := queueClient.Conn()
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to create consumer channel: %v", err)
	}

	// messages will be received on channel msgs
	msgs, err := channel.Consume(
		queue.JOB_QUEUE,
		"",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("failed to consume messages: %v", err)
	}
	log.Println("Sync service started...")

	// process messages as long as channel remains open
	for msg := range msgs {
		log.Println("Received job:", string(msg.Body))
	}

	log.Println("Sync service stopped")
}
