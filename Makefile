.PHONY: build test fmt clean

build: test
	go build

test: fmt
	golint ./...
	go test ./...

fmt:
	go fmt ./...

clean:
	go clean ./...
