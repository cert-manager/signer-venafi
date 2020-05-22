#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export TERM=dumb

OWD="${PWD}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$( cd "${SCRIPT_DIR}/../../.." && pwd )"

KIND="${ROOT_DIR}/bin/kind-0.8.1"

CWD="${OWD}/_demo"
rm -rf "${CWD}"
mkdir -p "${CWD}"
cd "${CWD}"

logger -s "Creating Kind cluster"
${KIND} create cluster --retain --config "${SCRIPT_DIR}/kind.conf.yaml" &
KIND_JOB="${!}"


logger -s "Getting Kube config"
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
           --vcert-config=${ROOT_DIR}/vcert.ini &
RUN_JOB="${!}"

logger -s "Waiting for kind create cluster to complete"
wait ${KIND_JOB}

logger -s "Stopping signer-venafi"
kill "${RUN_JOB}"
wait

logger -s "Waiting for all nodes to be ready"
kubectl wait --for condition=Ready node --all

logger -s "Cluster ready"
kubectl get node
