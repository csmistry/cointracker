FROM golang:1.24 AS builder
WORKDIR /app

# Copy go.mod / go.sum from root
COPY go.mod go.sum ./
RUN go mod download

# Copy everything from root
COPY . .

# Build api server binary
RUN go build -o api ./cmd/api

EXPOSE 8080
ENTRYPOINT ["./api"]