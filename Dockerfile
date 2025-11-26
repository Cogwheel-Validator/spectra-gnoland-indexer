FROM golang:1.25.4-trixie AS builder

WORKDIR /app

COPY . .

ARG GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
ARG GIT_TAG = $(shell git describe --tags --exact-match 2>/dev/null || echo "")
ARG GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
ARG VERSION = $(if $(GIT_TAG),$(GIT_TAG),$(GIT_BRANCH)-$(GIT_COMMIT))

RUN go install -ldflags="-X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=$(GIT_COMMIT) -X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=$(VERSION)" indexer/indexer.go && \
    chmod +x indexer

FROM debian:trixie-slim

WORKDIR /app

COPY --from=builder /app/indexer /app/indexer

ENTRYPOINT ["/app/indexer"]