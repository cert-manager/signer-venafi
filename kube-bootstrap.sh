#!/usr/bin/env bash
#
# TODO: Documentation
#

set -o errexit
set -o nounset
set -o pipefail

OWD="${PWD}"
SCRIPT="${BASH_SOURCE[0]}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="${SCRIPT_DIR}"
KIND="${ROOT_DIR}/bin/kind-0.8.1"
KUBEADM="${ROOT_DIR}/kubeadm"
WORK_DIR="${OWD}/signer-venafi-kube-bootstrap"
CERTIFICATES_DIR="${WORK_DIR}/certificates"
VCERT_INI="${ROOT_DIR}/vcert.ini"

function log() {
    echo >&2
    echo "# $(date --rfc-3339=ns) :: ${*}" >&2
}

function kubeadm_init_phase_certs() {
    while read cert; do
        log "Generating CSR ${cert}"
        ${KUBEADM} init phase certs "${cert}" --csr-only --csr-dir "${CERTIFICATES_DIR}" --cert-dir "${CERTIFICATES_DIR}" >/dev/null
        echo "${cert} ${CERTIFICATES_DIR}/${cert/#etcd-/etcd\/}.csr"
    done
}

function vcert_enroll() {
    while read nickname csr; do
        log "Enrolling CSR ${csr} with nickname ${nickname}"
        vcert enroll --config ${VCERT_INI} --nickname "${nickname}" --csr "file:${csr}" --cert-file "${csr/%.csr/.crt}"
    done
}

if [[ -d "${WORK_DIR}" ]]; then
    log "Deleting previous work dir: ${WORK_DIR}"
    rm -rf "${WORK_DIR}"
fi

log "Working in ${WORK_DIR}"
mkdir -p "${WORK_DIR}"
pushd "${WORK_DIR}"

log "Creating certificates"
mkdir "${CERTIFICATES_DIR}"

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

log "Copying Venafi CA cert"
cp "${ROOT_DIR}/ca.venafi.crt" "${CERTIFICATES_DIR}/ca.crt"
