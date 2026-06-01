#!/usr/bin/env bash
# Собирает и запускает 4 Go-сервиса, нужных тесту: auth, user, restaurant, gateway.
# Каждый сервис ходит в свою БД (<service>_db). Логи и pid — в perf_test/.run/.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_DIR"

mkdir -p "$RUN_DIR/logs"
# Переменные окружения сервисов (S3-креды, дефолтные URL, адреса gRPC) берём из .env.
set -a; [ -f "$REPO_DIR/.env" ] && source "$REPO_DIR/.env"; set +a
export REDIS_URL RABBITMQ_URL

echo ">> building binaries (go build)"
go build -o "$RUN_DIR/svc_auth"       ./services/auth/cmd/main.go
go build -o "$RUN_DIR/svc_user"       ./services/user/cmd/main.go
go build -o "$RUN_DIR/svc_restaurant" ./services/restaurant/cmd/main.go
go build -o "$RUN_DIR/svc_gateway"    ./api-gateway/cmd/api/main.go

start() { # name binary db_name config
  local s=$1 bin=$2 db=$3 cfg=$4
  ( cd "$REPO_DIR"
    [ -n "$db" ] && export DATABASE_URL="$(db_url "$db")"
    CONFIG_PATH="$cfg" nohup "$RUN_DIR/$bin" > "$RUN_DIR/logs/$s.log" 2>&1 &
    echo $! > "$RUN_DIR/logs/$s.pid" )
  echo ">> started $s (pid $(cat "$RUN_DIR/logs/$s.pid"))"
}

start auth       svc_auth       auth       services/auth/config.yaml
start user       svc_user       user       services/user/config.yaml
start restaurant svc_restaurant restaurant services/restaurant/config.yaml
sleep 5
start gateway    svc_gateway    ""         api-gateway/config.yaml
sleep 5

echo ">> health: GET /api/restaurants/brands"
curl -s "$BASE_URL/api/restaurants/brands?limit=1" -o /dev/null -w "HTTP %{http_code}\n" || echo "gateway not ready"
