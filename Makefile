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
	@echo "  make test      - Запустить тесты"
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

test:
	@echo "Запуск тестов..."
	@go test ./...

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