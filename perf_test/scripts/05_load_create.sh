#!/usr/bin/env bash
# Нагрузочный тест эндпоинта СОЗДАНИЯ основной сущности:
#   POST /api/owner/restaurants (multipart, owner + CSRF, уникальный Idempotency-Key).
# Очищает restaurant_brand, генерирует N уникальных целей и заливает их vegeta.
# По умолчанию N=100000. Тег результата задаётся 1-м аргументом (default: baseline).
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_DIR"

TAG=${1:-baseline}
N=${N:-100000}
WORKERS=${WORKERS:-30}
source "$RUN_DIR/creds.env"

echo ">> generate $N create targets"
go run ./perf_test/tools/gen -mode=create -n="$N" -out="$TARGETS_DIR/create_${N}.json" \
  -base="$BASE_URL" -session="$SESSION_ID" -csrf="$CSRF_TOKEN"

echo ">> TRUNCATE restaurant_brand (чистый старт для ровно $N строк)"
psql_db restaurant_db -c "TRUNCATE restaurant_brand RESTART IDENTITY CASCADE;" >/dev/null

echo ">> vegeta attack create (workers=$WORKERS, rate=max)"
vegeta attack -targets="$TARGETS_DIR/create_${N}.json" -format=json -rate=0 \
  -max-workers="$WORKERS" -lazy -timeout=30s > "$RESULTS_DIR/create_${TAG}.bin"
vegeta report "$RESULTS_DIR/create_${TAG}.bin" | tee "$RESULTS_DIR/create_${TAG}.txt"
vegeta report -type=json "$RESULTS_DIR/create_${TAG}.bin" > "$RESULTS_DIR/create_${TAG}.json"

echo -n ">> rows in restaurant_brand: "
psql_db restaurant_db -t -c "SELECT count(*) FROM restaurant_brand;"
