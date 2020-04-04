.PHONY: build docker_build test test-int %-test-cov bench cover fmt clean tidy loc depgraph

DOCKER = docker run --rm -v '$(shell pwd):/src' go_dev

build:
	go build ./...

docker_build:
	docker build -t go_dev .

test: fmt
	go vet ./...
	golint -set_exit_status ./...
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
	gofmt -s -w .

clean:
	rm -f *.out
	go clean ./...

tidy:
	go mod tidy -v

loc:
	tokei --sort lines
	tokei --sort lines --exclude "*_test.go"

depgraph:
	godepgraph -nostdlib $(wildcard **/*.go) | dot -Tpng -o depgraph.png
