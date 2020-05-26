#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export SCRIPT="${BASH_SOURCE[0]}"
export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export ROOT_DIR="$( cd "${SCRIPT_DIR}/../../.." && pwd )"
export KIND="${ROOT_DIR}/bin/kind-0.8.1"

function log() {
    echo
    echo "# $(date --rfc-3339=ns) :: ${*}"
}

function kind_create_cluster() {
    ${KIND} create cluster --retain --config "${SCRIPT_DIR}/kind.conf.yaml"
    sleep INFINITY
}

function start_operator() {
    log "Waiting for Kube config"
    until ${KIND} get kubeconfig > kube.config 2>/dev/null; do
        sleep 1
    done

    export KUBECONFIG="${PWD}/kube.config"

    log "Waiting for API server"
    until kubectl get nodes; do
        sleep 1
    done

    log "Appending Venafi CA Cert to Cluster CA"
    docker exec -i kind-control-plane bash -c 'cat >> /etc/kubernetes/pki/ca.crt' < ${ROOT_DIR}/ca.venafi.crt

    log "Restarting the Kube-apiserver"
    docker exec kind-control-plane bash -c 'kill $(pidof kube-apiserver)'

    log "Waiting for API server"
    until kubectl get nodes; do
        sleep 1
    done

    log "Starting signer-venafi"
    ${ROOT_DIR}/bin/manager \
               --signer-name=kubernetes.io/kube-apiserver-client-kubelet \
               --vcert-config=${ROOT_DIR}/vcert.ini
}

function wait_for_nodes() {
    log "Waiting for Kube config"
    until ${KIND} get kubeconfig > kube.config 2>/dev/null; do
        sleep 1
    done

    export KUBECONFIG="${PWD}/kube.config"

    log "Waiting for worker node"
    until kubectl get nodes kind-worker >/dev/null 2>&1; do
        sleep 1
    done

    log "Waiting for all nodes to be Ready"
    kubectl wait --timeout 5m --for condition=Ready node --all

    log "Cluster ready"
    kubectl get node

    sleep 10
    tmux kill-session
}

function main() {
    tmux \
        new-session -d "${SCRIPT} kind_create_cluster" \; \
        split-window -d -v "${SCRIPT} start_operator" \; \
        split-window -d -h "${SCRIPT} wait_for_nodes" \; \
        attach
    kind delete cluster
}

${1:-main}
