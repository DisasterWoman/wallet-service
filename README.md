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

🧪 Тестирование
Запуск всех тестов:
make test-all
Отдельные группы тестов:
make test-unit           # Быстрые unit тесты
make test-integration    # Интеграционные тесты с БД
make test-load           # Нагрузочные тесты (684+ RPS)
make test-e2e            # End-to-end тесты API
Пример нагрузочного тестирования:
# 1000 запросов в секунду на один кошелёк
make test-load

Результаты тестов:
✅ 684 RPS на операциях пополнения
✅ Защита от race condition через SELECT FOR UPDATE
✅ Консистентность баланса при конкурентных операциях


🐳 Docker
Сборка и запуск:
make build          # Сборка бинарника
make start          # Запуск всех сервисов
make docker-run     # Только приложение

Управление контейнерами:
make stop           # Остановка
make restart        # Перезапуск
make logs           # Просмотр логов
make status         # Статус сервисов

База данных:
make db-shell       # Подключение к PostgreSQL

🔧 Разработка
Локальная разработка:
make dev            # Запуск БД
make run            # Запуск приложения
make test-unit      # Быстрые тесты при разработке

Миграции базы данных:
sql
-- Автоматически выполняется при запуске
CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0
);

Мониторинг здоровья:
curl http://localhost:8080/health
