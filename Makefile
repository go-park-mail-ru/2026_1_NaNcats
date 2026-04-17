# Загружаем переменные из .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME = foodcourt
MAIN_PKG = ./cmd/api/main.go
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

.PHONY: all run build clean test gen cover migrate-create migrate-up migrate-down swagger proto

# Команда по умолчанию
all: run

# Запуск проекта
run:
	go run $(MAIN_PKG)

# Сборка бинарника
build:
	go build -o $(APP_NAME) $(MAIN_PKG)

# Удаление бинарника и временных файлов
clean:
	rm -f $(APP_NAME)
	rm -f $(COVERAGE_FILE)
	find shared/proto -name "*.pb.go" -type f -delete

# Генерация моков
gen:
	go generate ./...

# Генерация proto файлов
proto:
	protoc -I shared/proto \
			--go_out=shared/proto --go_opt=paths=source_relative \
			--go-grpc_out=shared/proto --go-grpc_opt=paths=source_relative \
			shared/proto/address/address.proto \
			shared/proto/auth/auth.proto \
			shared/proto/cart/cart.proto \
			shared/proto/order/order.proto \
			shared/proto/payment/payment.proto \
			shared/proto/restaurant/restaurant.proto \
			shared/proto/user/user.proto

# Тестирование с правильным подсчетом покрытия
test:
	@echo "Запуск тестов...\n"
# Прогоняем тесты и записываем сырой результат в файл покрытия
	-go test -coverprofile=$(COVERAGE_FILE) ./...

	@echo "\nОчистка покрытия от моков...\n"
# Удаляем все строчки, где есть слово "mock", из файла покрытия
	grep -Ev "mock|main.go|tracer.go|migrator.go|_easyjson|\.pb\.go" $(COVERAGE_FILE) > coverage_clean.out
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

# Создать новую миграцию (например: make migrate-create name=add_users_table)
migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

# Накатить миграции
migrate-up:
	docker compose exec backend ./migrate -path db/migrations -database "$(DATABASE_URL)" up

# Откатить последнюю миграцию
migrate-down:
	docker compose exec backend ./migrate -path db/migrations -database "$(DATABASE_URL)" down

swagger:
	swag init -g $(MAIN_PKG) --parseInternal --parseDependency

# --- ЛОГИ ---

logs:
	docker compose logs -f

# Логи только бэкенда (Go приложение)
logs-api:
	docker logs -f go_backend

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
# это требует прав sudo или доступа к папке docker
logs-clear:
	sudo sh -c "truncate -s 0 /var/lib/docker/containers/*/*-json.log"
