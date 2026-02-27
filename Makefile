.PHONY: build test

build:
	go build -o "build/foreman-builder"

test:
	go test ./...