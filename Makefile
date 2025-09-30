.PHONY: build install clean build-experimental install-experimental

build:
	mkdir -p build
	go build -o build/indexer indexer/main.go

install:
	go install indexer/main.go

clean:
	rm -rf build

# experimental build with greentea garbage collection
# use at your own risk
build-experimental:
	GOEXPERIMENT=greenteagc go build -o build/indexer indexer/main.go

# experimental install with greentea garbage collection
# use at your own risk
install-experimental:
	GOEXPERIMENT=greenteagc go install indexer/main.go