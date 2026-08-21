GO ?= go
BINARY ?= uberlauncher
CMD ?= ./cmd/uberlauncher

.PHONY: help
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
	@echo "  make install    Install binary via go install"
	@echo "  make clean      Remove built binary"

.PHONY: build
build:
	$(GO) build -o $(BINARY) $(CMD)

.PHONY: install
install:
	$(MAKE) build
	$(GO) install $(CMD)

.PHONY: run
run:
	$(GO) run $(CMD) --verbose --config ./config/default.toml

.PHONY: test
test:
	$(GO) test ./...

.PHONY: fmt
fmt:
	find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -w

.PHONY: fmt-check
fmt-check:
	@test -z "$(shell find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -l)" || \
	(echo "Run 'make fmt' to format files" && exit 1)

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || \
	(echo "golangci-lint not found."; exit 127)
	golangci-lint run

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: check
check: fmt-check vet test lint

.PHONY: clean
clean:
	rm -f $(BINARY)
