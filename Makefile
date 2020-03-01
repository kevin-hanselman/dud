.PHONY: build docker_build test test-int %-test-cov bench cover fmt clean shell

DOCKER = docker run --rm -v '$(shell pwd):/src' go_dev

build:
	go build ./...

docker_build:
	docker build -t go_dev .

test: fmt
	golint ./...
	go test -short ./...

test-int: test
	go test -run Integration ./...

bench: test
	go test ./... -bench .

%-test-cov: %-test-cov.out
	go tool cover -html=$<

unit-test-cov.out:
	go test -short ./... -coverprofile=$@

int-test-cov.out:
	go test -run Integration ./... -coverprofile=$@

all-test-cov.out:
	go test ./... -coverprofile=$@

fmt:
	go fmt ./...

clean:
	rm -f *.out
	go clean ./...
