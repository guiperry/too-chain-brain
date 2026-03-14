## ============================================================
##  Tool-Chain-Brain — Makefile
## ============================================================

BINARY      := tcb
CMD_PKG     := ./cmd/tcb
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS     := -s -w \
               -X 'github.com/tool-chain-brain/tcb/internal/scanner.TCBVersion=$(VERSION)'

GOBIN       ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN       := $(shell go env GOPATH)/bin
endif

GOOS        ?= $(shell go env GOOS)
GOARCH      ?= $(shell go env GOARCH)

.DEFAULT_GOAL := build

# ── Build ────────────────────────────────────────────────────
.PHONY: build
build:  ## Build the binary for the current platform
	@echo "→ Building $(BINARY) ($(GOOS)/$(GOARCH)) $(VERSION)"
	@go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(CMD_PKG)
	@echo "✓ $(BINARY) ready"

.PHONY: build-all
build-all:  ## Cross-compile for Linux, macOS (amd64 + arm64) and Windows
	@mkdir -p dist
	@for PLATFORM in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		OS=$$(echo $$PLATFORM | cut -d/ -f1); \
		ARCH=$$(echo $$PLATFORM | cut -d/ -f2); \
		OUT=dist/$(BINARY)_$${OS}_$${ARCH}; \
		[ "$$OS" = "windows" ] && OUT=$$OUT.exe; \
		echo "→ Building $$OUT"; \
		GOOS=$$OS GOARCH=$$ARCH go build -ldflags="$(LDFLAGS)" -o $$OUT $(CMD_PKG); \
	done
	@echo "✓ All binaries in dist/"

# ── Install / Uninstall ──────────────────────────────────────
.PHONY: install
install: build  ## Install tcb to GOBIN (default: $(go env GOPATH)/bin)
	@echo "→ Installing $(BINARY) to $(GOBIN)"
	@cp $(BINARY) $(GOBIN)/$(BINARY)
	@echo "✓ Installed: $(GOBIN)/$(BINARY)"
	@echo "  Run: tcb --help"

.PHONY: uninstall
uninstall:  ## Remove tcb from GOBIN
	@rm -f $(GOBIN)/$(BINARY)
	@echo "✓ Uninstalled $(BINARY)"

# ── go install ───────────────────────────────────────────────
.PHONY: go-install
go-install:  ## Install via `go install` (requires Go 1.21+)
	@go install -ldflags="$(LDFLAGS)" $(CMD_PKG)
	@echo "✓ Installed via go install"

# ── Dev ──────────────────────────────────────────────────────
.PHONY: run
run: build  ## Build and run tcb scan
	./$(BINARY) scan

.PHONY: tidy
tidy:  ## Tidy go modules
	go mod tidy

.PHONY: deps
deps:  ## Download all dependencies
	go mod download

.PHONY: fmt
fmt:  ## Format all Go source files
	gofmt -w .

.PHONY: vet
vet:  ## Run go vet
	go vet ./...

.PHONY: lint
lint:  ## Run golangci-lint (must be installed)
	golangci-lint run ./...

# ── Test ─────────────────────────────────────────────────────
.PHONY: test
test:  ## Run all tests
	go test -v ./...

.PHONY: test-cover
test-cover:  ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# ── Release ──────────────────────────────────────────────────
.PHONY: release-dry
release-dry:  ## Dry-run goreleaser (no publish)
	goreleaser release --snapshot --clean

.PHONY: release
release:  ## Full goreleaser release (requires GITHUB_TOKEN)
	goreleaser release --clean

# ── Clean ────────────────────────────────────────────────────
.PHONY: clean
clean:  ## Remove build artefacts
	@rm -f $(BINARY)
	@rm -rf dist/ coverage.out coverage.html
	@echo "✓ Clean"

# ── Help ─────────────────────────────────────────────────────
.PHONY: help
help:  ## Show this help message
	@echo ""
	@echo "  Tool-Chain-Brain (tcb) — Makefile targets"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
