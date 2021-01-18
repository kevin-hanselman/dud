docker_image = dud-dev
base_dir = $(shell pwd)
GOPATH ?= ~/go
GOBIN ?= $(GOPATH)/bin

dud: test
	go build -o dud \
		-ldflags "-s -w -X 'github.com/kevin-hanselman/dud/src/cmd.Version=$(shell git rev-parse --short HEAD)'"

.PHONY: install
install: $(GOBIN)/dud

.PHONY: cli-docs
cli-docs: dud
	./dud gen-docs hugo/content/cli

.PHONY: docs
# TODO: add hugo/content/%.md as a prerequisite
docs: cli-docs
	cd hugo && hugo --minify

$(GOBIN)/dud: dud
	cp dud $(GOBIN)

.PHONY: docker%
# Create an interactive session in the development Docker image.
docker: docker-image
	docker run \
		--rm \
		-it \
		-p 8888:8888 \
		-v $(base_dir):/dud \
		$(docker_image)

# Run any rule in this Makefile in the development Docker image.
docker-%: docker-image
	docker run \
		--rm \
		-p 8888:8888 \
		-v $(base_dir):/dud \
		$(docker_image) make $(patsubst docker-%,%,$@)

docker-image:
	docker build \
		-t $(docker_image) \
		-f ./integration/Dockerfile \
		.

.PHONY: fmt
fmt: $(GOBIN)/goimports
	goimports -w .
	gofmt -s -w .

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint ./...

.PHONY: test-%
test-short: fmt lint
	go test -short ./...

test: fmt lint
	go test -cover -race ./...

.PHONY: bench
bench: test-short
	go test ./... -benchmem -bench .

.PHONY: serve-jupyter
serve-jupyter: $(GOBIN)/dud
	jupyter notebook -y --ip=0.0.0.0 ./hugo/notebooks/

hugo/notebooks/%.md:
	jupyter nbconvert \
		--to markdown \
		--TagRemovePreprocessor.remove_input_tags 'hide_input' \
		--TagRemovePreprocessor.remove_all_outputs_tags 'hide_output' \
		'$(patsubst %.md,%.ipynb,$@)'

hugo/content/%.md: hugo/notebooks/%.md
	mkdir -p '$(dir $@)'
	awk --lint=fatal -f ./hugo/notebooks/fix_md.awk '$<' > '$@'
	$(eval supporting_files = $(wildcard $(patsubst %.md,%_files,$<)/*.*))
	if test -n "$(supporting_files)"; then cp -v $(supporting_files) $(dir $@); fi

# xargs trims whitespace from the hostname below
.PHONY: serve-hugo
serve-hugo:
	cd hugo && \
	hugo server \
		--disableFastRender \
		--bind 0.0.0.0 \
		--baseUrl $(shell hostname -i | xargs)/dud/

.PHONY: %-test-cov
%-test-cov: %-test.coverage
	go tool cover -html=$<

unit-test.coverage:
	go test -short ./... -coverprofile=$@

int-test.coverage:
	go test -run Integration ./... -coverprofile=$@

all-test.coverage:
	go test ./... -coverprofile=$@

.PHONY: integration-%
integration-test: $(GOBIN)/dud
	python $(base_dir)/integration/run_tests.py

integration-bench: $(GOBIN)/dud
	mkdir -p ~/dud_integration_benchmarks
	cd ~/dud_integration_benchmarks && python $(base_dir)/integration/run_benchmarks.py

.PHONY: deep-lint
deep-lint:
	docker run \
		--rm \
		-v $(base_dir):/app \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run

.PHONY: clean
clean:
	rm -rf hugo/content/docs/cli/dud*.md ./docs/*
	rm -f *.coverage *.bin depgraph.png mockery $(GOBIN)/dud
	go clean ./...

.PHONY: clean-docker
clean-docker:
	docker rmi -f $(docker_image)

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: loc
loc:
	tokei --sort lines --exclude 'src/mocks/' ./src/ ./integration/
	tokei --sort lines --exclude 'src/mocks/' --exclude '*_test.go' ./src/

mockery:
	curl -L https://github.com/vektra/mockery/releases/download/v2.2.1/mockery_2.2.1_Linux_x86_64.tar.gz \
		| tar -zxvf - mockery

.PHONY: mocks
mocks: mockery
	./mockery --all --output src/mocks

# The awk command removes all graph edge definitions that don't include dud
depgraph.png:
	godepgraph -nostdlib . \
		| awk '/^[^"]/ || /dud/ {print}' \
		| dot -Tpng -o $@

%mb_random.bin:
	dd if=/dev/urandom of=$@ bs=1M count=$(patsubst %mb_random.bin,%,$@)

.PHONY: hyperfine
hyperfine: 50mb_random.bin dud
	hyperfine -L cmd sha1sum,md5sum,sha256sum,b2sum,xxh64sum,'./dud checksum' \
		'{cmd} $<'
	hyperfine -L bufsize 4096,8192,16384,32768,65536,131072,262144,524288,1048576 \
		'./dud checksum -b{bufsize} $<'

$(GOBIN)/goimports:
	go install golang.org/x/tools/cmd/goimports

$(GOBIN)/golint:
	go install golang.org/x/lint/golint
