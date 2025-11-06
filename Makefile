.PHONY: build run test clean

build:
	go build -o bin/pulsardb cmd/pulsardb/main.go

run:
	go run cmd/pulsardb/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf data/

dev:
	go run cmd/pulsardb/main.go -config config.dev.json

lint:
	golangci-lint run

fmt:
	go fmt ./...

