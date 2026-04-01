.PHONY: build clean install test

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

test:
	go test -v ./...

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