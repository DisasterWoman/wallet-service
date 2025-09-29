# Базовый образ для сборки 
FROM golang:1.23-alpine AS builder  

# Устанавливаем зависимости
RUN apk add --no-cache git

# Рабочая директория
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /wallet-service ./cmd/wallet-service

# Финальный образ 
FROM alpine:latest

# Устанавливаем зависимости для здоровья приложения
RUN apk add --no-cache ca-certificates

# Копируем бинарник
COPY --from=builder /wallet-service /wallet-service

# Создаем не-root пользователя для безопасности
RUN adduser -D -s /bin/sh appuser
USER appuser

# Экспонируем порт
EXPOSE 8080

# Запуск
CMD ["/wallet-service"]