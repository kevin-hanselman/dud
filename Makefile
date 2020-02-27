.PHONY: build test fmt clean

build: test
	go build

test: fmt
	golint ./...
	go test ./...

bench: test
	go test ./... -bench .

fmt:
	go fmt ./...

clean:
	go clean ./...
