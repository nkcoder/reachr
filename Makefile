.PHONY: build test lint tidy run

build:
	go build -o reachr .

test:
	go test ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

run: build
	./reachr --help
