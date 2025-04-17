# Optimized Makefile for Faster Local Dev

# Image URL for Docker builds
IMG ?= controller:latest

# Go bin path detection
ifeq (,$(shell go env GOBIN))
  GOBIN := $(shell go env GOPATH)/bin
else
  GOBIN := $(shell go env GOBIN)
endif

CONTAINER_TOOL ?= docker
SHELL := /usr/bin/env bash -o pipefail
.SHELLFLAGS := -ec

# Directory for downloaded tools
LOCALBIN ?= $(shell pwd)/bin
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

# Tool versions
KUSTOMIZE_VERSION ?= v5.6.0
CONTROLLER_TOOLS_VERSION ?= v0.17.2

# Binaries to download lazily
.PHONY: kustomize controller-gen
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

## Tool installation macro
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
	set -e; \
	package=$(2)@$(3) ;\
	echo "Downloading $${package}" ;\
	rm -f $(1) || true ;\
	GOBIN=$(LOCALBIN) go install $${package} ;\
	mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

##@ General
.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Fast Dev
.PHONY: dev
# Dev loop for local testing
# Skips CRD generation unless needed, just builds + runs
# Usage: make dev

dev: build-only run

.PHONY: dev-run
# Run only without any build or CRD generation
# Assumes code is already built and CRDs installed

dev-run:
	go run ./cmd/main.go

.PHONY: build-only
build-only: ## Fast build only (no manifests/gen)
	go build -o bin/manager cmd/main.go

.PHONY: run
run: ## Run operator locally
	go run ./cmd/main.go

.PHONY: restart
restart: clean uninstall install build-only run ## Clean, reinstall CRDs, rebuild, and run operator

.PHONY: reload
reload: uninstall install ## Re-apply only CRDs (no rebuild)

.PHONY: redeploy
redeploy: uninstall install deploy ## Uninstall and redeploy operator with CRDs

##@ Build + Install
.PHONY: build
build: manifests generate fmt vet ## Full build with manifests/gen
	go build -o bin/manager cmd/main.go

.PHONY: install
install: manifests kustomize ## Install CRDs into cluster
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

##@ Generate
.PHONY: manifests generate fmt vet
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt:
	go fmt ./...

vet:
	go vet ./...

##@ Clean
.PHONY: clean
clean:
	rm -rf dist *.out

.PHONY: uninstall
uninstall:
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f - --ignore-not-found=true
