# Шаг 1: Сборка
FROM golang:1.26.1-alpine AS builder
WORKDIR /app

RUN apk add --no-cache curl gcc musl-dev

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/api

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/db/migrations ./db/migrations

EXPOSE 8080
CMD ["./main"]