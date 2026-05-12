# Build stage
FROM golang:1.25.3 AS builder

WORKDIR /app

# 1. Копируем только зависимости → максимизируем кэширование
COPY go.mod go.sum ./
RUN go mod download

# 2. Копируем исходный код
#    Используем .dockerignore, чтобы исключить ненужное
COPY . .

# 3. Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

WORKDIR /app 

# Копируем бинарник
COPY --from=builder /app/main .

# Создаём папку для загрузок
RUN mkdir -p ./uploads

EXPOSE 8080

# Запускаем с явной временной зоной (опционально, но рекомендуется)
ENV TZ=Europe/Moscow

CMD ["./main"]