package queue

import (
	"fmt"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type QueueClient struct {
	conn  *amqp091.Connection
	queue amqp091.Queue
}

// InitQueue initializes a queue
func InitQueue() (*QueueClient, error) {
	uri := os.Getenv("RABBITMQ_URI")

	var conn *amqp091.Connection
	var channel *amqp091.Channel
	var queue amqp091.Queue
	var err error

	// retry connection
	for i := 0; i < 5; i++ {
		conn, err = amqp091.Dial(uri)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	// test channel creation
	channel, err = conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}
	defer channel.Close()

	// create queue
	queue, err = channel.QueueDeclare(
		"sync_jobs",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	return &QueueClient{
		conn:  conn,
		queue: queue,
	}, nil
}

// Ping checks if a queue connection is active
func (qc *QueueClient) Ping() error {
	c, err := qc.conn.Channel()
	if err != nil {
		return err
	}
	c.Close()
	return nil
}
