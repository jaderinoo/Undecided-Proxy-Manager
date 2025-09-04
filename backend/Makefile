.PHONY: swagger build run clean

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated successfully!"

# Build the application
build:
	@echo "Building the application..."
	@go build -o bin/upm-backend main.go
	@echo "Build completed!"

# Run the application
run: swagger
	@echo "Starting the application..."
	@go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/
	@echo "Clean completed!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies installed!"

# Install Swagger CLI
install-swagger:
	@echo "Installing Swagger CLI..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Swagger CLI installed!"

# Full setup
setup: install-swagger deps swagger
	@echo "Setup completed!"
