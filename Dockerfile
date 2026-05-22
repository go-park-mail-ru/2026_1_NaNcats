FROM migrate/migrate:v4.17.0 AS migrate

FROM golang:1.26.1-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY --from=migrate /migrate /app/migrate

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG BUILD_PATH

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/bin/service ${BUILD_PATH}

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/api-gateway ./api-gateway
COPY --from=builder /app/bin/service ./main
COPY --from=builder /app/migrate ./migrate
COPY --from=builder /app/services ./services

EXPOSE 8080 50051 50052 50053 50054 50055 50056 50057
CMD ["./main"]
