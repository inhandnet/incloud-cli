VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MODULE  := github.com/inhandnet/incloud-cli/internal/build

LDFLAGS := -X $(MODULE).Version=$(VERSION) \
           -X $(MODULE).Commit=$(COMMIT) \
           -X $(MODULE).Date=$(DATE)

BINARY := incloud

PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: build build-all install clean fmt lint test

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/incloud

build-all:
	@for pair in $(PLATFORMS); do \
		OS=$${pair%/*}; ARCH=$${pair#*/}; EXT=""; \
		[ "$$OS" = "windows" ] && EXT=".exe"; \
		echo "Building $${OS}/$${ARCH}..."; \
		CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build \
			-ldflags "$(LDFLAGS)" \
			-o bin/$(BINARY)-$${OS}-$${ARCH}$${EXT} ./cmd/incloud; \
	done

install:
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/incloud

test:
	CGO_ENABLED=0 go test ./... -v

fmt:
	golangci-lint fmt ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
