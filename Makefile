BINARY_NAME=wallet-service
DOCKER_COMPOSE=docker-compose

.PHONY: help start stop restart clean test build swagger

help:
	@echo "💰 Wallet Service - Available Commands:"
	@echo ""
	@echo "  Development:"
	@echo "    make dev         - Start only database"
	@echo "    make run         - Run Go application locally"
	@echo "    make build       - Build binary"
	@echo "    make swagger     - Generate Swagger documentation"
	@echo ""
	@echo "  Docker:"
	@echo "    make start       - Start all services with Docker Compose"
	@echo "    make stop        - Stop all services"
	@echo "    make restart     - Restart all services"
	@echo "    make docker-run  - Run only app in Docker"
	@echo "    make logs        - View logs"
	@echo "    make status      - Check service status"
	@echo ""
	@echo "  Testing:"
	@echo "    make test        - Run all main tests"
	@echo "    make test-unit   - Run only unit tests"
	@echo "    make test-integration - Run integration tests"
	@echo "    make test-load   - Run load tests"
	@echo "    make test-e2e    - Run E2E tests"
	@echo "    make test-all    - Run complete test suite"
	@echo ""
	@echo "  Database:"
	@echo "    make db-shell    - Connect to database"
	@echo ""
	@echo "  Maintenance:"
	@echo "    make clean       - Clean everything"
	@echo "    make deps        - Install dependencies"

# Development
dev:
	@echo "🐘 Starting database only..."
	@$(DOCKER_COMPOSE) up -d postgres
	@echo "✅ Database running on localhost:5433"
	@echo "💡 Now you can run the app: make run"

run:
	@echo "🚀 Running Go application locally..."
	@DOCKER_CONTAINER=false go run cmd/$(BINARY_NAME)/main.go

build:
	@echo "🔨 Building application..."
	@go build -o $(BINARY_NAME) cmd/$(BINARY_NAME)/main.go
	@echo "✅ Built: $(BINARY_NAME)"

# Swagger documentation
swagger:
	@echo "📚 Generating Swagger documentation..."
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g cmd/$(BINARY_NAME)/main.go -o docs
	@echo "✅ Swagger docs generated in docs/"
	@echo "🌐 Swagger UI will be available at: http://localhost:8080/swagger/index.html"

# Docker
start: swagger
	@echo "🐳 Starting all services with Docker Compose..."
	@$(DOCKER_COMPOSE) up --build -d
	@echo "✅ Services running:"
	@echo "   - Database:    localhost:5433"
	@echo "   - Application: http://localhost:8080"
	@echo "   - Swagger UI:  http://localhost:8080/swagger/index.html"
	@echo "   - Health:      http://localhost:8080/health"

stop:
	@echo "🛑 Stopping all services..."
	@$(DOCKER_COMPOSE) down

restart: stop start

docker-run:
	@echo "🐳 Running application in Docker..."
	@$(DOCKER_COMPOSE) up --build app

logs:
	@$(DOCKER_COMPOSE) logs -f

status:
	@echo "=== Container Status ==="
	@docker ps --filter name=wallet --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "=== Application Health ==="
	@curl -s http://localhost:8080/health || echo "❌ Application not running"

# Testing
test: test-unit test-integration
	@echo "✅ All main tests passed!"

test-unit:
	@echo "🧪 Running UNIT tests..."
	@go test ./internal/handler/... ./internal/service/... -v -short

test-integration:
	@echo "🐘 Running INTEGRATION tests (requires running DB)..."
	@echo "   Make sure DB is running: make dev"
	@go test ./internal/repository/... -v

test-load:
	@echo "📊 Running LOAD tests (long-running)..."
	@echo "   Make sure DB is running: make dev"
	@go test ./internal/load/... -v -timeout=10m

test-e2e:
	@echo "🌐 Running E2E tests (requires running app)..."
	@echo "   Make sure app is running: make start"
	@go test ./internal/e2e/... -v -timeout=5m

test-all: test-unit test-integration test-load test-e2e
	@echo "🎉 ALL TESTS PASSED SUCCESSFULLY!"

# Database
db-shell:
	@echo "💾 Connecting to database..."
	@docker exec -it wallet_postgres psql -U wallet_user -d wallet_db

# Maintenance
clean:
	@echo "🧹 Cleaning up..."
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@go clean -cache
	@rm -f $(BINARY_NAME)
	@docker system prune -f
	@echo "✅ Cleanup completed"

deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

