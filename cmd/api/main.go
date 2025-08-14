package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/csmistry/cointracker/pkg/db"
	"github.com/csmistry/cointracker/pkg/operations"
	"github.com/go-chi/chi/v5"
)

func main() {

	// Connect to DB
	dbClient, err := db.InitDB()
	if err != nil {
		log.Fatalf("failed to create db client: %v", err)
	}

	err = dbClient.Ping()
	if err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("Connected to db")

	// Define routes
	router := chi.NewRouter()

	router.Get("/address/{addr}/balance", operations.GetAddressBalance)
	router.Get("/address/{addr}/transactions", operations.GetAddressTransactions)
	router.Post("/address/add", operations.AddAddress)
	router.Post("/address/remove", operations.RemoveAddress)

	fmt.Println("Serving requests on port :8080")
	http.ListenAndServe(":8080", router)
}
