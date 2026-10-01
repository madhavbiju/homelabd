.PHONY: all build clean test lint docker

BINARY_NAME=homelabd

all: build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/homelabd

clean:
	go clean
	rm -f bin/$(BINARY_NAME)

test:
	go test -v ./...

lint:
	golangci-lint run

docker:
	docker build -t homelabd:latest .
