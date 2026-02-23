BINARY_NAME=hypercacheio-server
SOURCE_FILE=main.go
OUT_DIR ?= ../build

.PHONY: all clean build-mac-arm64 build-mac-amd64 build-linux-arm64 build-linux-amd64

all: build-mac-arm64 build-mac-amd64 build-linux-arm64 build-linux-amd64

clean:
	rm -rf $(OUT_DIR)

build-mac-arm64:
	mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 $(SOURCE_FILE)

build-mac-amd64:
	mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(SOURCE_FILE)

build-linux-arm64:
	mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(OUT_DIR)/$(BINARY_NAME)-linux-arm64 $(SOURCE_FILE)

build-linux-amd64:
	mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 $(SOURCE_FILE)
