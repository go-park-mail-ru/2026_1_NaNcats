# Шаг 1: Сборка
FROM golang:1.26.1-alpine AS builder
WORKDIR /app

RUN apk add --no-cache curl gcc musl-dev

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG BUILD_PATH

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 GOOS=linux go build -o /app/bin/service ${BUILD_PATH}

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/service ./main
COPY --from=builder /app/migrate ./migrate
COPY --from=builder /app/services ./services

EXPOSE 8080 50051 50052 50053 50054 50055 50056 50057
CMD ["./main"]
