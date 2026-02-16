# ============================
# Config (override as needed)
# ============================

APP_NAME       ?= jitter-probe
IMAGE_NAME     ?= jitter-probe
IMAGE_TAG      ?= local
FULL_IMAGE     := $(IMAGE_NAME):$(IMAGE_TAG)

K3D_CLUSTER    ?= k3d-local

# Runtime env vars
PING_TARGETS       ?= 1.1.1.1,8.8.8.8
SAMPLE_INTERVAL_MS ?= 500

# Go build settings (default to host OS/arch for local dev)
CGO_ENABLED    ?= 0

# ============================
# Targets
# ============================

.PHONY: help
help:
	@echo ""
	@echo "jitter-probe Makefile"
	@echo ""
	@echo "Local development:"
	@echo "  make run                Run jitter-probe locally with env vars"
	@echo ""
	@echo "Build artifacts:"
	@echo "  make build-bin          Build Go binary for host OS/arch"
	@echo "  make build-linux-amd64  Build linux/amd64 binary"
	@echo "  make build-linux-arm64  Build linux/arm64 binary"
	@echo "  make build-all          Build both linux/amd64 and linux/arm64 binaries"
	@echo "  make build-image        Build Docker image for host arch"
	@echo "  make build-image-all    Build Docker images for amd64 and arm64"
	@echo ""
	@echo "k3d:"
	@echo "  make push-k3d           Import image into k3d cluster"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean"
	@echo ""

# ============================
# Local run
# ============================

.PHONY: run
run:
	@echo ">> Running $(APP_NAME) locally"
	PING_TARGETS="$(PING_TARGETS)" \
	SAMPLE_INTERVAL_MS="$(SAMPLE_INTERVAL_MS)" \
	go run .

# ============================
# Go build
# ============================

.PHONY: build-bin
build-bin:
	@echo ">> Building Go binary (host OS/arch)"
	CGO_ENABLED=$(CGO_ENABLED) go build -o $(APP_NAME)

.PHONY: build-linux-amd64
build-linux-amd64:
	@echo ">> Building Go binary (linux/amd64)"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -o $(APP_NAME)-linux-amd64

.PHONY: build-linux-arm64
build-linux-arm64:
	@echo ">> Building Go binary (linux/arm64)"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -o $(APP_NAME)-linux-arm64

.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64

# ============================
# Docker build
# ============================

.PHONY: build-image
build-image:
	@echo ">> Building Docker image $(FULL_IMAGE)"
	docker build -t $(FULL_IMAGE) .

.PHONY: build-image-amd64
build-image-amd64:
	@echo ">> Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)-amd64"
	docker build --build-arg TARGETARCH=amd64 -t $(IMAGE_NAME):$(IMAGE_TAG)-amd64 .

.PHONY: build-image-arm64
build-image-arm64:
	@echo ">> Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)-arm64"
	docker build --build-arg TARGETARCH=arm64 -t $(IMAGE_NAME):$(IMAGE_TAG)-arm64 .

.PHONY: build-image-all
build-image-all: build-image-amd64 build-image-arm64

# ============================
# Push to k3d
# ============================

.PHONY: push-k3d
push-k3d: build-image
	@echo ">> Importing image into k3d cluster $(K3D_CLUSTER)"
	k3d image import $(FULL_IMAGE) -c $(K3D_CLUSTER)

# ============================
# Cleanup
# ============================

.PHONY: clean
clean:
	@echo ">> Cleaning up"
	rm -f $(APP_NAME) $(APP_NAME)-linux-amd64 $(APP_NAME)-linux-arm64
