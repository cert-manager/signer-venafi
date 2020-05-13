MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
.SUFFIXES:
.DEFAULT_GOAL := help

# The semver version number which will be used as the Docker image tag
# Defaults to the output of git describe.
VERSION ?= $(shell git describe --tags --dirty)

# Docker image name parameters
DOCKER_PREFIX ?= quay.io/cert-manager/signer-venafi-
DOCKER_TAG ?= ${VERSION}
DOCKER_IMAGE ?= ${DOCKER_PREFIX}controller:${DOCKER_TAG}

OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

# BIN is the directory where build tools such controller-gen and kustomize will
# be installed.
# BIN is inherited and exported so that it gets passed down to the make process
# that is launched by verify.sh
# This ensures that test tools get installed in the original directory rather
# than in the temporary copy.
export BIN ?= ${CURDIR}/bin

# Make sure BIN is on the PATH
export PATH := $(BIN):$(PATH)

# controller-tools
CONTROLLER_GEN_VERSION := 0.3.0
CONTROLLER_GEN := ${BIN}/controller-gen-0.3.0

# Kustomize
KUSTOMIZE_VERSION := 3.5.5
KUSTOMIZE_DOWNLOAD_URL := https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_${OS}_${ARCH}.tar.gz
KUSTOMIZE_LOCAL_ARCHIVE := /tmp/kustomize_v${KUSTOMIZE_VERSION}_${OS}_${ARCH}.tar.gz
KUSTOMIZE := ${BIN}/kustomize-${KUSTOMIZE_VERSION}

# Kind
KIND_VERSION := 0.8.1
KIND := ${BIN}/kind-${KIND_VERSION}

# from https://suva.sh/posts/well-documented-makefiles/
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

.PHONY: manager
manager: ## Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

.PHONY: run
run: ## Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

.PHONY: deploy
deploy: ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: ${KUSTOMIZE}
	cd config/manager && ${KUSTOMIZE} edit set image controller=${DOCKER_IMAGE}
	${KUSTOMIZE} build config/default | kubectl apply -f -

.PHONY: manifests
manifests: ## Generate manifests e.g. CRD, RBAC etc.
manifests: ${CONTROLLER_GEN}
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths="./..." output:rbac:artifacts:config=config/rbac

.PHONY: fmt
fmt: ## Run go fmt against code
fmt:
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
vet:
	go vet ./...

.PHONY: generate
generate: ## Generate code
generate: ${CONTROLLER_GEN}
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: docker-build
docker-build: ## Build the docker image
docker-build:
	docker build . -t ${DOCKER_IMAGE}

.PHONY: docker-push
docker-push: ## Push the docker image
docker-push:
	docker push ${DOCKER_IMAGE}

.PHONY: kind-load
kind-load: ## Load the docker image into the Kind cluster
kind-load: ${KIND}
	${KIND} load docker-image ${DOCKER_IMAGE}

# ==================================
# Download: tools in ${BIN}
# ==================================
${BIN}:
	mkdir -p ${BIN}

${CONTROLLER_GEN}: | ${BIN}
# Prevents go get from modifying our go.mod file.
# See https://github.com/kubernetes-sigs/kubebuilder/issues/909
	cd /tmp; GOBIN=${BIN} GO111MODULE=on go get sigs.k8s.io/controller-tools/cmd/controller-gen@v${CONTROLLER_GEN_VERSION}
	mv ${BIN}/controller-gen ${CONTROLLER_GEN}

${KUSTOMIZE}: | ${BIN}
	curl -sSL -o ${KUSTOMIZE_LOCAL_ARCHIVE} ${KUSTOMIZE_DOWNLOAD_URL}
	tar -C ${BIN} -x -f ${KUSTOMIZE_LOCAL_ARCHIVE}
	mv ${BIN}/kustomize ${KUSTOMIZE}

${KIND}: ${BIN}
	curl -sSL -o ${KIND} https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-${OS}-${ARCH}
	chmod +x ${KIND}
