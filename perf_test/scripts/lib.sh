#!/usr/bin/env bash
set -euo pipefail

PERF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_DIR="$(cd "$PERF_DIR/.." && pwd)"

# --- Параметры изолированного стенда -----------------------------------------
PG_CONTAINER=${PG_CONTAINER:-perf_pg}
REDIS_CONTAINER=${REDIS_CONTAINER:-perf_redis}
RABBIT_CONTAINER=${RABBIT_CONTAINER:-perf_rabbit}
DOCKER_NET=${DOCKER_NET:-perfnet}

DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}

REDIS_URL=${REDIS_URL:-redis://localhost:6379/0}
RABBITMQ_URL=${RABBITMQ_URL:-amqp://guest:guest@localhost:5672/}

BASE_URL=${BASE_URL:-http://localhost:8080}

# Каталоги артефактов.
TARGETS_DIR="$PERF_DIR/targets"
RESULTS_DIR="$PERF_DIR/results"
RUN_DIR="$PERF_DIR/.run"          # pid/log запущенных сервисов, креды
mkdir -p "$TARGETS_DIR" "$RESULTS_DIR" "$RUN_DIR"

export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"

DOCKER="docker"
if ! docker ps >/dev/null 2>&1; then DOCKER="sudo docker"; fi

db_url() { echo "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/${1}_db?sslmode=disable"; }

psql_db() { $DOCKER exec -i "$PG_CONTAINER" psql -U "$DB_USER" -d "$1" "${@:2}"; }
