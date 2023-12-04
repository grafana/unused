PKGS = ./...
TESTFLAGS = -race -vet all -mod readonly
BUILDFLAGS = -v
BENCH = .
BENCHFLAGS = -benchmem -bench=${BENCH}

build:
	go build ${BUILDFLAGS} ${PKGS}

test:
	go test ${TESTFLAGS} ${PKGS}

benchmark:
	go test ${TESTFLAGS} ${BENCHFLAGS} ${PKGS}

checks:
	go vet ${PKGS}
	staticcheck ${PKGS}
	make lint

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -c .golangci.yml
