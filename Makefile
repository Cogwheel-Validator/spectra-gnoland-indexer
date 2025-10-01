.PHONY: build install clean build-experimental install-experimental

build:
	mkdir -p build
	go build -o build/indexer indexer/main.go

install:
	cd indexer && go install ./...

clean:
	rm -rf build

# experimental build with greentea garbage collection
# use at your own risk
build-experimental:
	GOEXPERIMENT=greenteagc go build -o build/indexer-tea indexer/main.go

# experimental install with greentea garbage collection
# use at your own risk
install-experimental:
	cd indexer && GOEXPERIMENT=greenteagc go install ./...