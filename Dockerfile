# Этап сборки
# FROM golang:1.24.5-alpine3.20 AS builder
FROM golang:1.24-alpine AS builder

# на будущее
# Сборка:
# Базовая
# docker build -t hwwach:latest .

# bash
# С версионированием
# docker build \
#  --build-arg VERSION=1.0.0 \
#  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
#  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#  -t hwwach:1.0.0 .

# ARG VERSION=dev
# ARG BUILD_TIME
# ARG GIT_COMMIT

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Кэширование зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
# -w — убирает информацию для отладки DWARF
# -s — убирает таблицу символов
#Экономия: ~30% размера бинарника
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o main ./cmd/main.go

# Финальный образ
# Используйте точные версии:
# Преимущество: Воспроизводимые сборки. Образ будет одинаковым через месяц/год.
# FROM alpine:3.20.3
FROM alpine:3.20

WORKDIR /app

# Установка tzdata для часовой зоны и ca-certificates для HTTPS
RUN apk --no-cache add tzdata ca-certificates && \
  addgroup -S appgroup && \
  adduser -S appuser -G appgroup && \
  chown -R appuser:appgroup /app

# Копирование бинарного файла
COPY --from=builder /app/main .

# Копирование миграций (если будут)
RUN if [ -d /app/migrations ]; then mkdir -p ./migrations && cp -r /app/migrations/* ./migrations/ 2>/dev/null || true; fi

# Запуск от непривилегированного пользователя
USER appuser

# Настройка часовой зоны
ENV TZ=Europe/Moscow

# Порт приложения
EXPOSE 8080

# Запуск приложения
CMD ["./main"]

# Требование: Нужен endpoint /health в Go:
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1