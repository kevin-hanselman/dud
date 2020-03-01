.PHONY: build docker_build test test-integration bench cover fmt clean shell

DOCKER = docker run --rm -v '$(shell pwd):/src' go_dev

build:
	go build

docker_build:
	docker build -t go_dev .

test: fmt
	golint ./...
	go test -short ./...

test-integration: docker_build test
	$(DOCKER) go test -run Integration ./...

bench: test
	go test ./... -bench .

unit-test-cov: unit-test-cov.out
	go tool cover -html=unit-test-cov.out

# TODO: DRY this out
int-test-cov: int-test-cov.out
	go tool cover -html=int-test-cov.out

unit-test-cov.out:
	go test -short ./... -coverprofile=unit-test-cov.out

int-test-cov.out: docker_build
	$(DOCKER) go test -run Integration ./... -coverprofile=int-test-cov.out

fmt:
	go fmt ./...

clean:
	rm -f cover.out
	go clean ./...
