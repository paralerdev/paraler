.PHONY: build install clean run test lint

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

BINARY := paraler
BUILD_DIR := bin

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/paraler

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/paraler

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run the application
run:
	go run ./cmd/paraler

# Run with a specific config
run-example:
	go run ./cmd/paraler -config paraler.example.yaml

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all: build-darwin build-linux

build-darwin:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/paraler
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/paraler

build-linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/paraler
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/paraler

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run the application"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  deps        - Download dependencies"
	@echo "  build-all   - Build for all platforms"
