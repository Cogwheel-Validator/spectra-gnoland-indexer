.PHONY: build install clean build-experimental install-experimental build-api

# Get git information
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
VERSION := $(if $(GIT_TAG),$(GIT_TAG),dev-$(GIT_COMMIT))

build:
	mkdir -p build
	go build -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(VERSION)" -o build/indexer indexer/main.go

install:
	cd indexer && go install ./... -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

build-api:
	mkdir -p build
	go build -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(VERSION)" -o build/api api/main.go

clean:
	rm -rf build

# experimental build with greentea garbage collection
# use at your own risk
build-experimental:
	GOEXPERIMENT=greenteagc go build -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(VERSION)" -o build/indexer-tea indexer/main.go

# experimental install with greentea garbage collection
# use at your own risk
install-experimental:
	cd indexer && GOEXPERIMENT=greenteagc go install ./... -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(VERSION)"