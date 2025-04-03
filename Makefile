# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Project parameters
BINARY_NAME=vibe
MAIN_PACKAGE=./examples/todo
PACKAGES=$(shell $(GOCMD) list ./... | grep -v /examples/)

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test coverage lint fmt vet tidy help examples docs

all: test build

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Build examples
examples:
	$(GOBUILD) -o examples/todo/todo examples/todo

# Clean build files
clean:
	$(GOCMD) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -f examples/todo/todo

# Run tests
test:
	$(GOTEST) -v $(PACKAGES)

# Run tests with race detection
test-race:
	$(GOTEST) -race -v $(PACKAGES)

# Generate test coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out $(PACKAGES)
	$(GOCMD) tool cover -html=coverage.out

# Run short tests only
test-short:
	$(GOTEST) -short -v $(PACKAGES)

# Run linter
lint:
	$(GOLINT) run ./...

lintfix:
	$(GOLINT) run ./... --fix

# Format code
fmt:
	$(GOFMT) ./...

# Run go vet
vet:
	$(GOVET) $(PACKAGES)

# Update dependencies
tidy:
	$(GOMOD) tidy

# Generate documentation
docs:
	$(GOCMD) doc -all .

# Install dependencies
deps:
	$(GOGET) -u golang.org/x/lint/golint
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem $(PACKAGES)

# Help command
help:
	@echo "Available commands:"
	@echo "  make build       - Build the project"
	@echo "  make examples    - Build example applications"
	@echo "  make test        - Run tests"
	@echo "  make test-race   - Run tests with race detection"
	@echo "  make coverage    - Generate test coverage report"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"
	@echo "  make vet         - Run go vet"
	@echo "  make tidy        - Update dependencies"
	@echo "  make clean       - Clean build files"
	@echo "  make docs        - Generate documentation"
	@echo "  make deps        - Install development dependencies"
	@echo "  make bench       - Run benchmarks"