#!/usr/bin/env bash
# Применяет оптимизации (миграции restaurant 000009 pg_trgm и 000010 btree)
# пошагово через `migrate goto` и делает ANALYZE. Аргумент — целевая версия:
#   ./07_optimize.sh 9    # только триграммные индексы (оптимизация поиска)
#   ./07_optimize.sh 10   # + btree(promotion_tier DESC, id) (оптимизация листинга)
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_DIR"

TARGET=${1:-10}
echo ">> migrate restaurant -> $TARGET"
migrate -path ./services/restaurant/db/migrations -database "$(db_url restaurant)" goto "$TARGET" 2>&1 | tail -2
psql_db restaurant_db -c "ANALYZE restaurant_brand;" >/dev/null
echo ">> indexes on restaurant_brand:"
psql_db restaurant_db -c "SELECT indexrelname, pg_size_pretty(pg_relation_size(indexrelid)) FROM pg_stat_user_indexes WHERE relname='restaurant_brand' ORDER BY 1;"
