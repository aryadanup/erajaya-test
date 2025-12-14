
APP_NAME = erajaya-test
BUILD_DIR = bin
MOCKS_DIR = internal/mocks

.PHONY: all build run test clean docker-build docker-up docker-down migrate-up migrate-down mocks lint swagger

all: build

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) main.go

run:
	@echo "Running $(APP_NAME)..."
	@go mod tidy
	@go run main.go

test:
	@echo "Running Tests..."
	@go test -v -coverprofile=coverage.out ./...
	@grep -v -E "/mocks/|_mock.go" coverage.out > coverage.final.out
	@go tool cover -func=coverage.final.out
	@rm coverage.final.out

test-race:
	@echo "Running tests with race detector..."
	@go test -v -race ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf swagger/swagger.*

mocks:
	@echo "Generating mocks with mockery v3..."
	@mockery

mockery-setup:
	@echo "Installing mockery..."
	@go install github.com/vektra/mockery/v3@v3.5.2

swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g main.go -o swagger

docker-build:
	@echo "Building Docker image..."
	@docker compose build

docker-up:
	@echo "Starting Docker containers..."
	@docker compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker compose down

lint:
	@echo "Running linter..."
	@go vet ./...

migrate-setup:
	@echo "Installing golang-migrate..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-up:
	@echo "Running migrations up..."
	@migrate -path migrations -database "postgres://user:password@localhost:5432/erajaya-test_db?sslmode=disable" up

migrate-down:
	@echo "Running migrations down..."
	@migrate -path migrations -database "postgres://user:password@localhost:5432/erajaya-test_db?sslmode=disable" down
