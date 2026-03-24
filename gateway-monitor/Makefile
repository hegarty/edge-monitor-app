# ============================
# Config (override as needed)
# ============================

APP_NAME       ?= gateway-monitor
IMAGE_NAME     ?= gateway-monitor
IMAGE_TAG      ?= local
FULL_IMAGE     := $(IMAGE_NAME):$(IMAGE_TAG)

K3D_CLUSTER    ?= k3d-local

REGISTRY       ?= localhost:5000
K3S_REGISTRY   ?= pi-1.local:5000
KUBE_CONTEXT   ?=
CHART          := ./charts/$(APP_NAME)
NAMESPACE      ?= gateway-monitor
HELM_CONTEXT_ARG := $(if $(KUBE_CONTEXT),--kube-context $(KUBE_CONTEXT),)
KUBECTL_CONTEXT_ARG := $(if $(KUBE_CONTEXT),--context $(KUBE_CONTEXT),)

# Runtime env vars
GATEWAY_IP       ?= 192.168.1.1
WAN_TARGET       ?= 1.1.1.1
INTERVAL_SECONDS ?= 2

# Go build settings (default to host OS/arch for local dev)
CGO_ENABLED    ?= 0

# ============================
# Targets
# ============================

.PHONY: help
help:
	@echo ""
	@echo "gateway-monitor Makefile"
	@echo ""
	@echo "Local development:"
	@echo "  make run                Run gateway-monitor locally with env vars"
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
	@echo "Registry:"
	@echo "  make push               Tag and push image to registry"
	@echo ""
	@echo "Helm deploy:"
	@echo "  make deploy             Build, push, and deploy via Helm"
	@echo "  make deploy-k3s         Build, push, and deploy to k3s via Helm values-k3s"
	@echo "  make rollout            Wait for deployment rollout"
	@echo "  make logs               Tail logs for deployed pods"
	@echo "  make describe           Describe deployed pods"
	@echo "  make delete             Uninstall Helm release and clean up"
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
	GATEWAY_IP="$(GATEWAY_IP)" \
	WAN_TARGET="$(WAN_TARGET)" \
	INTERVAL_SECONDS="$(INTERVAL_SECONDS)" \
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
# Registry push
# ============================

.PHONY: push
push: build-image
	@echo ">> Tagging and pushing to registry $(REGISTRY)"
	docker tag $(FULL_IMAGE) $(REGISTRY)/$(APP_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(APP_NAME):$(IMAGE_TAG)

.PHONY: push-k3s
push-k3s: build-image
	@echo ">> Tagging and pushing to k3s registry $(K3S_REGISTRY)"
	docker tag $(FULL_IMAGE) $(K3S_REGISTRY)/$(APP_NAME):$(IMAGE_TAG)
	docker push $(K3S_REGISTRY)/$(APP_NAME):$(IMAGE_TAG)

# ============================
# Helm deploy
# ============================

.PHONY: require-kube-context
require-kube-context:
	@test -n "$(KUBE_CONTEXT)" || (echo "KUBE_CONTEXT is required for Helm and kubectl targets" >&2; exit 1)

.PHONY: deploy
deploy: push require-kube-context
	@echo ">> Deploying $(APP_NAME) via Helm"
	helm upgrade --install $(APP_NAME) $(CHART) \
	  $(HELM_CONTEXT_ARG) \
	  --namespace $(NAMESPACE) \
	  --set image.repository=k3d-edge-registry:5000/$(APP_NAME) \
	  --set image.tag=$(IMAGE_TAG)

.PHONY: deploy-k3s
deploy-k3s: push-k3s require-kube-context
	@echo ">> Deploying $(APP_NAME) to k3s via Helm"
	helm upgrade --install $(APP_NAME) $(CHART) \
	  $(HELM_CONTEXT_ARG) \
	  --namespace $(NAMESPACE) \
	  -f $(CHART)/values-k3s.yaml \
	  --set image.tag=$(IMAGE_TAG)

.PHONY: rollout
rollout: require-kube-context
	@echo ">> Waiting for rollout of $(APP_NAME)"
	kubectl $(KUBECTL_CONTEXT_ARG) rollout status deployment/$(APP_NAME) -n $(NAMESPACE)

.PHONY: logs
logs: require-kube-context
	kubectl $(KUBECTL_CONTEXT_ARG) logs -l app=$(APP_NAME) -f -n $(NAMESPACE)

.PHONY: describe
describe: require-kube-context
	kubectl $(KUBECTL_CONTEXT_ARG) describe pod -l app=$(APP_NAME) -n $(NAMESPACE)

.PHONY: delete
delete: require-kube-context
	helm uninstall $(APP_NAME) $(HELM_CONTEXT_ARG) -n $(NAMESPACE) || true
	kubectl $(KUBECTL_CONTEXT_ARG) delete deployment,svc,ingress $(APP_NAME) -n $(NAMESPACE) || true

# ============================
# Cleanup
# ============================

.PHONY: clean
clean:
	@echo ">> Cleaning up"
	rm -f $(APP_NAME) $(APP_NAME)-linux-amd64 $(APP_NAME)-linux-arm64
