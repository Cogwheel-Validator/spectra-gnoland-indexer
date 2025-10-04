FROM golang:1.25.1-trixie AS builder

WORKDIR /app

COPY . .

ARG VERSION=unknown
RUN GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") && \
    go build -ldflags="-X main.Commit=${GIT_COMMIT} -X main.Version=${VERSION}" -o indexer indexer/main.go && \
    chmod +x indexer

FROM debian:trixie-slim

WORKDIR /app

COPY --from=builder /app/indexer /app/indexer

ENTRYPOINT ["/app/indexer"]