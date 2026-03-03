# Congiguration ingestion service
BUILD_DIR := bin
BINARY_NAME := ingestion-service
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)

.PHONY: build run

# Tagert 
build:
	go build -o $(BINARY_PATH) cmd/server/main.go

run: build
	$(BINARY_PATH)