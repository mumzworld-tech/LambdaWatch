.PHONY: build build-arm64 build-amd64 package clean test

BINARY_NAME := lambdawatch
BUILD_DIR := build
LAYER_DIR := $(BUILD_DIR)/layer/extensions

# Go build flags for smaller binary
LDFLAGS := -s -w
GCFLAGS :=

# Build for current platform
build:
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/extension

# Build for ARM64 (Graviton)
build-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-arm64 ./cmd/extension

# Build for AMD64 (x86_64)
build-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-amd64 ./cmd/extension

# Build for both architectures
build-all: build-arm64 build-amd64

# Package as Lambda Layer (ARM64)
package: build-arm64
	mkdir -p $(LAYER_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME)-arm64 $(LAYER_DIR)/$(BINARY_NAME)
	chmod +x $(LAYER_DIR)/$(BINARY_NAME)
	cd $(BUILD_DIR)/layer && zip -r ../lambdawatch-layer-arm64.zip extensions/

# Package as Lambda Layer (AMD64)
package-amd64: build-amd64
	mkdir -p $(LAYER_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME)-amd64 $(LAYER_DIR)/$(BINARY_NAME)
	chmod +x $(LAYER_DIR)/$(BINARY_NAME)
	cd $(BUILD_DIR)/layer && zip -r ../lambdawatch-layer-amd64.zip extensions/

# Package for both architectures
package-all:
	$(MAKE) package
	rm -rf $(LAYER_DIR)
	$(MAKE) package-amd64

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Deploy layer to AWS (ARM64)
deploy: package
	aws lambda publish-layer-version \
		--layer-name lambdawatch \
		--zip-file fileb://$(BUILD_DIR)/lambdawatch-layer-arm64.zip \
		--compatible-architectures arm64 \
		--compatible-runtimes provided.al2023 provided.al2

# Deploy layer to AWS (AMD64)
deploy-amd64: package-amd64
	aws lambda publish-layer-version \
		--layer-name lambdawatch \
		--zip-file fileb://$(BUILD_DIR)/lambdawatch-layer-amd64.zip \
		--compatible-architectures x86_64 \
		--compatible-runtimes provided.al2023 provided.al2
