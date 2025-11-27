FROM golang:1.25.4-trixie AS builder

WORKDIR /app

COPY . .

ARG GIT_COMMIT=""
ARG GIT_TAG=""
ARG GIT_BRANCH=""
ARG VERSION=""

RUN if [ -z "$VERSION" ]; then \
    if [ -n "$GIT_TAG" ]; then VERSION="$GIT_TAG"; \
    else VERSION="${GIT_BRANCH}-${GIT_COMMIT}"; \
    fi; \
    fi && \
    go build -ldflags="-X 'github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=${GIT_COMMIT}' -X 'github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=${VERSION}'" -o indexer ./indexer

RUN chmod +x indexer

FROM debian:trixie-slim

WORKDIR /app

COPY --from=builder /app/indexer /app/indexer

ENTRYPOINT ["/app/indexer"]