.PHONY: build docker_build test test-int %-test-cov bench fmt clean tidy loc mocks depgraph hyperfine

DOCKER = docker run --rm -v '$(shell pwd):/src' go_dev

build: test
	go build -o duc

docker_build:
	docker build -t go_dev .

test: fmt
	go vet ./...
	go test -short ./...
	golint ./...

test-int: test
	go test -run Integration ./...

bench: test
	go test ./... -benchmem -bench .

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
	rm -f *.out depgraph.png mockery
	go clean ./...

tidy:
	go mod tidy -v

loc:
	tokei --sort lines
	tokei --sort lines --exclude "*_test.go"

mockery:
	curl -L https://github.com/vektra/mockery/releases/download/v1.1.2/mockery_1.1.2_Linux_x86_64.tar.gz | tar -zxvf - mockery

mocks: mockery
	./mockery -all

depgraph:
	godepgraph -nostdlib $(wildcard **/*.go) | dot -Tpng -o depgraph.png

50mb_random.bin:
	dd if=/dev/urandom of=$@ bs=1M count=50

hyperfine: build 50mb_random.bin
	hyperfine 'sha1sum 50mb_random.bin'
	hyperfine 'md5sum 50mb_random.bin'
	hyperfine 'sha256sum 50mb_random.bin'
	hyperfine -L bufsize 1000,10000,100000,1000000,10000000 './duc checksum -b{bufsize} 50mb_random.bin'
