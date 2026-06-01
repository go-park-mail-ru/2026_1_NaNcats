#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
TAG=${1:-baseline}

COLS='id,owner_profile_id,name,description,promotion_tier,logo_url,banner_url,created_at,updated_at'

echo "===== SEARCH (q='perf' — неселективный, матчит все строки) ====="
psql_db restaurant_db -c "EXPLAIN (ANALYZE, BUFFERS) SELECT $COLS FROM \"restaurant_brand\" WHERE name ILIKE '%perf%' OR description ILIKE '%perf%' ORDER BY promotion_tier DESC, id ASC LIMIT 20 OFFSET 0;" | tee "$RESULTS_DIR/explain_search_${TAG}.txt"

echo "===== SEARCH (q='xyz' — селективный) ====="
psql_db restaurant_db -c "EXPLAIN (ANALYZE, BUFFERS) SELECT $COLS FROM \"restaurant_brand\" WHERE name ILIKE '%xyz%' OR description ILIKE '%xyz%' ORDER BY promotion_tier DESC, id ASC LIMIT 20 OFFSET 0;" | tee "$RESULTS_DIR/explain_search_${TAG}_selective.txt"

echo "===== LIST (deep offset 50000) ====="
psql_db restaurant_db -c "EXPLAIN (ANALYZE, BUFFERS) SELECT $COLS FROM \"restaurant_brand\" ORDER BY promotion_tier DESC, id ASC LIMIT 20 OFFSET 50000;" | tee "$RESULTS_DIR/explain_list_${TAG}.txt"
