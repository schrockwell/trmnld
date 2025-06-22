# Variables
APP_NAME = trmnl-api-server
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT = $(shell git rev-parse --short HEAD)

# Build flags
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT) -w -s"

# Directories
BUILD_DIR = build
DIST_DIR = dist

# Default target
.PHONY: all
all: clean build

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)

# Create build directories
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Build for current platform
.PHONY: build
build: $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

# Cross-compile for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

# Linux builds
.PHONY: build-linux
build-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 .

# macOS builds
.PHONY: build-darwin
build-darwin: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

# Windows builds
.PHONY: build-windows
build-windows: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe .

# Create distribution packages
.PHONY: dist
dist: build-all $(DIST_DIR)
	# Linux AMD64
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-amd64
	
	# Linux ARM64
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-arm64
	
	# macOS AMD64
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-amd64
	
	# macOS ARM64
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-arm64
	
	# Windows AMD64
	zip -j $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe
	
	# Windows ARM64
	zip -j $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-arm64.zip $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe

# Generate checksums
.PHONY: checksums
checksums: dist
	cd $(DIST_DIR) && sha256sum *.tar.gz *.zip > checksums.txt

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Development build (with race detector)
.PHONY: dev
dev: $(BUILD_DIR)
	go build -race $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-dev .

# Run the application
.PHONY: run
run:
	go run .

# Install the application
.PHONY: install
install:
	go install $(LDFLAGS) .

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build for current platform"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Cross-compile for all platforms"
	@echo "  build-linux  - Build for Linux (amd64, arm64)"
	@echo "  build-darwin - Build for macOS (amd64, arm64)"
	@echo "  build-windows- Build for Windows (amd64, arm64)"
	@echo "  dist         - Create distribution packages"
	@echo "  checksums    - Generate checksums for distribution packages"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  deps         - Install dependencies"
	@echo "  dev          - Build development version with race detector"
	@echo "  run          - Run the application"
	@echo "  install      - Install the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help message"