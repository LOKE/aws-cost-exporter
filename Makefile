.PHONY: build run test clean tidy help

# Build the application
build:
	go build -o aws-cost-exporter .

# Run the application
run:
	go run main.go

# Run tests
test:
	go test ./...

# Clean up dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -f aws-cost-exporter

# Show help
help:
	@echo "Available targets:"
	@echo "  build  - Build the application"
	@echo "  run    - Run the application"
	@echo "  test   - Run all tests"
	@echo "  tidy   - Clean up dependencies"
	@echo "  clean  - Clean build artifacts"
	@echo "  help   - Show this help message"