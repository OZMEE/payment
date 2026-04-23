# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS builder

# В Alpine по умолчанию нет git и CA-сертификатов. Они нужны для go mod download
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# 1. Копируем только манифесты. Слой будет переиспользован, если зависимости не менялись
COPY go.mod go.sum ./

# 2. Скачиваем зависимости с монтированием кэша (ускоряет CI/CD в разы)
RUN --mount=type=cache,target=/go/pkg/mod \
    GOPROXY=https://proxy.golang.org,direct \
    go mod download && go mod verify

# 3. Копируем исходники
COPY . .

# 4. Собираем статический бинарник. Монтируем кэш компиляции для ускорения
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags '-static'" \
             -o /out/app \
             ./cmd/server

# ──────────────────────────────────────────────────────────────
# Stage 2: Runtime
FROM alpine:3.20 AS runtime

# Минимальный набор для работы Go-бинарника: TLS-сертификаты и часовые пояса
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/app .
COPY --from=builder /src/deploy ./deploy
COPY --from=builder /src/config.yaml .

# Непривилегированный пользователь (аналог USER 1000:1000 в Java-образах)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser:appgroup

EXPOSE 8080

ENTRYPOINT ["/app/app"]