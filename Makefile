BINARY_NAME=wallet-service
DOCKER_COMPOSE=docker-compose

.PHONY: help start stop restart clean test build

help:
	@echo "Доступные команды:"
	@echo "  make start     - Запустить все сервисы через Docker Compose"
	@echo "  make stop      - Остановить все сервисы"
	@echo "  make restart   - Перезапустить все сервисы"
	@echo "  make dev       - Запустить только базу данных"
	@echo "  make run       - Запустить только Go приложение (локально)"
	@echo "  make build     - Собрать бинарный файл"
	@echo "  make clean     - Очистить всё"
	@echo "  make test      - Запустить ВСЕ тесты"
	@echo "  make test-unit - Запустить только unit тесты"
	@echo "  make test-integration - Запустить интеграционные тесты"
	@echo "  make test-load - Запустить нагрузочные тесты"
	@echo "  make test-e2e  - Запустить E2E тесты"
	@echo "  make db-shell  - Подключиться к базе данных"

start:
	@echo "Запуск всех сервисов через Docker Compose..."
	@$(DOCKER_COMPOSE) up --build -d
	@echo "Сервисы запущены:"
	@echo "  - База данных: localhost:5433"
	@echo "  - Приложение:  localhost:8081"

dev:
	@echo "Запуск только базы данных..."
	@$(DOCKER_COMPOSE) up -d postgres
	@echo "База данных запущена на localhost:5433"
	@echo "Теперь можно запустить приложение: make run"

run:
	@echo "Запуск Go приложения локально..."
	@DOCKER_CONTAINER=false go run cmd/$(BINARY_NAME)/main.go

stop:
	@echo "Остановка всех сервисов..."
	@$(DOCKER_COMPOSE) down

restart: stop start

clean:
	@echo "Очистка..."
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@go clean -cache
	@rm -f $(BINARY_NAME)
	@docker system prune -f
	@echo "Очистка завершена"

build:
	@echo "Сборка приложения..."
	@go build -o $(BINARY_NAME) cmd/$(BINARY_NAME)/main.go
	@echo "Собран файл: $(BINARY_NAME)"

test: test-unit test-integration
	@echo "✅ Все основные тесты пройдены!"

test-unit:
	@echo "🚀 Запуск UNIT тестов..."
	@go test ./internal/handler/... ./internal/service/... -v -short

# Интеграционные тесты (требуют БД)
test-integration:
	@echo "🐘 Запуск INTEGRATION тестов (требует запущенной БД)..."
	@echo "   Убедись что БД запущена: make dev"
	@go test ./internal/repository/... -v

# Нагрузочные тесты (долгие)
test-load:
	@echo "📊 Запуск LOAD тестов (долгие)..."
	@echo "   Убедись что БД запущена: make dev"
	@go test ./internal/load/... -v -timeout=10m

# E2E тесты (требуют запущенного приложения)
test-e2e:
	@echo "🌐 Запуск E2E тестов (требует запущенного приложения)..."
	@echo "   Убедись что приложение запущено: make start"
	@go test ./internal/e2e/... -v -timeout=5m

# Полный прогон всех тестов (последовательно)
test-all: test-unit test-integration test-load test-e2e
	@echo "🎉 ВСЕ ТЕСТЫ УСПЕШНО ПРОЙДЕНЫ!"

db-shell:
	@echo "Подключение к базе данных..."
	@docker exec -it wallet_postgres psql -U wallet_user -d wallet_db

status:
	@echo "=== Статус контейнеров ==="
	@docker ps --filter name=wallet
	@echo ""
	@echo "=== Проверка приложения ==="
	@curl -s http://localhost:8081/health || echo "Приложение не запущено"

logs:
	@$(DOCKER_COMPOSE) logs -f

docker-run:
	@echo "Запуск приложения в Docker..."
	@$(DOCKER_COMPOSE) up --build app