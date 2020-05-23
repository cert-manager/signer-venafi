#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export SCRIPT="${BASH_SOURCE[0]}"
export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export ROOT_DIR="$( cd "${SCRIPT_DIR}/../../.." && pwd )"
export KIND="${ROOT_DIR}/bin/kind-0.8.1"

function kind_create_cluster() {
    ${KIND} create cluster --retain --config "${SCRIPT_DIR}/kind.conf.yaml"
    sleep INFINITY
}

function start_operator() {
    logger -s "Waiting for Kube config"
    until ${KIND} get kubeconfig > kube.config 2>/dev/null; do
        sleep 1
    done

    export KUBECONFIG="${PWD}/kube.config"

    logger -s "Waiting for API server"
    until kubectl get nodes; do
        sleep 1
    done

    logger -s "Appending Venafi CA Cert to Cluster CA"
    docker exec -i kind-control-plane bash -c 'cat >> /etc/kubernetes/pki/ca.crt' < ${ROOT_DIR}/ca.venafi.crt

    logger -s "Restarting the Kube-apiserver"
    docker exec kind-control-plane bash -c 'kill $(pidof kube-apiserver)'

    logger -s "Waiting for API server"
    until kubectl get nodes; do
        sleep 1
    done

    logger -s "Starting signer-venafi"
    ${ROOT_DIR}/bin/manager \
               --signer-name=kubernetes.io/kube-apiserver-client-kubelet \
               --vcert-config=${ROOT_DIR}/vcert.ini
}

function wait_for_nodes() {
    logger -s "Waiting for Kube config"
    until ${KIND} get kubeconfig > kube.config 2>/dev/null; do
        sleep 1
    done

    export KUBECONFIG="${PWD}/kube.config"

    logger -s "Waiting for worker node"
    until kubectl get nodes kind-worker >/dev/null 2>&1; do
        sleep 1
    done

    logger -s "Waiting for all nodes to be Ready"
    kubectl wait --timeout 5m --for condition=Ready node --all

    logger -s "Cluster ready"
    kubectl get node

    sleep 10
    tmux kill-session
}

function main() {
    tmux \
        new-session -d "${SCRIPT} kind_create_cluster" \; \
        split-window -d "${SCRIPT} wait_for_nodes" \; \
        split-window -d "${SCRIPT} start_operator" \; \
        select-layout even-vertical \; \
        attach
    kind delete cluster
}

${1:-main}
