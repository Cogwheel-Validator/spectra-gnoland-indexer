.PHONY: build-indexer install-indexer clean experimental-indexer experimental-install-indexer

build-indexer:
	mkdir -p build
	go build -o build/indexer indexer/main.go

install-indexer:
	go install indexer/main.go

clean:
	rm -rf build

# experimental build with greentea garbage collection
# use at your own risk
experimental-indexer:
	GOEXPERIMENT=greentea go build -o build/indexer indexer/main.go

# experimental install with greentea garbage collection
# use at your own risk
experimental-install-indexer:
	GOEXPERIMENT=greentea go install indexer/main.go