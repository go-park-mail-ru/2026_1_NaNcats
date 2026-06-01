#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

EMAIL=${OWNER_EMAIL:-owner_perf@test.local}
PASS=${OWNER_PASS:-password123}

echo ">> register $EMAIL (если уже есть — игнор)"
curl -s -X POST "$BASE_URL/api/auth/register" -H 'Content-Type: application/json' \
  -H "Idempotency-Key: reg-$EMAIL" \
  -d "{\"name\":\"PerfOwner\",\"email\":\"$EMAIL\",\"password\":\"$PASS\"}" -o /dev/null

echo ">> promote to owner in user_db"
psql_db user_db -c "UPDATE \"user\" SET user_role='owner' WHERE email='$EMAIL';" >/dev/null

echo ">> login -> capture session_id + csrf"
LJ=$(curl -s -c "$RUN_DIR/cookies.txt" -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' -d "{\"login\":\"$EMAIL\",\"password\":\"$PASS\"}")
CSRF=$(echo "$LJ" | jq -r .csrf_token)
SID=$(grep -i session_id "$RUN_DIR/cookies.txt" | awk '{print $7}')

[ -z "$SID" ] && { echo "!! не удалось получить session_id"; exit 1; }
printf 'SESSION_ID=%s\nCSRF_TOKEN=%s\n' "$SID" "$CSRF" > "$RUN_DIR/creds.env"
echo ">> creds -> $RUN_DIR/creds.env (SID=${SID:0:8}..., CSRF=${CSRF:0:8}...)"
