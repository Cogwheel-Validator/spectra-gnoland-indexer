.PHONY: build install clean build-experimental install-experimental build-api test-race integration-test test

########################################################
# Build and install the indexer
########################################################


# Get git information
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
VERSION := $(if $(GIT_TAG),$(GIT_TAG),$(GIT_BRANCH)-$(GIT_COMMIT))

build:
	mkdir -p build
	go build -ldflags="-X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=$(GIT_COMMIT) -X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=$(VERSION)" -o build/indexer indexer/main.go

install:
	go install -ldflags="-X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=$(GIT_COMMIT) -X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=$(VERSION)" indexer/main.go

build-api:
	mkdir -p build
	go build -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(VERSION)" -o build/api api/main.go

clean:
	rm -rf build

# experimental build with greentea garbage collection
# use at your own risk
build-experimental:
	GOEXPERIMENT=greenteagc go build -ldflags="-X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=$(GIT_COMMIT) -X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=$(VERSION)" -o build/indexer-tea indexer/main.go

# experimental install with greentea garbage collection
# use at your own risk
install-experimental:
	cd indexer && GOEXPERIMENT=greenteagc go install ./... -ldflags="-X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=$(GIT_COMMIT) -X github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=$(VERSION)"

########################################################
# Test the indexer
########################################################

test:
	go test -v ./...

test-race:
	go test -v -race ./...

integration-test:
	cd integration && go test -v -tags=integration -timeout=20m ./...