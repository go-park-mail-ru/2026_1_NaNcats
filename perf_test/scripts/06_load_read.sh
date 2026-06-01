#!/usr/bin/env bash
# Нагрузочный тест эндпоинтов ЧТЕНИЯ основной сущности (публичные, без авторизации):
#   search: GET /api/restaurants/search?q=...      (ILIKE-поиск)
#   list:   GET /api/restaurants/brands?limit&offset (сортировка + deep offset)
# Каждый прогон — 30с на максимальной скорости при WORKERS воркерах.
# Тег результата задаётся 1-м аргументом (baseline | opt).
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_DIR"

TAG=${1:-baseline}
WORKERS=${WORKERS:-20}
DURATION=${DURATION:-30s}
NT=${NT:-2000}

go run ./perf_test/tools/gen -mode=search -n="$NT" -out="$TARGETS_DIR/search.json" -base="$BASE_URL"
go run ./perf_test/tools/gen -mode=list   -n="$NT" -out="$TARGETS_DIR/list.json"   -base="$BASE_URL"

run() { # endpoint targetfile
  local ep=$1 tf=$2
  echo ">> read $ep [$TAG] (workers=$WORKERS, $DURATION)"
  vegeta attack -targets="$TARGETS_DIR/$tf" -format=json -rate=0 \
    -max-workers="$WORKERS" -duration="$DURATION" > "$RESULTS_DIR/read_${ep}_${TAG}.bin"
  vegeta report "$RESULTS_DIR/read_${ep}_${TAG}.bin" | tee "$RESULTS_DIR/read_${ep}_${TAG}.txt"
  vegeta report -type=json "$RESULTS_DIR/read_${ep}_${TAG}.bin" > "$RESULTS_DIR/read_${ep}_${TAG}.json"
}
run search search.json
run list   list.json
