#!/usr/bin/env bash
# Поднимает изолированный стенд БД/брокера в Docker (отдельно от прод-деплоя)
# и создаёт нужные базы данных. Образы и версии совпадают с docker-compose.yml.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

$DOCKER network create "$DOCKER_NET" 2>/dev/null || true

run_container() {
  local name=$1; shift
  $DOCKER rm -f "$name" >/dev/null 2>&1 || true
  $DOCKER run -d --name "$name" --network "$DOCKER_NET" "$@" >/dev/null
  echo ">> started $name"
}

run_container "$PG_CONTAINER" -p 5432:5432 \
  -e POSTGRES_USER="$DB_USER" -e POSTGRES_PASSWORD="$DB_PASSWORD" -e POSTGRES_DB=postgres \
  --shm-size=256m postgis/postgis:18-3.6-alpine
run_container "$REDIS_CONTAINER" -p 6379:6379 redis:7-alpine
run_container "$RABBIT_CONTAINER" -p 5672:5672 rabbitmq:3-management-alpine

echo ">> waiting for postgres..."
for i in $(seq 1 30); do
  $DOCKER exec "$PG_CONTAINER" pg_isready -U "$DB_USER" >/dev/null 2>&1 && break
  sleep 1
done

for db in auth_db user_db restaurant_db; do
  $DOCKER exec "$PG_CONTAINER" psql -U "$DB_USER" -c "CREATE DATABASE $db;" 2>&1 | head -1
done
echo ">> infra ready"
