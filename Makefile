APP_NAME := ccw
BIN_DIR := bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/ccw/ccw/cmd.version=$(VERSION)

.PHONY: build
build:
	@go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .

.PHONY: test
test:
	@go test ./...

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: run-menubar
run-menubar:
	@go build -ldflags "$(LDFLAGS)" -o ccw .
	@open CCWMenubar.xcworkspace

.PHONY: build-menubar
build-menubar:
	@go build -ldflags "$(LDFLAGS)" -o ccw .
	@xcodebuild -workspace CCWMenubar.xcworkspace -scheme CCWMenubar -configuration Debug build
