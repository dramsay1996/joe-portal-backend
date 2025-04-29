.PHONY: build run generate-hash install lint format

# Install dependencies
install:
	go mod tidy
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Build the Go application
build: install
	go build -o bin/server cmd/server/main.go

# Run the Go application
run: install
	go run cmd/server/main.go

# Generate a password hash
generate-hash:
	cd scripts/generate-hash && go run main.go -password $(password)

# Run linter
lint:
	golangci-lint run

# Format code
format:
	go fmt ./...
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/ 