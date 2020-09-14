.PHONY: fmt lint test test-all %-test-cov clean tidy loc mocks hyperfine build-benchmark

duc: test-all
	go build -o duc

fmt:
	goimports -w .
	gofmt -s -w .

lint:
	go vet ./...
	golint ./...

test: fmt lint
	go test -short ./...

test-all: fmt lint
	go test -race ./...

bench: test
	go test ./... -benchmem -bench .

%-test-cov: %-test.coverage
	go tool cover -html=$<

unit-test.coverage:
	go test -short ./... -coverprofile=$@

int-test.coverage:
	go test -run Integration ./... -coverprofile=$@

all-test.coverage:
	go test ./... -coverprofile=$@

clean:
	rm -f *.coverage *.bin depgraph.png mockery
	go clean ./...

tidy:
	go mod tidy -v

loc:
	tokei --sort lines --exclude 'mocks/'
	tokei --sort lines --exclude 'mocks/' --exclude '*_test.go'

mockery:
	curl -L https://github.com/vektra/mockery/releases/download/v2.2.1/mockery_2.2.1_Linux_x86_64.tar.gz \
		| tar -zxvf - mockery

mocks: mockery
	./mockery --all

# The awk command removes all graph edge definitions that don't include duc
depgraph.png:
	godepgraph -nostdlib . \
		| awk '/^[^"]/ || /duc/ {print;}' \
		| dot -Tpng -o $@

%mb_random.bin:
	dd if=/dev/urandom of=$@ bs=1M count=$(patsubst %mb_random.bin,%,$@)

hyperfine: 50mb_random.bin duc
	hyperfine -L cmd sha1sum,md5sum,sha256sum,b2sum,xxh64sum,'./duc checksum' \
		'{cmd} $<'
	hyperfine -L bufsize 1000,10000,100000,1000000,10000000 \
		'./duc checksum -b{bufsize} $<'

build-benchmark:
	docker build -t duc:benchmark -f benchmarking/Dockerfile .
