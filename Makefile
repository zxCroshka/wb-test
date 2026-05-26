.PHONY: help build run test bench docker-up docker-down clean fmt lint check

# Переменные
BINARY_NAME=trending-service
CONFIG_PATH=configs/config.yaml

GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

help: ## Показать справку
	@echo "${GREEN}Доступные команды:${NC}"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-18s${NC} %s\n", $$1, $$2}'

build: ## Собрать бинарник
	@echo "${GREEN}Building...${NC}"
	go build -o bin/${BINARY_NAME} ./cmd/server

run: build ## Запустить сервис
	@echo "${GREEN}Starting service...${NC}"
	./bin/${BINARY_NAME} -config ${CONFIG_PATH}

fmt: ## Форматировать код
	@echo "${GREEN}Formatting...${NC}"
	go fmt ./...

lint: ## Запустить линтер
	@echo "${GREEN}Running linter...${NC}"
	command -v golangci-lint >/dev/null 2>&1 && golangci-lint run --timeout=5m ./... || echo "golangci-lint not installed, run 'make setup'"

test: ## Запустить тесты
	@echo "${GREEN}Running tests...${NC}"
	go test -race  ./...

test-verbose: ## Запустить тесты с подробным выводом
	@echo "${GREEN}Running tests with verbose...${NC}"
	go test -race  -v ./...

bench: ## Запустить бенчмарки
	@echo "${GREEN}Running benchmarks...${NC}"
	go test -bench=. -benchmem ./...

bench-cache: ## Бенчмарки кэша
	@echo "${GREEN}Running cache benchmarks...${NC}"
	go test -bench=BenchmarkTopCache -benchmem ./internal/cache/...

bench-detector: ## Бенчмарки детектора
	@echo "${GREEN}Running detector benchmarks...${NC}"
	go test -bench=BenchmarkAnomalyDetector -benchmem ./internal/detector/...

bench-stoplist: ## Бенчмарки стоп-листа
	@echo "${GREEN}Running stoplist benchmarks...${NC}"
	go test -bench=BenchmarkStopList -benchmem ./internal/stoplist/...

bench-domain: ## Бенчмарки окна
	@echo "${GREEN}Running domain benchmarks...${NC}"
	go test -bench=BenchmarkSlidingWindow -benchmem ./internal/domain/...


check: fmt test ## Проверить код (формат + тесты)
	@echo "${GREEN}All checks passed!${NC}"

docker-up: ## Запустить Docker Compose
	@echo "${GREEN}Starting Docker...${NC}"
	cd deployments && docker compose up -d
	@echo "${YELLOW}API: http://localhost:8080${NC}"
	@echo "${YELLOW}Kafka UI: http://localhost:8081${NC}"
	@echo "${YELLOW}Prometheus: http://localhost:9091${NC}"

docker-down: ## Остановить Docker
	@echo "${GREEN}Stopping Docker...${NC}"
	cd deployments && docker compose down

docker-down-volumes: ## Остановить Docker и удалить данные
	@echo "${RED}Stopping Docker and removing volumes...${NC}"
	cd deployments && docker compose down -v

docker-logs: ## Логи сервиса
	cd deployments && docker compose logs -f app

docker-logs-kafka: ## Логи Kafka
	cd deployments && docker compose logs -f kafka

docker-build: ## Пересобрать Docker образ
	@echo "${GREEN}Rebuilding Docker image...${NC}"
	cd deployments && docker compose build --no-cache app

docker-shell: ## Зайти в контейнер
	docker exec -it trending-service sh

# Добавьте в Makefile

docker-restart: ## Полный перезапуск: очистка, пересборка, поднятие
	@echo "${GREEN}Full restart: cleaning, rebuilding, starting...${NC}"
	cd deployments && docker compose down -v
	@echo "${GREEN}Removing old images...${NC}"
	docker rmi deployments-app -f 2>/dev/null || true
	@echo "${GREEN}Pruning Docker cache...${NC}"
	docker system prune -f 2>/dev/null || true
	@echo "${GREEN}Rebuilding image...${NC}"
	cd deployments && docker compose build --no-cache
	@echo "${GREEN}Starting services...${NC}"
	cd deployments && docker compose up -d
	@echo "${YELLOW}Waiting for services to be ready (15 sec)...${NC}"
	sleep 15
	@echo "${GREEN}Creating Kafka topic...${NC}"
	docker exec trending-kafka kafka-topics --create \
		--topic search-events \
		--bootstrap-server localhost:9092 \
		--partitions 1 \
		--replication-factor 1 \
		--if-not-exists 2>/dev/null || true
	@echo "${GREEN}All done!${NC}"
	@echo "${YELLOW}API: http://localhost:8080${NC}"
	@echo "${YELLOW}Kafka UI: http://localhost:8081${NC}"
	@echo "${YELLOW}Prometheus: http://localhost:9091${NC}"

clean: ## Очистить
	@echo "${GREEN}Cleaning...${NC}"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

deps: ## Установить зависимости
	@echo "${GREEN}Installing dependencies...${NC}"
	go mod download
	go mod tidy

setup: ## Установить инструменты разработки
	@echo "${GREEN}Installing tools...${NC}"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

kafka-create-topic: ## Создать топик
	@echo "${GREEN}Creating topic...${NC}"
	docker exec trending-kafka kafka-topics --create \
		--topic search-events \
		--bootstrap-server localhost:9092 \
		--partitions 3 \
		--replication-factor 1 \
		--if-not-exists

kafka-send: ## Отправить тестовое сообщение
	@echo "${GREEN}Sending test message...${NC}"
	docker exec trending-kafka sh -c 'echo "{\"query\":\"iphone\",\"user_id\":\"user123\",\"timestamp\":'$$(date +%s)000'}" | kafka-console-producer --topic search-events --bootstrap-server localhost:9092 --property "parse.key=false"'
kafka-consume: ## Потребить сообщения (для отладки)
	@echo "${GREEN}Consuming messages...${NC}"
	docker exec trending-kafka kafka-console-consumer \
		--topic search-events \
		--bootstrap-server localhost:9092 \
		--from-beginning

api-top: ## Запрос топа
	curl -s "http://localhost:8080/top?limit=10" | jq . 2>/dev/null || curl -s "http://localhost:8080/top?limit=10"

api-health: ## Health check
	curl -s "http://localhost:8080/health" | jq . 2>/dev/null || curl -s "http://localhost:8080/health"

api-stoplist-add: ## Добавить слово в стоп-лист (make WORD=spam api-stoplist-add)
	@echo "${GREEN}Adding '$(WORD)' to stoplist...${NC}"
	curl -s -X POST "http://localhost:8080/stoplist" \
		-H "Content-Type: application/json" \
		-d '{"word":"$(WORD)"}' | jq . 2>/dev/null || curl -s -X POST "http://localhost:8080/stoplist" -H "Content-Type: application/json" -d '{"word":"$(WORD)"}'

api-stoplist-get: ## Получить стоп-лист
	curl -s "http://localhost:8080/stoplist" | jq . 2>/dev/null || curl -s "http://localhost:8080/stoplist"

api-stoplist-check: ## Проверить слово (make WORD=spam api-stoplist-check)
	curl -s "http://localhost:8080/stoplist/check?word=$(WORD)" | jq . 2>/dev/null || curl -s "http://localhost:8080/stoplist/check?word=$(WORD)"

prometheus-targets: ## Проверить targets в Prometheus
	curl -s "http://localhost:9091/api/v1/targets" | jq '.data.activeTargets[] | {job: .labels.job, status: .health}' 2>/dev/null || echo "Prometheus not running or jq not installed"

all: deps fmt test build ## Всё сразу
	@echo "${GREEN}All done!${NC}"
