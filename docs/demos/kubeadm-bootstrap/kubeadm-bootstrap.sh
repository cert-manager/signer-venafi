#!/usr/bin/env bash
#
# TODO: Documentation
#

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export SCRIPT="${SCRIPT_DIR}/$(basename ${BASH_SOURCE[0]})"
export ROOT_DIR="$( cd "${SCRIPT_DIR}/../../.." && pwd )"
OWD="${PWD}"
WORK_DIR="${OWD}/demo-kubeadm-bootstrap"
KUBERNETES_DIR="${WORK_DIR}/etc_kubernetes"
CERTIFICATES_DIR="${KUBERNETES_DIR}/pki"
VCERT_ZONE="TLS/SSL\Certificates\Kubernetes\Cluster1"
VCERT_CA="${SCRIPT_DIR}/ca.venafi.crt"
KUBEADM_CONF="${SCRIPT_DIR}/kubeadm.conf"

: ${VCERT_INI:?}
: ${KIND:?}
: ${KUBEADM:?}
: ${STEP_DELAY:=0}

function log() {
    echo >&2
    echo "# $(date --rfc-3339=ns) :: ${*}" >&2
    sleep ${STEP_DELAY}
}

function vcert_enroll() {
    find ${KUBERNETES_DIR} -name '*.csr' | \
    while read csr; do
        nickname=$(basename $csr .csr)
        if echo $csr | fgrep '/etcd/' >/dev/null; then
            nickname="etcd-${nickname}"
        fi
        cert="${csr/%.csr/.crt}"
        policy="$(echo "${nickname}" | grep -o '\(client\|server\|peer\)$' || true)"
        # Special case for etcd-server which needs both server and client usage
        # See https://clusterise.com/articles/kbp-2-certificates/ and
        # https://kubernetes.io/docs/setup/best-practices/certificates/#all-certificates
        if [[ "${nickname}" == "etcd-server" ]]; then
            policy=peer
        fi
        if echo "${nickname}" | grep '\.conf$' > /dev/null; then
            policy=client
        fi
        log "Enrolling CSR ${csr} with nickname ${nickname} to ${cert} with policy ${policy}"
        vcert enroll \
              -z "${VCERT_ZONE}\\${policy}"  \
              --config ${VCERT_INI} \
              --nickname "${nickname}" \
              --csr "file:${csr}" \
              --pickup-id-file "${cert}.puid" \
              --no-pickup
        rm "${csr}"
    done
}

function vcert_pickup() {
    find ${KUBERNETES_DIR} -name '*.crt.puid' | \
        while read pickup_id_file; do
            cert="${pickup_id_file/%.crt.puid/.crt}"
            log "Collecting certificate ${cert}"
            vcert pickup \
              --config ${VCERT_INI} \
              --pickup-id-file "${pickup_id_file}" \
              --cert-file "${cert}"
            rm "${pickup_id_file}"
    done
}

function create_cluster() {
    pushd "${WORK_DIR}"
    log "Creating certificates directory ${CERTIFICATES_DIR}"
    mkdir -p "${CERTIFICATES_DIR}"

    log "Generating certificate signing requests"
    ${KUBEADM} alpha certs generate-csr \
               --config ${KUBEADM_CONF} \
               --cert-dir ${CERTIFICATES_DIR} \
               --kubeconfig-dir ${KUBERNETES_DIR}

    log "Sending CSRs to Venafi for signing"
    vcert_enroll

    log "Getting signed certificates from Venafi"
    vcert_pickup

    log "Setting Venafi CA in all kubeconfigs"
    venafi_ca_data=$(base64 -w 0 < "${SCRIPT_DIR}/ca.venafi.crt")
    find "${KUBERNETES_DIR}" -name '*.conf'  | \
        while read path; do
            kubectl --kubeconfig=${path} config set clusters.kubernetes.certificate-authority-data "${venafi_ca_data}"
            context=$(kubectl --kubeconfig "${path}" config current-context)
            kubectl --kubeconfig=${path} config set users.${context%%@kubernetes}.client-certificate-data "$(base64 -w 0 < "${path}.crt")"
        done

    log "Creating SA"
    ${KUBEADM} init phase certs sa --cert-dir "${CERTIFICATES_DIR}" >/dev/null

    log "Installing Venafi CA cert"
    cp "${VCERT_CA}" "${CERTIFICATES_DIR}/ca.crt"
    cp "${VCERT_CA}" "${CERTIFICATES_DIR}/front-proxy-ca.crt"
    cp "${VCERT_CA}" "${CERTIFICATES_DIR}/etcd/ca.crt"

    log "Creating Kind config"
    export KUBERNETES_DIR
    envsubst < ${SCRIPT_DIR}/kind.conf.yaml > kind.conf.yaml

    log "Starting Kind"
    ${KIND} create cluster --retain --config kind.conf.yaml

    log "Waiting for all nodes to be Ready"
    kubectl wait --timeout 5m --for condition=Ready node --all

    log "Cluster ready"
    kubectl get node

    sleep 10
    tmux kill-session
}

function start_operator() {
    pushd "${WORK_DIR}"
    watch --errexit --interval 1 "if test -e kind.conf.yaml; then exit 1; fi; tree ." < /dev/null || true

    log "Waiting for Kube config"
    until ${KIND} get kubeconfig > kube.config 2>/dev/null; do
        sleep 1
    done

    export KUBECONFIG="${PWD}/kube.config"

    log "Waiting for API server"
    until kubectl get nodes; do
        sleep 1
    done

    log "Starting signer-venafi"
    ${ROOT_DIR}/bin/manager \
               --signer-name=kubernetes.io/kube-apiserver-client-kubelet \
               --vcert-config=${ROOT_DIR}/vcert.ini
}


function main() {
    log "Working in ${WORK_DIR}"
    rm -rf "${WORK_DIR}"
    mkdir -p "${WORK_DIR}"

    tmux \
        new-session -d "${SCRIPT} create_cluster || sleep INFINITY" \; \
        split-window -d -h "${SCRIPT} start_operator || sleep INFINITY" \; \
        attach
    kind delete cluster
}

${1:-main}
