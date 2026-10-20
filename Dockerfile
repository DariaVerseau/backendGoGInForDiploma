# syntax=docker/dockerfile:1.7

# ---------- Build stage ----------
ARG GO_VERSION=1.26.9
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# 1. Копируем только зависимости → максимизируем кэширование
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 2. Копируем исходный код
COPY . .

# 3. Собираем бинарник (стат. линковка, без debug-инфо)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/main ./cmd/server

# ---------- Final stage ----------
FROM alpine:3.20

# ca-certificates — для HTTPS, tzdata — для TZ
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Непривилегированный пользователь
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Копируем бинарник (владелец — appuser)
COPY --from=builder --chown=appuser:appgroup /app/main /app/main

# Папка для uploads (владелец — appuser)
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080
ENV TZ=Europe/Moscow

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/main"]