VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build install test clean

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o bin/merchant-cli ./cmd/merchant-cli

install:
	go install -ldflags "-s -w -X main.version=$(VERSION)" ./cmd/merchant-cli

test:
	go test -v ./...

clean:
	rm -rf bin/
