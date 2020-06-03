# Bootstrapping a Kubernetes Cluster With an External (Venafi) Certificate Authority

In this demo you will learn, step-by-step, how to create a Kubernetes cluster using [Kubeadm] in [External CA mode],
where the CA is managed by [Venafi Trust Protection Platform].
You will create a Kubernetes cluster using [Kind] and [Kubeadm],
where the Etcd server, Kubernetes API server serving certificates are signed by an external certificate authority (CA).

## Introduction

By default `kubeadm` will create a self-signed certificate authority (CA) comprising a private key and public certificate.
It then signs the serving and client certificates of all the Kubernetes components using that CA private key.
These signed certificates and keys are placed in the `/etc/kubernetes/pki` directory (by default)
and that directory is then read by the control-plane Kubelet
and mounted in to the static pods of the kube-apiserver, kube-controller-manager, and kube-scheduler.

Instead, if you pre-populate the `/etc/kubernetes/pki` directory with *just* the `ca.crt` certificate file
(not the `ca.key`) from an external CA provider, such as [Venafi Trust Protection Platform],
and pre-sign the other component certificates using that CA,
then `kubeadm` will operate in [External CA mode].
In [External CA mode], `kubeadm` will not attempt to generate any certificates;
it will only verify that all the certificates are present and that they are correctly configured and signed.

## Configuring Venafi Trust Protection Platform (TPP)

You will need certificates configured with specific sets of [Extended Key Usage] attributes.
Kubernetes requires some serving certificates and some client certificates.
Etcd peer servers require certificates that can be used for both serving and client authentication.

In TPP you will need to create three "certificate policy folders" linked to three different "CA templates", as follows:

* Policy
  * TLS / SSL
    * Administration
      * CA Templates
        * kubernetes-client - with extended usages: clientAuth
        * kubernetes-server - with extended usages: serverAuth
        * kubernetes-peer - with extended usages: clientAuth, serverAuth
    * Certificates
      * Kubernetes
        * cluster-1
          * client - linked to CA template: kubernetes-client
          * server - linked to CA template: kubernetes-server
          * peer - linked to CA template: kubernetes-peer

Log in to the TPP Web Admin page and create the policy folder structure above (or similar).

Additionally, log in to the TPP Aperture page and navigate to Configuration > Folders > Policy > TLS / SSL > Certificates > Kubernetes > Certificate Policy > Advanced Settings .
Set `SAN Types Allowed: DNS, IP`.
This allows certificates with both DNS names and IP addresses as Subject Alternative Names,
which is required for external access to the Kubernetes API server of a [Kind] cluster.

Download and install [Vcert], the Venafi CLI tool and create a `vcert.ini` file:

```ini
tpp_url = https://example.com/vedsdk
tpp_user = <tpp_username>
tpp_password = <tpp_password>
tpp_zone = TLS/SSL\For\Example
```
We will use this later to send certificates to TPP for signing.

## Generating Certificate Signing Requests

[Kubeadm] allows you to perform individual phases of its Kubernetes control-plane creation sequence.
Use the sub-command `kubeadm init --help` to see an overview of these phases and
use `kubeadm init phase --help` to see how to run a single phase.

Run sub-steps of the `certs` phase to generate certificate signing requests for each of the Kubernetes control-plane components,
which we will later submit to TPP using the `vcert` command.
For example:

```!sh
kubeadm init phase certs apiserver \
  --apiserver-cert-extra-sans=kind-control-plane,127.0.0.1 \
  --csr-only \
  --csr-dir "${CERTIFICATES_DIR}" \
  --cert-dir "${CERTIFICATES_DIR}"
```
This will create two files `apiserver.key` and `apiserver.csr`.

Here is a snippet of bash which can be used to create these CSR files for all the components:

```!sh
function kubeadm_init_phase_certs() {
    while read cert; do
        log "Generating CSR ${cert}"
        local args=""
        if [[ "${cert}" == "apiserver" ]]; then
            args="--apiserver-cert-extra-sans=kind-control-plane,127.0.0.1"
        fi
        ${KUBEADM} init phase certs "${cert}" ${args} --csr-only --csr-dir "${CERTIFICATES_DIR}" --cert-dir "${CERTIFICATES_DIR}" >/dev/null
        echo "${cert} ${CERTIFICATES_DIR}/${cert/#etcd-/etcd\/}.csr"
    done
}

kubeadm_init_phase_certs <<EOF
  apiserver
  apiserver-etcd-client
  apiserver-kubelet-client
  etcd-healthcheck-client
  etcd-peer
  etcd-server
  front-proxy-client
EOF
```

You can examine the CSR files using the `openssl` command, as follows:

```!shell
$ openssl req -noout -text -in demo-kubeadm-bootstrap//etc_kubernetes/pki/apiserver.csr
Certificate Request:
    Data:
        Version: 1 (0x0)
        Subject: CN = kube-apiserver
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                RSA Public-Key: (2048 bit)
                Modulus:
                    00:9f:09:81:6f:68:fa:7f:0d:dd:52:6e:91:e9:a9:
                    63:17:2c:cd:42:de:04:c4:1a:df:ac:fa:e9:f0:26:
                    b9:4a:73:59:83:72:42:2a:b3:fb:9c:f4:1c:06:27:
                    a8:92:fe:e1:2f:99:29:b3:c6:fe:01:61:0c:47:46:
                    33:09:84:63:d9:4c:20:29:69:84:c7:9c:2b:b8:9a:
                    00:ef:f3:ab:16:5f:a6:61:be:02:ec:e0:78:9f:29:
                    86:d1:97:35:9f:7c:7a:9e:77:40:97:8b:94:aa:02:
                    6b:46:d5:44:6b:ea:44:0d:50:4d:44:97:81:42:6c:
                    00:28:42:42:d8:86:cc:7c:3e:68:9e:1b:bd:72:99:
                    0c:0c:98:c6:06:fb:c6:dc:0a:de:12:95:81:af:aa:
                    ef:70:5c:1b:79:4c:6f:ec:53:7d:e4:57:c9:1a:99:
                    76:a7:00:46:85:84:df:f1:6f:b0:e4:50:23:cd:77:
                    45:e4:fa:30:3f:b2:fa:bf:41:46:35:eb:0b:cb:a3:
                    2e:d8:23:f1:6d:01:ef:19:80:c4:de:b0:fd:5a:60:
                    93:b0:73:1f:6a:a6:fc:43:3b:6c:18:61:f9:02:d2:
                    12:19:86:05:1a:8a:16:51:b3:43:14:76:dd:e7:97:
                    88:28:7a:69:52:f2:43:5d:e5:68:4c:60:cb:53:2a:
                    87:af
                Exponent: 65537 (0x10001)
        Attributes:
        Requested Extensions:
            X509v3 Subject Alternative Name:
                DNS:drax, DNS:kubernetes, DNS:kubernetes.default, DNS:kubernetes.default.svc, DNS:kubernetes.default.svc.cluster.local, DNS:kind-control-plane, IP Address:10.96.0.1, IP Address:192.168.0.11, IP Address:127.0.0.1
    Signature Algorithm: sha256WithRSAEncryption
         38:e7:0f:43:c6:7a:c6:60:a8:b4:d5:d7:a1:2d:46:cc:a8:f4:
         f5:13:70:7f:79:58:30:74:76:c8:df:a6:64:16:4c:db:9e:a1:
         8f:5f:c9:bd:79:3d:c4:6c:ba:a9:4a:28:44:1a:4f:63:b2:f8:
         44:54:be:64:3f:20:35:5f:ca:6c:93:77:fe:c2:f6:dd:d1:d5:
         82:43:b2:c6:c8:4e:95:38:a7:4a:a4:70:a8:2d:92:e6:f7:17:
         67:0f:78:55:39:65:44:8d:4d:8f:a3:fa:b0:7f:4a:77:6e:5d:
         a4:a1:5c:69:85:87:57:72:4a:85:5c:d2:28:82:80:e3:09:c6:
         84:f8:93:ea:60:6d:e7:6a:37:fe:4a:61:c4:58:0e:f5:f8:b2:
         a0:46:2a:fe:eb:dc:46:f2:8c:1e:f8:19:ca:b0:92:2b:af:d1:
         b5:3e:15:d6:37:ff:e3:26:6d:e7:80:52:f6:e0:a7:64:a0:e3:
         52:71:2d:28:34:1b:9c:a8:53:b9:bf:15:6f:10:a9:cd:04:6f:
         6b:f3:5a:e2:8e:7d:16:b9:cb:78:4b:bb:c7:02:48:dd:72:3d:
         04:f8:f2:4c:dc:14:b3:a7:5d:e8:61:61:91:5c:97:3b:7a:96:
         88:4a:43:f7:82:02:19:68:73:45:ed:79:94:4d:15:38:2b:6c:
         f7:cb:67:67
```

Notice the mix of DNS and IP SANs.

## Links

* [Kind](https://kind.sigs.k8s.io)
* [Kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm/)
* [Venafi Trust Protection Platform](https://www.venafi.com/platform/trust-protection-platform)
* [External CA mode](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#external-ca-mode)
* [Extended Key Usage](https://tools.ietf.org/html/rfc5280#section-4.2.1.12)
* [Vcert](https://github.com/Venafi/vcert)
