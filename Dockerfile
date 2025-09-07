# Этап сборки
FROM golang:1.23-alpine AS builder
WORKDIR /app
# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download
# Копируем весь исходный код
COPY . .
# Компилируем бинарник
RUN go build -o minio-photo-service ./cmd/app/main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app
# Копируем бинарник из этапа сборки
COPY --from=builder /app/minio-photo-service .
# Копируем .env файл для конфигурации
COPY .env .
# Указываем команду запуска
CMD ["./minio-photo-service"]