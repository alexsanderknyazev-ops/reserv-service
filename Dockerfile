# Simple Dockerfile for reserv-service
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости
RUN apk add --no-cache git ca-certificates tzdata

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение (main.go в корне)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o reserv-service .

# Минимальный runtime образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/reserv-service .

USER appuser

# Открываем порт
EXPOSE 8074

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:8074/reserv/health || exit 1

# Запуск
CMD ["./reserv-service"]