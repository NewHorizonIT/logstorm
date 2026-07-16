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