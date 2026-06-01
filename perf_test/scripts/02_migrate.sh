#!/usr/bin/env bash
# Накатывает миграции схемы для сервисов, нужных тесту (auth, user, restaurant),
# и снимает исходный DDL restaurant_db в init.sql (состояние ДО оптимизаций —
# до версии 8 включительно, т.е. без перф-индексов 000009/000010).
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_DIR"

# restaurant поднимаем строго до версии 8 (baseline без перф-индексов).
echo ">> migrate auth (up)";       migrate -path ./services/auth/db/migrations       -database "$(db_url auth)"       up        2>&1 | tail -2 || true
echo ">> migrate user (up)";       migrate -path ./services/user/db/migrations       -database "$(db_url user)"       up        2>&1 | tail -2 || true
echo ">> migrate restaurant -> 8"; migrate -path ./services/restaurant/db/migrations -database "$(db_url restaurant)" goto 8    2>&1 | tail -2 || true

# user может «загрязниться» на сид-данных wordle (не влияет на регистрацию/логин).
migrate -path ./services/user/db/migrations -database "$(db_url user)" force 7 >/dev/null 2>&1 || true

echo ">> dump baseline DDL -> perf_test/init.sql"
$DOCKER exec "$PG_CONTAINER" pg_dump -U "$DB_USER" -d restaurant_db --schema-only --no-owner --no-privileges > "$PERF_DIR/init.sql"
grep -c 'CREATE TABLE' "$PERF_DIR/init.sql" | xargs echo ">> tables in init.sql:"
