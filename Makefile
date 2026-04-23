MODULE  := github.com/DavDaz/llm-wiki-generator
BINARY  := llm-wiki
VERSION ?= dev
LDFLAGS := -ldflags "-X '$(MODULE)/internal/version.Version=$(VERSION)'"

.PHONY: build test lint vet tidy clean release

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

release:
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "Usage: make release VERSION=v0.3.0"; \
		exit 1; \
	fi
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "Error: VERSION must be semver (e.g. v0.3.0)"; \
		exit 1; \
	fi
	@echo "Releasing $(VERSION)..."
	git tag $(VERSION)
	git push origin $(VERSION)
	@echo "Done — GitHub Actions will build and publish the release."
