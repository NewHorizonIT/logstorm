include .env
.PHONY: up down restart build run test

up:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose down
	docker-compose up -d

# Go related targets
build:
	@echo "Building the Go application..."
	cd app/api && go build -o bin/logstorm ./cmd/server
	@echo "Build completed successfully."

run: build
	@echo "Running the Go application..."
	cd app/api && ./bin/logstorm

test:
	@echo "Running tests..."
	cd app/api && go test ./... -v

# Migrations related targets
migrate-up:
	@echo "Running migrations up..."
	migrate \
	-path $(MIGRATION_DIR) \
	-database "$(DB_URL)" \
	up


migrate-down:
	@echo "Running migrations down..."
	migrate \
	-path $(MIGRATION_DIR) \
	-database "$(DB_URL)" \
	down 


migrate-create:
	@echo "Creating a new migration..."
	migrate create \
	-ext sql \
	-dir $(MIGRATION_DIR) \
	-seq \
	$(name)


migrate-version:
	@echo "Getting the current migration version..."
	migrate \
	-path $(MIGRATION_DIR) \
	-database "$(DB_URL)" \
	version