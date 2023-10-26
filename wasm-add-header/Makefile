# Makefile for wasm-add-header

# Go parameters
GOCMD = go
TINYGOCMD = tinygo
TINYGOBUILD = $(TINYGOCMD) build
GOTEST = $(GOCMD) test
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get

# WASM binary name
APP_NAME = wasm-add-header
TARGET_DIR = target/wasi/linux/amd64
WASM = $(TARGET_DIR)/$(APP_NAME).wasm

# Docker parameters
DOCKER = docker
DOCKER_IMAGE_NAME = wasm-add-header-image
DOCKER_IMAGE_TAG = latest
DOCKER_REPO = sreeharikmarar/wasm-add-header

# Default target
all: build

# Build the application
build:
	@mkdir -p $(TARGET_DIR)
	$(TINYGOBUILD) -o $(WASM) -scheduler=none -target=wasi ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(WASM)

# Run tests
test:
	$(GOTEST) -v ./...

# Install dependencies
get:
	$(GOGET)

# Build Docker image
docker-build:
	$(DOCKER) build --platform linux/amd64 --build-arg WASM_BINARY_PATH=$(WASM) -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

# Push Docker image to a registry
docker-push:
	$(DOCKER) tag $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) $(DOCKER_REPO):$(DOCKER_IMAGE_TAG)
	$(DOCKER) push $(DOCKER_REPO):$(DOCKER_IMAGE_TAG)

.PHONY: all build clean test get docker-build docker-push