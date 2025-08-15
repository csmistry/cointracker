package operations

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/csmistry/cointracker/pkg/db"
	"github.com/csmistry/cointracker/pkg/queue"
	"github.com/csmistry/cointracker/pkg/syncer"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	dBClient    *db.DBClient
	queueClient *queue.QueueClient
}

type AddressRequest struct {
	Address string `json:"address"`
}

func NewServer(dbClient *db.DBClient, queueClient *queue.QueueClient) *Server {
	return &Server{
		dBClient:    dbClient,
		queueClient: queueClient,
	}
}

// GetAddressBalance returns the balance for a bitcoin address
func (s *Server) GetAddressBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	addr := chi.URLParam(r, "addr")

	addrColl := s.dBClient.AddressCollection()
	var addrDoc struct {
		Balance int64 `bson:"balance"`
	}

	err := addrColl.FindOne(ctx, bson.M{"address": addr}).Decode(&addrDoc)
	if err != nil {
		http.Error(w, "address not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{
		"address": addr,
		"balance": addrDoc.Balance,
	})
}

// GetAddressTransactions returns transactions for a specific bitcoin address
func (s *Server) GetAddressTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	addr := chi.URLParam(r, "addr")

	// Pagination query params
	limit := 100
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	txColl := s.dBClient.TransactionCollection()
	cur, err := txColl.Find(ctx,
		bson.M{"address": addr},
		options.Find().SetSort(bson.M{"timestamp": -1}).SetSkip(int64(offset)).SetLimit(int64(limit)),
	)
	if err != nil {
		http.Error(w, "failed to fetch transactions", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var txs []syncer.Transaction
	if err := cur.All(ctx, &txs); err != nil {
		http.Error(w, "failed to decode transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{
		"address":      addr,
		"transactions": txs,
		"limit":        limit,
		"offset":       offset,
	})
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

	if err := s.queueClient.PublishJob(job); err != nil {
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

	if err := s.queueClient.PublishJob(job); err != nil {
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}
	log.Println("Enqueued REMOVE Job for address:", req.Address)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Enqueued REMOVE Job for address: %s", req.Address)))
}
