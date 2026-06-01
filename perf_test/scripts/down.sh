#!/usr/bin/env bash
# Останавливает Go-сервисы и удаляет контейнеры стенда. Данные не сохраняются.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

for s in gateway restaurant user auth; do
  pid_file="$RUN_DIR/logs/$s.pid"
  [ -f "$pid_file" ] && kill "$(cat "$pid_file")" 2>/dev/null && echo ">> stopped $s"
  rm -f "$pid_file"
done

$DOCKER rm -f "$PG_CONTAINER" "$REDIS_CONTAINER" "$RABBIT_CONTAINER" 2>/dev/null
$DOCKER network rm "$DOCKER_NET" 2>/dev/null || true
echo ">> down complete"
