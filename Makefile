PKGS = ./...
TESTFLAGS = -vet all -mod readonly
BUILDFLAGS = -v
BENCH = .
BENCHFLAGS = -benchmem -bench=${BENCH}

IMAGE_PREFIX ?= us.gcr.io/kubernetes-dev
GIT_VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: help

help: ## Prints this help message (Default)
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Builds the project
	go build ${BUILDFLAGS} ${PKGS}

test: ## Runs tests the project (non-benchmark)
	go test ${TESTFLAGS} ${PKGS}

benchmark: ## Runs benchmark tests
	go test ${TESTFLAGS} ${BENCHFLAGS} ${PKGS}

checks: ## Runs vetting, static checks, and linting checks
	go vet ${PKGS}
	go run honnef.co/go/tools/cmd/staticcheck@latest ${PKGS}
	make lint

lint: ## Runs linting checks
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -c .golangci.yml

docker: ## Builds docker image and applies a tag
	docker build --build-arg=GIT_VERSION=$(GIT_VERSION) -t $(IMAGE_PREFIX)/unused -f Dockerfile.exporter . --load
	docker tag $(IMAGE_PREFIX)/unused $(IMAGE_PREFIX)/unused:$(GIT_VERSION)

push: docker ## Pushes docker image (runs build first)
	docker push $(IMAGE_PREFIX)/unused:$(GIT_VERSION)
	docker push $(IMAGE_PREFIX)/unused:latest
