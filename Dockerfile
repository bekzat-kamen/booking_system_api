# Этап сборки
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache gcc musl-dev

# Копирование файлов зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o main ./cmd/api/main.go

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Установка runtime зависимостей (сертификаты и часовые пояса)
RUN apk add --no-cache ca-certificates tzdata

# Копирование исполняемого файла из этапа сборки
COPY --from=builder /app/main .
# Копирование файла окружения
COPY .env .

# Открытие порта
EXPOSE 8080

# Команда для запуска
CMD ["./main"]
