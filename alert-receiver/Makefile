# ============================
# Config (override as needed)
# ============================

APP_NAME       ?= alert-receiver
IMAGE_NAME     ?= alert-receiver
IMAGE_TAG      ?= local
FULL_IMAGE     := $(IMAGE_NAME):$(IMAGE_TAG)

K3D_CLUSTER    ?= k3d-local
REGISTRY       ?= localhost:5000
CHART          := ./charts/$(APP_NAME)
NAMESPACE      ?= alert-receiver

PORT                ?= 9094
PROMETHEUS_URL      ?= http://host.k3d.internal:9090
PROMETHEUS_LOOKBACK ?= 30m
PROMETHEUS_TIMEOUT  ?= 10s
LLM_TIMEOUT         ?= 30s
LLM_BACKENDS_JSON   ?= []

CGO_ENABLED    ?= 0

.PHONY: help
help:
	@echo ""
	@echo "alert-receiver Makefile"
	@echo ""
	@echo "Local development:"
	@echo "  make run                Run alert-receiver locally"
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
	@echo "  make rollout            Wait for deployment rollout"
	@echo "  make logs               Tail logs for deployed pods"
	@echo "  make describe           Describe deployed pods"
	@echo "  make delete             Uninstall Helm release and resources"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean"
	@echo ""

.PHONY: run
run:
	@echo ">> Running $(APP_NAME) locally"
	PORT="$(PORT)" \
	PROMETHEUS_URL="$(PROMETHEUS_URL)" \
	PROMETHEUS_LOOKBACK="$(PROMETHEUS_LOOKBACK)" \
	PROMETHEUS_TIMEOUT="$(PROMETHEUS_TIMEOUT)" \
	LLM_TIMEOUT="$(LLM_TIMEOUT)" \
	LLM_BACKENDS_JSON='$(LLM_BACKENDS_JSON)' \
	go run .

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

.PHONY: push-k3d
push-k3d: build-image
	@echo ">> Importing image into k3d cluster $(K3D_CLUSTER)"
	k3d image import $(FULL_IMAGE) -c $(K3D_CLUSTER)

.PHONY: push
push: build-image
	@echo ">> Tagging and pushing to registry $(REGISTRY)"
	docker tag $(FULL_IMAGE) $(REGISTRY)/$(APP_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(APP_NAME):$(IMAGE_TAG)

.PHONY: deploy
deploy: push
	@echo ">> Deploying $(APP_NAME) via Helm"
	helm upgrade --install $(APP_NAME) $(CHART) \
	  --namespace $(NAMESPACE) \
	  --set image.repository=k3d-edge-registry:5000/$(APP_NAME) \
	  --set image.tag=$(IMAGE_TAG)

.PHONY: rollout
rollout:
	@echo ">> Waiting for rollout of $(APP_NAME)"
	kubectl rollout status deployment/$(APP_NAME) -n $(NAMESPACE)

.PHONY: logs
logs:
	kubectl logs -l app=$(APP_NAME) -f -n $(NAMESPACE)

.PHONY: describe
describe:
	kubectl describe pod -l app=$(APP_NAME) -n $(NAMESPACE)

.PHONY: delete
delete:
	helm uninstall $(APP_NAME) -n $(NAMESPACE) || true
	kubectl delete deployment,svc $(APP_NAME) -n $(NAMESPACE) || true

.PHONY: clean
clean:
	@echo ">> Cleaning up"
	rm -f $(APP_NAME) $(APP_NAME)-linux-amd64 $(APP_NAME)-linux-arm64
