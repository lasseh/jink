.PHONY: all build build-linux rebuild install clean test vet fmt lint deps release release-snapshot demo demo-set demo-all help

# Project info
BINARY     := jink
DEMO       := jink-demo
BUILD_DIR  := build
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"

# Source files for dependency tracking
SRC := $(shell find . -name '*.go' -not -path './vendor/*')

# Default target
all: help

# Build binaries - rebuilds when ANY .go file changes
build: $(BUILD_DIR)/$(BINARY) $(BUILD_DIR)/$(DEMO)

# Cross-compile for linux/amd64
build-linux: $(SRC)
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/jink

# Force rebuild
rebuild: clean build

$(BUILD_DIR)/$(BINARY): $(SRC)
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $@ ./cmd/jink

$(BUILD_DIR)/$(DEMO): $(SRC)
	@mkdir -p $(BUILD_DIR)
	go build -o $@ ./cmd/jink-demo

# Install to GOPATH/bin or GOBIN
install: build
	@GOBIN=$${GOBIN:-$$(go env GOPATH)/bin}; \
	cp $(BUILD_DIR)/$(BINARY) "$$GOBIN/"; \
	echo "Installed $(BINARY) to $$GOBIN"

# Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean -cache -testcache

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Static analysis
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Lint (requires: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@which golangci-lint > /dev/null || (echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# Download and tidy dependencies
deps:
	go mod download
	go mod tidy

# GoReleaser: create a release (requires a git tag)
release:
	@which goreleaser > /dev/null || (echo "Install: go install github.com/goreleaser/goreleaser/v2@latest" && exit 1)
	goreleaser release --clean

# GoReleaser: test build locally without publishing
release-snapshot:
	@which goreleaser > /dev/null || (echo "Install: go install github.com/goreleaser/goreleaser/v2@latest" && exit 1)
	goreleaser release --snapshot --clean

# Demo targets
demo: $(BUILD_DIR)/$(DEMO)
	@./$(BUILD_DIR)/$(DEMO)

demo-set: $(BUILD_DIR)/$(DEMO)
	@./$(BUILD_DIR)/$(DEMO) -set

demo-all: $(BUILD_DIR)/$(DEMO)
	@./$(BUILD_DIR)/$(DEMO) -all

# Show help
help:
	@echo "jink - ink your JunOS config"
	@echo ""
	@echo "Build:"
	@echo "  make build        Build binaries to $(BUILD_DIR)/"
	@echo "  make build-linux  Cross-compile for Linux amd64"
	@echo "  make rebuild      Force rebuild (clean + build)"
	@echo "  make install   Install $(BINARY) to GOPATH/bin"
	@echo "  make clean     Remove build artifacts"
	@echo ""
	@echo "Test:"
	@echo "  make test      Run all tests"
	@echo "  make coverage  Run tests with coverage report"
	@echo "  make vet       Run go vet"
	@echo "  make fmt       Format code"
	@echo "  make lint      Run golangci-lint"
	@echo ""
	@echo "Demo:"
	@echo "  make demo      Show highlighting demo"
	@echo "  make demo-set  Show set-style config demo"
	@echo "  make demo-all  Show all themes side by side"
	@echo ""
	@echo "Release:"
	@echo "  make release           Run goreleaser (requires git tag)"
	@echo "  make release-snapshot  Test release build locally"
	@echo ""
	@echo "Other:"
	@echo "  make deps      Download and tidy dependencies"
