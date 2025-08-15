package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/csmistry/cointracker/pkg/db"
	"github.com/csmistry/cointracker/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Syncer object holds addresses in cache
type Syncer struct {
	dbClient     *db.DBClient
	queueClient  *queue.QueueClient
	addressCache map[string]*AddressState
}

// AddressState defines state of an address in the wallet
type AddressState struct {
	Address           string
	NextOffset        int
	TotalTransactions int
	Balance           int64
	Synced            bool
}

type Transaction struct {
	TxID      string
	Amount    int64
	Timestamp time.Time
}

// BlockchainResponse is the API response format
type BlockchainResponse struct {
	FinalBalance int64 `json:"final_balance"`
	NTx          int   `json:"n_tx"`
	Transactions []struct {
		Hash string `json:"hash"`
		Time int64  `json:"time"`
		Out  []struct {
			Value int64 `json:"value"`
		} `json:"out"`
	} `json:"txs"`
}

func NewSyncer(dbClient *db.DBClient, queueClient *queue.QueueClient) *Syncer {
	return &Syncer{
		dbClient:     dbClient,
		queueClient:  queueClient,
		addressCache: make(map[string]*AddressState),
	}
}

// HandleAdd synchronizes an address and its transactions
func (s *Syncer) HandleAdd(job queue.Job) {
	state, exists := s.addressCache[job.Address]
	if !exists {
		state = &AddressState{
			Address:    job.Address,
			NextOffset: 0,
			Balance:    0,
			Synced:     false,
		}
		s.addressCache[job.Address] = state
	} else if state.Synced {
		return
	}

	// start sync in a goroutine
	go func(address string) {
		log.Printf("Starting sync for address: %s", address)

		ctx := context.Background()
		offset := 0
		limit := 100

		// fetch until all transactions have synced
		for {
			balance, txs, more, err := s.FetchFromBlockchain(address, offset, limit)
			if err != nil {
				log.Printf("error fetching transactions for address %s: %v", address, err)
				return
			}

			// upsert address with latest balance and next offset
			filter := bson.M{"address": address}
			update := bson.M{
				"$set": bson.M{
					"balance":      balance,
					"tx_count":     offset + len(txs),
					"last_updated": time.Now(),
				},
				"$setOnInsert": bson.M{
					"syncing": true,
				},
			}
			opts := options.Update().SetUpsert(true)

			if _, err := s.dbClient.AddressCollection().UpdateOne(ctx, filter, update, opts); err != nil {
				log.Printf("error updating address %s: %v", address, err)
				return
			}

			// insert transactions
			var docs []interface{}
			for _, t := range txs {

				docs = append(docs, bson.M{
					"address":   address,
					"txid":      t.TxID,
					"amount":    t.Amount,
					"timestamp": t.Timestamp,
				})
			}
			if len(docs) > 0 {
				if _, err := s.dbClient.TransactionCollection().InsertMany(ctx, docs); err != nil {
					log.Printf("error inserting transactions for %s: %v", address, err)
					return
				}
			}

			if !more {
				// Done syncing this address
				if _, err := s.dbClient.AddressCollection().UpdateOne(ctx, filter, bson.M{"$set": bson.M{"syncing": false}}); err != nil {
					log.Printf("error clearing syncing flag for %s: %v", address, err)
				}
				state.Synced = true
				log.Printf("Finished syncing address %s", address)
				return
			}

			offset += limit
			time.Sleep(2 * time.Second) // simple rate limit
		}
	}(job.Address)
}

func (s *Syncer) HandleRemove(job queue.Job) {

}

func (s *Syncer) FetchFromBlockchain(address string, offset, limit int) (int64, []Transaction, bool, error) {
	url := fmt.Sprintf("https://blockchain.info/rawaddr/%s?offset=%d&limit=%d", address, offset, limit)

	resp, err := http.Get(url)
	if err != nil {
		return 0, nil, false, fmt.Errorf("error calling blockchain.info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, nil, false, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var data BlockchainResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, nil, false, fmt.Errorf("error decoding response: %w", err)
	}

	// Convert API txs to Transaction type
	var txs []Transaction
	for _, t := range data.Transactions {
		// tx amount
		totalOut := int64(0)
		for _, out := range t.Out {
			totalOut += out.Value
		}
		txs = append(txs, Transaction{
			TxID:      t.Hash,
			Amount:    totalOut,
			Timestamp: time.Unix(t.Time, 0),
		})
	}

	hasMore := offset+limit < data.NTx
	return data.FinalBalance, txs, hasMore, nil
}
