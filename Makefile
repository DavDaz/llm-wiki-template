MODULE  := github.com/DavDaz/llm-wiki-template
BINARY  := llm-wiki
VERSION ?= dev
LDFLAGS := -ldflags "-X '$(MODULE)/internal/version.Version=$(VERSION)'"

.PHONY: build test lint vet tidy clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test -race ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/
