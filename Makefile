VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)
BINARY := incloud

.PHONY: build install clean lint test

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/incloud

install:
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/incloud

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
