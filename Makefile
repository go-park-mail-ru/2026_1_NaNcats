# Загружаем переменные из .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME = foodcourt
GATEWAY_PKG = ./api-gateway/cmd/api/main.go
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

DB_HOST ?= localhost
DB_PORT ?= 5432

RABBITMQ_URL ?= amqp://guest:guest@localhost:5672/
REDIS_URL ?= redis://localhost:6379/0

MICROSERVICES = auth user restaurant cart address payment order support analytics
ALL_SERVICES = $(MICROSERVICES) api-gateway

$(shell mkdir -p .tmp_pids)

.PHONY: all run-all run stop-all stop status build clean test gen cover migrate-create migrate-up migrate-down migrate-up-all migrate-down-all swagger proto logs losg-api logs-db logs-redis logs-clear

# Команда по умолчанию
all: run-all

# Запуск проекта целиком
run-all:
	@echo "Запуск всех микросервисов и API Gateway..."
	@for service in $(ALL_SERVICES); do \
		$(MAKE) run s=$$service; \
	done
	@echo "Все сервисы запущены в фоне, логи и PID процессов в директории .tmp_pids/"

# Запуск конкретного сервиса.
run:
	@if [ -z "$(s)" ]; then echo "Укажите сервис"; exit 1; fi
	@if [ -f .tmp_pids/$(s).pid ]; then echo "Сервис уже запущен"; exit 1; fi
	@echo "Запускаем..."
	@if [ "$(s)" = "api-gateway" ]; then \
		CONFIG_PATH=api-gateway/config.yaml \
		RABBITMQ_URL='$(RABBITMQ_URL)' \
		REDIS_URL='$(REDIS_URL)' \
		nohup go run $(GATEWAY_PKG) > .tmp_pids/$(s).log 2>&1 & echo $$! > .tmp_pids/$(s).pid; \
	else \
		CONFIG_PATH=services/$(s)/config.yaml \
		DATABASE_URL='$(call get_db_url,$(s))' \
		RABBITMQ_URL='$(RABBITMQ_URL)' \
		REDIS_URL='$(REDIS_URL)' \
		nohup go run ./services/$(s)/cmd/main.go > .tmp_pids/$(s).log 2>&1 & echo $$! > .tmp_pids/$(s).pid; \
	fi
	@echo "Запуск $(s) завершен"

# Команда для принудительной очистки портов занятых микросервисами
kill-ports:
	@for port in 8080 50051 50052 50053 50054 50055 50056 50057 50058; do \
		PID=$$(lsof -t -i:$$port); \
		if [ -n "$$PID" ]; then \
			kill -9 $$PID; \
		fi; \
	done
	rm -f .tmp_pids/*.pid

# Остановка всего
stop-all:
	@echo "Останавливаем все сервисы..."
	@for service in $(ALL_SERVICES); do \
		$(MAKE) stop s=$$service; \
	done
	@echo "Все сервисы остановлены"

# Порты сервисов
PORT_api-gateway = 8080
PORT_address     = 50051
PORT_user        = 50052
PORT_restaurant  = 50053
PORT_auth        = 50054
PORT_cart        = 50055
PORT_payment     = 50056
PORT_order       = 50057
PORT_support     = 50058

# Остановка конкретного сервиса
stop:
	@if [ -z "$(s)" ]; then echo "Укажите сервис"; exit 1; fi
	@if [ -f .tmp_pids/$(s).pid ]; then \
		kill -15 `cat .tmp_pids/$(s).pid` 2>/dev/null || true; \
		rm -f .tmp_pids/$(s).pid; \
	fi
	@PORT=$(PORT_$(s)); \
	if [ -n "$$PORT" ]; then \
		CHILD_PID=$$(lsof -ti :$$PORT 2>/dev/null | head -1); \
		if [ -n "$$CHILD_PID" ]; then \
			kill -15 $$CHILD_PID 2>/dev/null || true; \
			sleep 1; \
			if kill -0 $$CHILD_PID 2>/dev/null; then kill -9 $$CHILD_PID 2>/dev/null || true; fi; \
		fi; \
	fi
	@echo "Сервис $(s) остановлен"

DC = docker compose

# Запуск всех контейнеров в фоне с пересборкой
d-up:
	$(DC) up -d --build

# Остановка и удаление контейнеров
d-down:
	$(DC) down

# Перезапуск всех сервисов
d-restart:
	$(DC) restart

# Полная пересборка без кэша
d-rebuild:
	$(DC) build --no-cache $(s)
	$(DC) up -d $(s)

# Посмотреть статус контейнеров
d-ps:
	$(DC) ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Просмотр работающих сервисов
status:
	@echo "Работающие сервисы:"
	@ls .tmp_pids/*.pid 2>/dev/null | sed 's/.tmp_pids\///' | sed 's/.pid//' || echo "Нет запущенных сервисов"


# Сборка бинарника
build:
	go build -o $(APP_NAME) $(GATEWAY_PKG)

# Удаление бинарника и временных файлов
clean: stop-all
	rm -f $(APP_NAME)
	rm -f $(COVERAGE_FILE)
	rm -rf .tmp_pids
	find shared/proto -name "*.pb.go" -type f -delete

# Генерация моков
gen:
	-go generate ./...

# Генерация proto файлов
PROTO_DIR = shared/proto
# Находим все .proto файлы в директории рекурсивно
PROTO_FILES = $(shell find $(PROTO_DIR) -name "*.proto")

proto:
	protoc -I $(PROTO_DIR) \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	go generate ./shared/proto/...

EXCLUDE_PATTERNS = "mock|\.pb\.go|_easyjson\.go|_tracing_mw\.go|main\.go|config\.go|docs\.go|tracer\.go|migrator\.go|debug_gen\.go"

# Тестирование с правильным подсчетом покрытия
test:
	@echo "Запуск тестов...\n"
# Прогоняем тесты и записываем сырой результат в файл покрытия
	-go test -coverprofile=$(COVERAGE_FILE) ./...

	@echo "\nОчистка покрытия от моков...\n"
# Удаляем все строчки, где есть слово "mock", из файла покрытия
	grep -Ev $(EXCLUDE_PATTERNS) $(COVERAGE_FILE) > coverage_clean.out
	mv coverage_clean.out $(COVERAGE_FILE)

	@echo "\nИтоговое покрытие кода:\n"
# Выводим финальную таблицу и итоговый процент (total)
	go tool cover -func=$(COVERAGE_FILE) | grep total

# Позволяет увидеть в браузере, какие именно строчки кода зеленые, а какие красные
cover: test
# Генерируем статический HTML файл
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "HTML отчет создан: $(COVERAGE_HTML)"
# Пытаемся открыть его (команда xdg-open для Linux, open для Mac)
	xdg-open $(COVERAGE_HTML) || open $(COVERAGE_HTML) || echo "Открой $(COVERAGE_HTML) в браузере вручную"

# --- РАБОТА С БД ---

define get_db_url
postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(1)_db?sslmode=disable
endef

# Создать новую миграцию (например: make migrate-create s=user name=add_users)
migrate-create:
	@if [ -z "$(s)" ]; then echo "Укажите сервис"; exit 1; fi
	migrate create -ext sql -dir services/$(s)/db/migrations -seq $(name)

# Накатить миграции для всех сервисов
migrate-up-all:
	@echo "Накатываем миграции для всех сервисов..."
	@for service in $(MICROSERVICES); do \
		$(MAKE) migrate-up s=$$service; \
	done
	@echo "Накатываем миграции для ClickHouse..."
	@$(MAKE) migrate-ch-up

# Накатить миграции конкретного сервиса (make migrate-up s=user)
migrate-up:
	@if [ -z "$(s)" ]; then echo "Укажите сервис"; exit 1; fi
	@echo "Migrating UP: $(s)..."
	migrate -path ./services/$(s)/db/migrations -database "$(call get_db_url,$(s))" up

# Откатить миграции для всех сервисов
migrate-down-all:
	@echo "Откатываем миграции для всех сервисов..."
	@for service in $(MICROSERVICES); do \
		$(MAKE) migrate-down s=$$service; \
	done

# Откатить последнюю миграцию
migrate-down:
	@if [ -z "$(s)" ]; then echo "Укажите сервис"; exit 1; fi
	@echo "Migrating DOWN: $(s)..."
	migrate -path ./services/$(s)/db/migrations -database "$(call get_db_url,$(s))" down

CH_HOST ?= localhost
CH_PORT ?= 9000
CH_DB   ?= analytics
CH_USER     ?= $(CLICKHOUSE_USER)
CH_PASSWORD ?= $(CLICKHOUSE_PASSWORD)

CLICKHOUSE_URL = "clickhouse://$(CH_HOST):$(CH_PORT)/$(CH_DB)?username=$(CH_USER)&password=$(CH_PASSWORD)&x-multi-statement=true"

# Создать миграцию для ClickHouse
migrate-ch-create:
	migrate create -ext sql -dir services/analytics/db/migrations -seq $(name)

# Накатить миграции ClickHouse
migrate-ch-up:
	@echo "Migrating ClickHouse up..."
	migrate -path ./services/analytics/db/migrations -database $(CLICKHOUSE_URL) up

# Откатить миграции ClickHouse
migrate-ch-down:
	@echo "Migrating ClickHouse down..."
	migrate -path ./services/analytics/db/migrations -database $(CLICKHOUSE_URL) down

swagger:
	swag init -g $(GATEWAY_PKG) --parseInternal --parseDependency

# --- ЛОГИ ---

logs:
	docker compose logs -f

# Логи только бэкенда (Go приложение)
logs-api:
	tail -f .tmp_pids/api_gateway.log

# Логи конкретного микросервиса (пример: make logs-s s=user)
logs-s:
	@if [ -z "$(s)" ]; then echo "Укажите сервис"; exit 1; fi
	tail -f .tmp_pids/$(s).log

# Логи базы данных (PostgreSQL + PostGIS)
logs-db:
	docker logs -f postgres_db

# Логи Redis
logs-redis:
	docker logs -f redis_db

# Посмотреть последние 100 строк логов бэкенда и ждать новых
logs-tail:
	docker logs -f --tail 100 go_backend

# Очистить логи контейнеров
logs-clear:
	sudo sh -c "truncate -s 0 /var/lib/docker/containers/*/*-json.log"
	rm -f .tmp_pids/*.log
