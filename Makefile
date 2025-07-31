.PHONY: build test clean install run lint coverage

# Binary name
BINARY_NAME=standup-bot
BINARY_PATH=./cmd/standup-bot

# Build the binary
build:
	go build -o $(BINARY_NAME) $(BINARY_PATH)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install the binary to $GOPATH/bin
install:
	go install $(BINARY_PATH)

# Run the application
run:
	go run $(BINARY_PATH)

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
	fi

# Format code
fmt:
	go fmt ./...

# Download dependencies
deps:
	go mod download

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy