BINARY_NAME=telegram-security-bot
GO_FILES=$(shell find . -name "*.go" -type f)

.PHONY: build build-linux build-windows clean run test deps

# Default target
build: build-linux build-windows

# Build for Linux
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-linux-amd64 .

# Build for Windows
build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-windows-amd64.exe .

# Build for current platform
build-local:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) .

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run the bot
run:
	go run .

# Run with custom config
run-config:
	go run . -config=$(CONFIG)

# Test
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Create directories
dirs:
	mkdir -p bin

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build for Linux and Windows"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-windows- Build for Windows"
	@echo "  build-local  - Build for current platform"
	@echo "  deps         - Install dependencies"
	@echo "  run          - Run the bot"
	@echo "  run-config   - Run with custom config (CONFIG=/path/to/config.json)"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help"