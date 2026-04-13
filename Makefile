.PHONY: build test

build:
	go build -o ./build/

clean:
	go clean -testcache

test:
	go test ./...

integration-test:
	go test -tags=integration -v -timeout 60m ./...
