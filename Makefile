PKGS = ./...
TESTFLAGS = -vet all -mod readonly
BUILDFLAGS = -v
BENCH = .
BENCHFLAGS = -benchmem -bench=${BENCH}


# For docker/push, set ENV=dev (default) or ENV=prod, e.g. make push ENV=prod
ENV ?= dev
ifeq ($(filter dev prod,$(ENV)),)
$(error ENV must be dev or prod (got: $(ENV)))
endif
ifeq ($(ENV),prod)
PROJECT := grafanalabs-global
else
PROJECT := grafanalabs-dev
endif
IMAGE_PREFIX ?= us-docker.pkg.dev/$(PROJECT)/docker-unused-$(ENV)
IMAGE_NAME ?= unused
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

lint: ## Runs linting checks
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0 run -c .golangci.yml

docker: ## Builds docker image and applies a tag
	docker buildx build --build-arg=GIT_VERSION=$(GIT_VERSION) --platform linux/amd64 -t $(IMAGE_PREFIX)/$(IMAGE_NAME) -f Dockerfile.exporter . --load
	docker tag $(IMAGE_PREFIX)/$(IMAGE_NAME) $(IMAGE_PREFIX)/$(IMAGE_NAME):$(GIT_VERSION)

push: docker ## Pushes docker image (runs build first; use ENV=dev or ENV=prod)
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):$(GIT_VERSION)
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):latest
