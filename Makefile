.PHONY: build clean install test test-unit test-integration test-e2e test-all test-coverage test-bench test-clean

BINARY_NAME=cloud189
VERSION=1.0.0
BUILD_DIR=./build

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/cloud189

build-windows:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/cloud189

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

clean:
	rm -rf $(BUILD_DIR)
	go clean

deps:
	go mod download
	go mod tidy

run:
	go run ./cmd/cloud189

release: build-linux build-windows build-mac

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/cloud189

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/cloud189

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/cloud189

test:
	go test -v -count=1 ./...

test-unit:
	go test -v -count=1 ./internal/... ./pkg/...

test-integration:
	go test -v -count=1 -tags=integration ./test/integration/...

test-e2e:
	go test -v -count=1 -tags=e2e ./test/e2e/...

test-all:
	go test -v -count=1 ./... -tags=integration,e2e

test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench:
	go test -bench=. -benchmem ./test/benchmark/...

test-clean:
	go clean -testcache

test-crypto:
	go test -v ./internal/crypto/...

test-utils:
	go test -v ./pkg/utils/...

test-api:
	go test -v ./internal/api/...

test-commands:
	go test -v ./internal/commands/...