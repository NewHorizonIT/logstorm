# ==============================
# CONFIG
# ==============================

REDPANDA_CONTAINER ?= deployments-redpanda-1
CLICKHOUSE_CONTAINER ?= deployments-clickhouse-1

DOCKER_RPK = docker exec -it $(REDPANDA_CONTAINER) rpk
DOCKER_CH  = docker exec -it $(CLICKHOUSE_CONTAINER) clickhouse-client
DEPLOYMENT_DIR ?= deployments
CMD_DIR ?= cmd

# ==============================
# INTERNAL FUNCTION
# ==============================

define require_var
	@if [ -z "$($1)" ]; then \
		echo "Error: $1 is required"; \
		exit 1; \
	fi
endef

# ==============================
# REDPANDA
# ==============================

.PHONY: topic-create topic-delete topic-list topic-produce topic-consume ch-query ch-file table-drop

topic-create:
	$(call require_var,TOPIC_NAME)
	$(DOCKER_RPK) topic create $(TOPIC_NAME) --partitions 6 

topic-delete:
	$(call require_var,TOPIC_NAME)
	$(DOCKER_RPK) topic delete $(TOPIC_NAME)

topic-list:
	$(DOCKER_RPK) topic list

topic-produce:
	$(call require_var,TOPIC_NAME)
	$(DOCKER_RPK) topic produce $(TOPIC_NAME)

topic-consume:
	$(call require_var,TOPIC_NAME)
	$(DOCKER_RPK) topic consume $(TOPIC_NAME)

# ==============================
# CLICKHOUSE
# ==============================

ch-query:
	$(call require_var,QUERY)
	$(DOCKER_CH) --query="$(QUERY)"

ch-file:
	$(call require_var,FILE_PATH)
	docker exec -i $(CLICKHOUSE_CONTAINER) clickhouse-client < $(FILE_PATH)

table-drop:
	$(call require_var,TABLE_NAME)
	$(DOCKER_CH) --query="DROP TABLE IF EXISTS $(TABLE_NAME)"

# ==============================
# SYSTEM
# ==============================

.PHONY: up down reset

up:
	docker compose -f $(DEPLOYMENT_DIR)/docker-compose.yml up -d

down:
	docker compose -f $(DEPLOYMENT_DIR)/docker-compose.yml down

reset:
	cd $(DEPLOYMENT_DIR) && docker compose down -v && docker compose up -d
build:
	go build -o bin/server cmd/server/main.go
run: build
	./bin/server

test:
	cd ./tests && run_tests.sh

# ==============================
# HELP
# ==============================

.PHONY: help

help:
	@echo "Available commands:"
	@echo "  make topic-create TOPIC_NAME=name"
	@echo "  make topic-delete TOPIC_NAME=name"
	@echo "  make topic-list"
	@echo "  make topic-produce TOPIC_NAME=name"
	@echo "  make topic-consume TOPIC_NAME=name"
	@echo "  make ch-query QUERY='SELECT 1'"
	@echo "  make ch-file FILE_PATH=init.sql"
	@echo "  make table-drop TABLE_NAME=name"
	@echo "  make up | down | reset"