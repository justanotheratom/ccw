APP_NAME := ccw
BIN_DIR := bin

.PHONY: build
build:
	@go build -o $(BIN_DIR)/$(APP_NAME) ./...

.PHONY: test
test:
	@go test ./...

.PHONY: tidy
tidy:
	@go mod tidy
