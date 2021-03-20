docker_image = dud-dev
base_dir = $(shell pwd)
GOPATH ?= ~/go
GOBIN ?= $(GOPATH)/bin

dud: test
	go build -o dud \
		-ldflags "-s -w -X 'main.version=$(shell git rev-parse --short HEAD)'"

.PHONY: install
install: $(GOBIN)/dud

$(GOBIN)/dud: dud
	cp -v dud $(GOBIN)

.PHONY: cli-docs
cli-docs: dud
	rm -f hugo/content/docs/cli/dud*.md
	./dud gen-docs hugo/content/cli

.PHONY: website
website: cli-docs
	rm -rf ./website
	cd hugo && hugo --minify

.PHONY: docker%
# Create an interactive session in the development Docker image.
docker: docker-image
	docker run \
		--rm \
		-it \
		-u $(shell id -u):$(shell id -g) \
		-p 8888:8888 \
		-v $(base_dir):/dud \
		-v dud-data:/home/user/dud-data \
		$(docker_image)

# Run any rule in this Makefile in the development Docker image.
docker-%: docker-image
	docker run \
		--rm \
		-p 8888:8888 \
		-v $(base_dir):/dud \
		-v dud-data:/dud-data \
		$(docker_image) make $(patsubst docker-%,%,$@)

docker-image:
	docker volume create dud-data
	docker build \
		-t $(docker_image) \
		-f ./integration/Dockerfile \
		.

.PHONY: fmt
fmt: $(GOBIN)/gofumpt $(GOBIN)/goimports
	goimports -w -l .
	gofumpt -w -l .

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint ./...

.PHONY: test%
test-short: src/mocks fmt lint
	go test -short ./...

test: src/mocks fmt lint
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
	jupyter nbconvert \
		--ClearOutputPreprocessor.enabled=True \
		--inplace \
		'$(patsubst %.md,%.ipynb,$@)'

# TODO: Make won't recognize this rule (and maybe others like it) when run with
# the docker- prefix.
hugo/content/benchmarks/_index.md: \
	integration/benchmarks/markdown/00_front_matter.md \
	integration/benchmarks/markdown/few_large_files/commit/table.md \
	integration/benchmarks/markdown/many_small_files/commit/table.md \
	integration/benchmarks/markdown/few_large_files/checkout/table.md \
	integration/benchmarks/markdown/many_small_files/checkout/table.md
	mkdir -p $(dir $@)
	find integration/benchmarks/markdown -type f -name '*.md' | sort | xargs cat > $@

.PHONY: bench-docs
bench-docs: hugo/content/benchmarks/_index.md

hugo/content/%.md: hugo/notebooks/%.md
	mkdir -p '$(dir $@)'
	awk --lint=fatal -f ./hugo/notebooks/fix_md.awk '$<' > '$@'
	$(eval supporting_files = $(wildcard $(patsubst %.md,%_files,$<)/*.*))
	if test -n "$(supporting_files)"; then cp -v $(supporting_files) $(dir $@); fi

# TODO: This currently needs to be run manually, as the rule that uses it
# (.../%/table.md below) is specific to a dataset and workflow pair (not just
# a dataset).
~/dud-data/%:
	mkdir $@
	./integration/benchmarks/datasets/$*.sh $@

integration/benchmarks/markdown/00_front_matter.md: $(GOBIN)/dud
	./integration/benchmarks/generate_front_matter.sh > $@

# The pipe ("|") makes the Dud executable an "order-only" prerequisite. The
# installed Dud executable is not found in the Docker image on boot, so it will
# always be built and installed. This installation will always result in the
# Dud executable being newer than the target (i.e. table.md), and thus Make
# will always run this rule. Order-only prerequisites ignore timestamps.
# See also: https://stackoverflow.com/a/58040049/857893
integration/benchmarks/markdown/%/table.md: | $(GOBIN)/dud
	$(eval parent_dirs = $(subst /, ,$*))
	mkdir ~/dud-bench
	./integration/benchmarks/hyperfine.sh \
		~/dud-bench \
		~/dud-data/$(word 1, $(parent_dirs)) \
		integration/benchmarks/workflows/$(word 2, $(parent_dirs)) \
		integration/benchmarks/markdown
	rm -rf ~/dud-bench

.PHONY: serve-hugo
serve-hugo:
	cd hugo && \
	hugo server \
		--disableFastRender \
		--bind 0.0.0.0 \
		--baseUrl $(shell hostname -i | xargs)/dud/
# xargs trims whitespace from the hostname

.PHONY: %-test-cov
%-test-cov: %-test.coverage
	go tool cover -html=$<

unit-test.coverage:
	go test -short ./... -coverprofile=$@

int-test.coverage:
	go test -run Integration ./... -coverprofile=$@

all-test.coverage:
	go test ./... -coverprofile=$@

.PHONY: integration-test
integration-test: $(GOBIN)/dud
	python $(base_dir)/integration/run_tests.py

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
	rm -rf hugo/content/docs/cli/dud*.md ./website
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
	tokei --sort lines --exclude src/mocks/ ./src/ ./integration/
	tokei --sort lines --exclude src/mocks/ --exclude '*_test.go' ./src/

mockery:
	curl -L https://github.com/vektra/mockery/releases/download/v2.2.1/mockery_2.2.1_Linux_x86_64.tar.gz \
		| tar -zxvf - mockery

goreleaser:
	curl -L https://github.com/goreleaser/goreleaser/releases/download/v0.159.0/goreleaser_Linux_x86_64.tar.gz \
		| tar -zxvf - goreleaser

src/mocks: mockery
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

$(GOBIN)/gofumpt:
	go install mvdan.cc/gofumpt@latest

$(GOBIN)/golint:
	go install golang.org/x/lint/golint@latest

$(GOBIN)/goimports:
	go install golang.org/x/tools/cmd/goimports@latest
