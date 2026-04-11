# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

# Копируем зависимости и скачиваем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник (CGO отключён для совместимости с Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS-запросов к ml-service
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder-стадии
COPY --from=builder /app/main .

# Создаём папку для загрузок (должна совпадать с uploadDir в main.go)
RUN mkdir -p ./uploads

# Экспонируем порт (внутри контейнера)
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]