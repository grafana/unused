PKGS = ./...
TESTFLAGS = -race -vet all -mod readonly
BUILDFLAGS = -v
BENCH = .
BENCHFLAGS = -benchmem -bench=${BENCH}

IMAGE_PREFIX ?= us.gcr.io/kubernetes-dev
GIT_VERSION ?= $(shell git describe --tags --always --dirty)

build:
	go build ${BUILDFLAGS} ${PKGS}

test:
	go test ${TESTFLAGS} ${PKGS}

benchmark:
	go test ${TESTFLAGS} ${BENCHFLAGS} ${PKGS}

checks:
	go vet ${PKGS}
	go run honnef.co/go/tools/cmd/staticcheck@latest ${PKGS}
	make lint

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -c .golangci.yml

docker:
	docker build --build-arg=GIT_VERSION=$(GIT_VERSION) -t $(IMAGE_PREFIX)/unused -f Dockerfile.exporter .
	docker tag $(IMAGE_PREFIX)/unused $(IMAGE_PREFIX)/unused:$(GIT_VERSION)

push: docker
	docker push $(IMAGE_PREFIX)/unused:$(GIT_VERSION)
	docker push $(IMAGE_PREFIX)/unused:latest
