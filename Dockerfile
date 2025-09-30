FROM golang:1.25.1-bullseye AS builder

WORKDIR /app

COPY . .

RUN go build -o indexer indexer/main.go

RUN chmod +x indexer

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/indexer /app/indexer

ENTRYPOINT ["/app/indexer"]