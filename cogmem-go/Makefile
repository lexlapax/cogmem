# Makefile for CogMem Go Library

.PHONY: all build test fmt vet tidy clean docker-up docker-down integration-test cover

# Default target: format, vet, then test
all: fmt vet test

# Compile packages (ensures code compiles)
build:
	go build ./...

# Run all unit and integration tests
test:
	go test ./...

# Run integration tests only
integration-test:
	@$(MAKE) docker-up
	@go test ./test/integration/...
	@$(MAKE) docker-down

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Tidy go.mod and go.sum
tidy:
	go mod tidy

# Run tests with coverage
cover:
	go test -coverprofile=coverage.out ./...

# Clean build artifacts
clean:
	@rm -f coverage.out

# Spin up local dependencies via Docker Compose
docker-up:
	docker-compose up -d

# Tear down Docker Compose services
docker-down:
	docker-compose down