.PHONY: help
help:
	@echo "💰 Currency Converter Commands:"
	@echo "  make build      - собрать приложение локально"
	@echo "  make run        - запустить приложение"
	@echo "  make test       - запустить тесты"
	@echo "  make docker-up  - запустить Docker Compose"
	@echo "  make docker-down - остановить Docker Compose"
	@echo "  make docker-logs - показать логи приложения"
	@echo "  make clean      - очистить всё"

# 🔨 Собрать приложение локально (без Docker)
.PHONY: build
build:
	@echo "🔨 Собираем приложение..."
	go build -o currency-converter ./cmd/server/

# 🚀 Запустить приложение локально
.PHONY: run
run: build
	@echo "🚀 Запускаем приложение..."
	./currency-converter

# 🧪 Запустить тесты
.PHONY: test
test:
	@echo "🧪 Запускаем тесты..."
	go test ./internal/handler -v

# 🐳 Запустить Docker Compose
.PHONY: docker-up
docker-up:
	@echo "🐳 Запускаем Docker Compose..."
	docker-compose up -d
	@echo "✅ Готово! Открой http://localhost:8080"

# 🛑 Остановить Docker Compose
.PHONY: docker-down
docker-down:
	@echo "🛑 Останавливаем Docker Compose..."
	docker-compose down

# 📋 Показать логи приложения
.PHONY: docker-logs
docker-logs:
	docker-compose logs -f app

# 🧹 Очистить всё
.PHONY: clean
clean:
	@echo "🧹 Очищаем..."
	docker-compose down -v
	rm -f currency-converter
	go clean
	@echo "✅ Всё очищено!"