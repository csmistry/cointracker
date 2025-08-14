FROM golang:1.24
WORKDIR /app

# Copy go.mod / go.sum from root
COPY go.mod go.sum ./
RUN go mod download

# Copy everything from root
COPY . .

# Build sync service binary
RUN go build -o sync ./cmd/sync

ENTRYPOINT ["./sync"]