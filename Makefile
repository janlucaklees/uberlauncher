GO ?= go
BINARY ?= uberlauncher
CMD ?= ./cmd/uberlauncher

.PHONY: help build run test fmt fmt-check vet lint tidy check clean

help:
	@echo "Available targets:"
	@echo "  make build      Build the launcher binary"
	@echo "  make run        Run launcher via go run"
	@echo "  make test       Run all Go tests"
	@echo "  make fmt        Format all Go files"
	@echo "  make fmt-check  Verify Go formatting"
	@echo "  make vet        Run go vet"
	@echo "  make lint       Run golangci-lint"
	@echo "  make tidy       Run go mod tidy"
	@echo "  make check      Run format check + vet + test + lint"
	@echo "  make clean      Remove built binary"

build:
	$(GO) build -o $(BINARY) $(CMD)

run:
	$(GO) run $(CMD) --verbose

test:
	$(GO) test ./...

fmt:
	find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -w

fmt-check:
	@test -z "$(shell find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -l)" || \
	(echo "Run 'make fmt' to format files" && exit 1)

vet:
	$(GO) vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || \
	(echo "golangci-lint not found."; exit 127)
	golangci-lintrun

tidy:
	$(GO) mod tidy

check: fmt-check vet test lint

clean:
	rm -f $(BINARY)
