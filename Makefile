.PHONY: help build test clean fmt lint install

# Go parameters
GOCMD=go
GOPATH=$(shell go env GOPATH)
BINARY_NAME=fritzboxctl
MAIN_PATH=./cmd/fritzboxctl

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GOCMD) build -o $(BINARY_NAME) $(MAIN_PATH)

test: ## Run all tests
	$(GOCMD) test ./... -v

test-short: ## Run tests (short mode)
	$(GOCMD) test ./... -short

test-coverage: ## Run tests with coverage
	$(GOCMD) test ./... -coverprofile=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

fmt: ## Format the code
	$(GOCMD) fmt ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

tidy: ## Tidy go modules
	$(GOCMD) mod tidy

install: ## Install the binary to GOPATH/bin
	$(GOCMD) install $(MAIN_PATH)

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

run: build ## Build and run (example: make run ARGS="device info")
	./$(BINARY_NAME) $(ARGS)

deps: ## Download dependencies
	$(GOCMD) mod download

all: clean fmt test build ## Run all (clean, fmt, test, build)
