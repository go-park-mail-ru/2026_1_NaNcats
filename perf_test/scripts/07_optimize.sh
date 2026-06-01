#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_DIR"

TARGET=${1:-10}
echo ">> migrate restaurant -> $TARGET"
migrate -path ./services/restaurant/db/migrations -database "$(db_url restaurant)" goto "$TARGET" 2>&1 | tail -2
psql_db restaurant_db -c "ANALYZE restaurant_brand;" >/dev/null
echo ">> indexes on restaurant_brand:"
psql_db restaurant_db -c "SELECT indexrelname, pg_size_pretty(pg_relation_size(indexrelid)) FROM pg_stat_user_indexes WHERE relname='restaurant_brand' ORDER BY 1;"
