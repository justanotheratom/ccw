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

.PHONY: run-menubar
run-menubar:
	@go build -o ccw .
	@open CCWMenubar.xcworkspace

.PHONY: build-menubar
build-menubar:
	@go build -o ccw .
	@xcodebuild -workspace CCWMenubar.xcworkspace -scheme CCWMenubar -configuration Debug build
