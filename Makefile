APP_NAME = foodcourt
MAIN_PKG = ./cmd/api/main.go
COVERAGE_FILE = coverage.out

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
	grep -v "mock" $(COVERAGE_FILE) > coverage_clean.out
	mv coverage_clean.out $(COVERAGE_FILE)

	@echo "\nИтоговое покрытие кода:\n"
# Выводим финальную таблицу и итоговый процент (total)
	go tool cover -func=$(COVERAGE_FILE) | grep total

# Позволяет увидеть в браузере, какие именно строчки кода зеленые, а какие красные
cover: test
	go tool cover -html=$(COVERAGE_FILE)

# Команда для быстрой проверки, которая сама за собой уберет файл покрытия
test-once: test
	rm -f $(COVERAGE_FILE)