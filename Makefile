.PHONY: build test bench cover fmt clean

build: test
	go build

test: fmt
	golint ./...
	go test ./...

bench: test
	go test ./... -bench .

cover: cover.out
	go tool cover -html=cover.out

cover.out:
	go test ./... -coverprofile=cover.out

fmt:
	go fmt ./...

clean:
	rm -f cover.out
	go clean ./...
