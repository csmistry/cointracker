package operations

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/csmistry/cointracker/pkg/db"
	"github.com/csmistry/cointracker/pkg/queue"
)

type Server struct {
	DBClient    *db.DBClient
	QueueClient *queue.QueueClient
}

type AddressRequest struct {
	Address string `json:"address"`
}

// GetAddressBalance returns the balance for a bitcoin address
func (s *Server) GetAddressBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetAddressTransactions returns transactions for a specific bitcoin address
func (s *Server) GetAddressTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// AddAddress adds a new address to the wallet
func (s *Server) AddAddress(w http.ResponseWriter, r *http.Request) {
	var req AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}

	job := queue.Job{
		Type:    "ADD",
		Address: req.Address,
	}

	if err := s.QueueClient.PublishJob(job); err != nil {
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}
	log.Println("Enqueued ADD Job for address:", req.Address)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Enqueued ADD Job for address: %s", req.Address)))
}

// RemoveAddress removes an address from the wallet
func (s *Server) RemoveAddress(w http.ResponseWriter, r *http.Request) {
	var req AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}

	job := queue.Job{
		Type:    "REMOVE",
		Address: req.Address,
	}

	if err := s.QueueClient.PublishJob(job); err != nil {
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}
	log.Println("Enqueued REMOVE Job for address:", req.Address)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Enqueued REMOVE Job for address: %s", req.Address)))
}
