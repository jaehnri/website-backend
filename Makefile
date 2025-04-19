BINARY_NAME := website-backend
BUILD_DIR := bin

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

.PHONY: build run clean