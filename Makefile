.PHONY: build test

build:
	go build -o ./build/

clean:
	go clean -testcache
test:
	go test ./...

e2e-test:
	go test -tags=e2e -v -timeout 60m ./...
