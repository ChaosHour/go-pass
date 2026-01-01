.PHONY: build test clean run lint fmt

# Build the application
build:
	mkdir -p bin
	go build -o bin/go-pass ./cmd/pass

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run the application (example usage)
run:
	./bin/go-pass -h

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Cross-compile for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/go-pass-linux-amd64 ./cmd/pass
	GOOS=darwin GOARCH=amd64 go build -o bin/go-pass-darwin-amd64 ./cmd/pass
	GOOS=windows GOARCH=amd64 go build -o bin/go-pass-windows-amd64.exe ./cmd/pass