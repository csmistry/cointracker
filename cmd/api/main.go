package main

import (
	"fmt"
	"net/http"

	"github.com/csmistry/cointracker/pkg/operations"
	"github.com/go-chi/chi/v5"
)

func main() {

	router := chi.NewRouter()

	// Define routes
	router.Get("/address/{addr}/balance", operations.GetAddressBalance)
	router.Get("/address/{addr}/transactions", operations.GetAddressTransactions)
	router.Post("/address/add", operations.AddAddress)
	router.Post("/address/remove", operations.RemoveAddress)

	fmt.Println("Serving requests on port :8080")
	http.ListenAndServe(":8080", router)
}
