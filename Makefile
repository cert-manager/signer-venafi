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

ARGS ?=

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
KUSTOMIZE := ${BIN}/kustomize-${KUSTOMIZE_VERSION}

# Kind
KIND_VERSION := 0.8.1
export KIND := ${BIN}/kind-${KIND_VERSION}

# Kube API Server
KUBE_APISERVER_VERSION := 1.18.2

# Kubebuilder
KUBEBUILDER_VERSION := 2.3.1
KUBEBUILDER_DOWNLOAD_URL := https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_${OS}_${ARCH}.tar.gz
KUBEBUILDER_BIN := ${BIN}/kubebuilder_${KUBEBUILDER_VERSION}_${OS}_${ARCH}/bin
KUBEBUILDER := ${KUBEBUILDER_BIN}/kubebuilder
export KUBEBUILDER_ASSETS := ${KUBEBUILDER_BIN}
export TEST_ASSET_KUBE_APISERVER := ${KUBEBUILDER_BIN}/kube-apiserver-${KUBE_APISERVER_VERSION}
export TEST_ASSET_ETCD := ${KUBEBUILDER_BIN}/etcd
export TEST_ASSET_KUBECTL := ${KUBEBUILDER_BIN}/kubectl
KUBEBUILDER_TEST_ASSETS := ${TEST_ASSET_KUBE_APISERVER} ${TEST_ASSET_ETCD} ${TEST_ASSET_KUBECTL}
export KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT := "true"

# Kubeadm
KUBEADM_VERSION := 1.19.0-rc.1
export KUBEADM ?= ${BIN}/kubeadm-${KUBEADM_VERSION}

export VCERT_INI := ${CURDIR}/vcert.ini

# Stop go build tools from silently modifying go.mod and go.sum
export GOFLAGS := -mod=readonly

# from https://suva.sh/posts/well-documented-makefiles/
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: verify
verify: ## Run all static checks
verify: verify-gomod verify-manifests verify-generate verify-fmt vet

# Run the supplied make target argument in a temporary workspace and diff the results.
verify-%: FORCE
	./hack/verify.sh ${MAKE} -s $*
FORCE:

.PHONY: test
test: ## Run tests
test: ${KUBEBUILDER_TEST_ASSETS}
	go test -v ./... -coverprofile cover.out
	go tool cover -func=cover.out

.PHONY: coverage_html
coverage_html: ## Run tests and open coverage report in web browser
coverage_html: test
	go tool cover -html=cover.out

.PHONY: manager
manager: ## Build manager binary
	go build -o bin/manager main.go

.PHONY: run
run: ## Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go ${ARGS}

.PHONY: deploy
deploy: ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: ${KUSTOMIZE}
	cd config/manager && ${KUSTOMIZE} edit set image controller=${DOCKER_IMAGE}
	${KUSTOMIZE} build config/default | kubectl apply -f -

.PHONY: deploy-example-signer
deploy-example-signer: ## Deploy a signer for example.com/foo
deploy-example-signer: ${KUSTOMIZE}
	cd config/manager && ${KUSTOMIZE} edit set image controller=${DOCKER_IMAGE}
	cp ${VCERT_INI} config/manager/vcert.ini
	${KUSTOMIZE} build docs/demos/example-signer | kubectl apply -f -

.PHONY: deploy-kubelet-signer
deploy-kubelet-signer: ## Deploy as a Kubelet CSR signer
deploy-kubelet-signer: ${KUSTOMIZE}
	cd config/manager && ${KUSTOMIZE} edit set image controller=${DOCKER_IMAGE}
	cp ${VCERT_INI} config/manager/vcert.ini
	${KUSTOMIZE} build docs/demos/kubelet-signer | kubectl apply -f -

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

.PHONY: gomod
gomod: ## Update the go.mod and go.sum files
	go mod tidy

.PHONY: go-get-patch
go-get-patch: ## Update Golang dependencies to latest patch versions
	go get -u=patch -t

.PHONY: docker-build
docker-build: ## Build the docker image
docker-build:
	docker build . -t ${DOCKER_IMAGE}

.PHONY: docker-push
docker-push: ## Push the docker image
docker-push:
	docker push ${DOCKER_IMAGE}

.PHONY: kind-create-cluster
kind-create-cluster: ## Create a Kind cluster for E2E testing
kind-create-cluster: ${KIND}
	${KIND} create cluster

.PHONY: kind-load
kind-load: ## Load the docker image into the Kind cluster
kind-load: ${KIND}
	${KIND} load docker-image ${DOCKER_IMAGE}

.PHONY: demo-kubelet-signer
demo-kubelet-signer: ## A demo showing how to set up a Kubernetes cluster with External CA and sign Kubelet client certificates
demo-kubelet-signer: ${KIND} ${KUBEADM} ${VCERT_INI}
	docs/demos/kubelet-signer/demo.sh

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

${KUSTOMIZE}: KUSTOMIZE_LOCAL_ARCHIVE=${BIN}/kustomize_v${KUSTOMIZE_VERSION}_${OS}_${ARCH}.tar.gz
${KUSTOMIZE}: | ${BIN}
	curl -sSL -o ${KUSTOMIZE_LOCAL_ARCHIVE} ${KUSTOMIZE_DOWNLOAD_URL}
	tar -C ${BIN} -x -f ${KUSTOMIZE_LOCAL_ARCHIVE}
	mv ${BIN}/kustomize ${KUSTOMIZE}

${KUBEBUILDER}: KUBEBUILDER_LOCAL_ARCHIVE=${BIN}/kubebuilder_v${KUBEBUILDER_VERSION}_${OS}_${ARCH}.tar.gz
${KUBEBUILDER}: | ${BIN}
	curl -sSL -o ${KUBEBUILDER_LOCAL_ARCHIVE} ${KUBEBUILDER_DOWNLOAD_URL}
	tar -C ${BIN} -x -f ${KUBEBUILDER_LOCAL_ARCHIVE}

${TEST_ASSET_KUBE_APISERVER}: | ${BIN}
	curl -sSL -o ${KUBEBUILDER_BIN}/kube-apiserver-${KUBE_APISERVER_VERSION} \
			https://storage.googleapis.com/kubernetes-release/release/v${KUBE_APISERVER_VERSION}/bin/${OS}/${ARCH}/kube-apiserver
	chmod +x ${TEST_ASSET_KUBE_APISERVER}

${KUBEBUILDER_TEST_ASSETS}: ${KUBEBUILDER}

${KIND}: | ${BIN}
	curl -sSL -o ${KIND} https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-${OS}-${ARCH}
	chmod +x ${KIND}

${KUBEADM}: | ${BIN}
	curl -sSL -o ${KUBEADM} \
			https://storage.googleapis.com/kubernetes-release/release/v${KUBEADM_VERSION}/bin/${OS}/${ARCH}/kubeadm
	chmod +x ${KUBEADM}
