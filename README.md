# 💰 Wallet Service

**REST API для управления виртуальными кошельками с поддержкой конкурентных операций.**

[![Go](https://img.shields.io/badge/Go-1.24-blue)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-green)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-20.10-blue)](https://www.docker.com/)

---

## 📌 О проекте

Сервис предоставляет API для работы с виртуальными кошельками:
- Пополнение (`DEPOSIT`) и списание (`WITHDRAW`) средств.
- Получение текущего баланса.
- Поддержка **1000+ RPS** на один кошелёк (блокировки на уровне строк).

---

## 🚀 Быстрый старт

### Предварительные требования
- [Go 1.24+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) + [Docker Compose](https://docs.docker.com/compose/)
- [PostgreSQL 15+](https://www.postgresql.org/) (опционально, если не используешь Docker)

---

### Установка и запуск

#### 1. Клонируй репозиторий
```bash
git clone https://github.com/DisasterWoman/wallet-service.git
cd wallet-service

2. Настройка окружения
Скопируй .env.example в .env и отредактируй при необходимости:
cp .env.example .env

3. Запуск через Docker (рекомендуется)
 make start

Приложение: http://localhost:8080
База данных: localhost:5433

4. Локальный запуск (без Docker)
# Убедись, что PostgreSQL запущен локально
cp .env.local .env  # Использует localhost:5433
make run

📦 Структура проекта
wallet-service/
├── cmd/               # Точка входа
├── internal/
│   ├── handler/       # HTTP-хендлеры
│   ├── repository/    # Работа с PostgreSQL
│   ├── service/       # Бизнес-логика
│   └── models/        # Модели данных
├── migrations/        # SQL-миграции
├── Dockerfile         # Контейнер для Go-приложения
├── docker-compose.yml # Оркестрация (PostgreSQL + App)
├── Makefile           # Автоматизация задач
└── README.md          # Документация

🧪 Тестирование
Запуск тестов
make test
Тестирование конкурентности
Используй wrk для симуляции нагрузки (1000 RPS):
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/wallet -s post.lua


🐳 Docker
Сборка образа
make build
Запуск контейнеров
docker compose up -d
Подключение к базе данных
make db-shell