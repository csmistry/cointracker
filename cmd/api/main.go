package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/csmistry/cointracker/pkg/db"
	"github.com/csmistry/cointracker/pkg/operations"
	"github.com/csmistry/cointracker/pkg/queue"
	"github.com/go-chi/chi/v5"
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

	s := &operations.Server{
		DBClient:    dbClient,
		QueueClient: queueClient,
	}

	// Define routes
	router := chi.NewRouter()

	router.Get("/address/{addr}/balance", s.GetAddressBalance)
	router.Get("/address/{addr}/transactions", s.GetAddressTransactions)
	router.Post("/address/add", s.AddAddress)
	router.Post("/address/remove", s.RemoveAddress)

	fmt.Println("Serving requests on port :8080")
	http.ListenAndServe(":8080", router)
}
