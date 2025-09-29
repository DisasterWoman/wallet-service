# Базовый образ для сборки (обновлено до Go 1.24)
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости
RUN apk add --no-cache git postgresql-client

# Копируем go.mod и go.sum для кэширования зависимостей
WORKDIR /app
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем остальные файлы
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /wallet-service ./cmd/wallet-service

# Финальный образ 
FROM alpine:latest

# Копируем бинарник
COPY --from=builder /wallet-service /wallet-service

# Запуск
CMD ["/wallet-service"]
