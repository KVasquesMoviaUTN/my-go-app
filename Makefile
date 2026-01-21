# Variables
BINARY_NAME=bot
DOCKER_IMAGE=arbitrage-bot
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")

# Default target
all: build

# Build the binary
build:
	@echo "Building..."
	go build -o bin/$(BINARY_NAME) cmd/bot/main.go

# Run the bot
run: build
	@echo "Running..."
	./bin/$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

# Run linter
lint:
	@echo "Linting..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Docker run
docker-run:
	docker-compose up --build

.PHONY: all build run test test-integration lint clean docker-build docker-run
