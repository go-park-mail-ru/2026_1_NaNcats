APP_NAME = foodcourt
MAIN_PKG = ./cmd/api/main.go
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

.PHONY: all run build clean test gen cover

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

# Генерация моков
gen:
	go generate ./...

# Тестирование с правильным подсчетом покрытия
test:
	@echo "Запуск тестов...\n"
# Прогоняем тесты и записываем сырой результат в файл покрытия
	-go test -coverprofile=$(COVERAGE_FILE) ./...

	@echo "\nОчистка покрытия от моков...\n"
# Удаляем все строчки, где есть слово "mock", из файла покрытия
	grep -Ev "mock" $(COVERAGE_FILE) > coverage_clean.out
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


# БД
# Переменные
DB_URL=postgres://user:password@localhost:5432/delivery_db?sslmode=disable

# Создать новую миграцию (например: make migrate-create name=add_users_table)
migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

# Накатить миграции
migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

# Откатить последнюю миграцию
migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down
