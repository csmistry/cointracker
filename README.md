# cointracker

## Design
![IMG_2742](https://github.com/user-attachments/assets/c1f7f19a-c748-44d7-852f-00857782789b)



## System Overview

This service tracks cryptocurrency wallet activity (balances and transactions) for a set of addresses. It is composed of four main components:  

## 1. API Server
- **Purpose:** Exposes HTTP endpoints for clients to add, remove, and query wallet addresses.
- **Key Endpoints:**
  - `POST /address/add` — Add an address to be tracked.
  - `POST /address/remove` — Remove (or archive) an address from tracking.
  - `GET /address/{addr}/balance` — Get the current balance for an address.
  - `GET /address/{addr}/transactions?offset=0?limit=10` — Get paginated transaction history for an address.
- **Responsibilities:**
  - Push new address sync or remove jobs into the **Queue**.
  - Query the **Database** for balances and transactions.

---

## 2. Queue (RabbitMQ)
- **Purpose:** Decouples incoming requests from the background syncing process.
- **Responsibilities:**
  - Store pending “sync” jobs when new addresses are added.
  - Ensure the **Syncer** processes addresses in order without overwhelming external APIs.
- **Implementation Detail:** RabbitMQ is used as the message broker to reliably handle job delivery and persistence.

---

## 3. Syncer
- **Purpose:** Background worker that processes jobs from the **Queue**.
- **Responsibilities:**
  - Fetch address details (balance, transactions) from external blockchain APIs.
  - Handle pagination and rate limits by spreading API calls over time.
  - Write or update address and transaction records in the **Database**.
  - Support re-syncing of archived addresses

---

## 4. Database (MongoDB)
- **Purpose:** Persistent storage of tracked addresses and their transactions.
- **Collections and Fields:**

### `addresses`
| Field           | Type      | Description |
|-----------------|-----------|-------------|
| `address`       | string    | Wallet address (unique) |
| `balance`       | int64     | Latest balance fetched from blockchain |
| `tx_count`      | int       | Number of transactions synced so far |
| `last_updated`  | datetime  | Timestamp of the last sync |
| `archived`      | bool      | Whether the address has been removed/archived |
| `syncing`       | bool      | Indicates if the address is currently being synced |

- Uses **upsert** logic to create or update an address when new data arrives.

### `transactions`
| Field        | Type      | Description |
|--------------|-----------|-------------|
| `address`    | string    | Wallet address associated with this transaction |
| `txid`       | string    | Transaction ID |
| `amount`     | int64     | Transaction amount |
| `timestamp`  | datetime  | When the transaction occurred |

- New transactions are **inserted in bulk** as they are fetched from the blockchain API.

## Setup

This project uses Docker Compose to run all services locally. The stack includes:

- **MongoDB** (port `27017`)
- **API Server** (port `8080`)
- **RabbitMQ** (port `5672`)
- **Syncer** (background worker)

### Steps

1. Make sure you have [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/) installed.

2. From the repo root, start all services:

```bash
docker-compose up --build
```

## Testing

Sample curl commands to test functionality:

#### Add Address
```
curl -X POST http://localhost:8080/address/add \
     -H "Content-Type: application/json" \
     -d '{"address":"3E8ociqZa9mZUSwGdSmAEMAoAxBK3FNDcd"}'
```

#### Get Address Balance
```
curl -X GET http://localhost:8080/address/3E8ociqZa9mZUSwGdSmAEMAoAxBK3FNDcd/balance
```

#### Get Address Transactions
```
curl -X GET http://localhost:8080/address/3E8ociqZa9mZUSwGdSmAEMAoAxBK3FNDcd/transactions?offset=0&limit=10
```

#### Remove (archive) Address
```
curl -X POST http://localhost:8080/address/remove \
     -H "Content-Type: application/json" \
     -d '{"address":"3E8ociqZa9mZUSwGdSmAEMAoAxBK3FNDcd"}'
```

## Further Improvements
  - Handle retries and backoff for failed sync attempts
  - Make docker images smaller using a separate builder stage
  - Better error handling
  - More clearly defined API response structure
  - Add Job ID to enable job tracking functionality
  - Stronger use of in-memory address cache
  - Database efficiences (indexes)
  - Background job for address removal (delete from DB if address has been archived for a certain amount of time)
