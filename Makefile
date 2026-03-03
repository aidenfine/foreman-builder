.PHONY: build test

build:
	go build -o ./build/

test:
	go test ./...