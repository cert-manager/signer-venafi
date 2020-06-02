#!/usr/bin/env bash
#
# TODO: Documentation
#

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

OWD="${PWD}"
SCRIPT="${BASH_SOURCE[0]}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="${SCRIPT_DIR}"
KIND="${ROOT_DIR}/bin/kind-0.8.1"
KUBEADM="${ROOT_DIR}/kubeadm"
WORK_DIR="${OWD}/signer-venafi-kube-bootstrap"
KUBERNETES_DIR="${WORK_DIR}/etc_kubernetes"
CERTIFICATES_DIR="${KUBERNETES_DIR}/pki"
VCERT_INI="${ROOT_DIR}/vcert.ini"
VCERT_ZONE="TLS/SSL\Certificates\Kubernetes"

function log() {
    echo >&2
    echo "# $(date --rfc-3339=ns) :: ${*}" >&2
}

function kubeadm_init_phase_certs() {
    while read cert; do
        log "Generating CSR ${cert}"
        local args=""
        if [[ "${cert}" == "apiserver" ]]; then
            args="--apiserver-cert-extra-sans=kind-control-plane"
        fi
        ${KUBEADM} init phase certs "${cert}" ${args} --csr-only --csr-dir "${CERTIFICATES_DIR}" --cert-dir "${CERTIFICATES_DIR}" >/dev/null
        echo "${cert} ${CERTIFICATES_DIR}/${cert/#etcd-/etcd\/}.csr"
    done
}

function vcert_enroll() {
    while read nickname csr; do
        cert="${csr/%.csr/.crt}"
        policy="$(echo "${nickname}" | grep -o '\(client\|server\|peer\)$')"
        # Special case for etcd-server which needs both server and client usage
        # See https://clusterise.com/articles/kbp-2-certificates/ and
        # https://kubernetes.io/docs/setup/best-practices/certificates/#all-certificates
        if [[ "${nickname}" == "etcd-server" ]]; then
            policy=peer
        fi
        log "Enrolling CSR ${csr} with nickname ${nickname} to ${cert}"
        vcert enroll -z "${VCERT_ZONE}\\${policy}"  --config ${VCERT_INI} --nickname "${nickname}" --csr "file:${csr}" --cert-file "${cert}" >/dev/null
    done
}

if [[ -d "${WORK_DIR}" ]]; then
    log "Deleting previous work dir: ${WORK_DIR}"
    rm -rf "${WORK_DIR}"
fi

log "Working in ${WORK_DIR}"
mkdir -p "${WORK_DIR}"
pushd "${WORK_DIR}"

log "Creating certificates directory ${CERTIFICATES_DIR}"
mkdir -p "${CERTIFICATES_DIR}"

log "Generating certificate signing requests"
kubeadm_init_phase_certs <<EOF | vcert_enroll
  apiserver
  apiserver-etcd-client
  apiserver-kubelet-client
  etcd-healthcheck-client
  etcd-peer
  etcd-server
  front-proxy-client
EOF

log "Creating self-signed certificate authority"
${KUBEADM} init phase certs ca --cert-dir "${CERTIFICATES_DIR}" >/dev/null

log "Creating SA"
${KUBEADM} init phase certs sa --cert-dir "${CERTIFICATES_DIR}" >/dev/null

log "Installing Venafi CA cert"
# venafi cert must come first
cat "${ROOT_DIR}/ca.venafi.crt" "${CERTIFICATES_DIR}/ca.crt" > "${CERTIFICATES_DIR}/ca.crt.new"
mv "${CERTIFICATES_DIR}/ca.crt.new" "${CERTIFICATES_DIR}/ca.crt"
cp "${ROOT_DIR}/ca.venafi.crt" "${CERTIFICATES_DIR}/front-proxy-ca.crt"
cp "${ROOT_DIR}/ca.venafi.crt" "${CERTIFICATES_DIR}/etcd/ca.crt"

log "Starting Kind"
${KIND} create cluster --retain --config ${ROOT_DIR}/kind.conf.yaml
