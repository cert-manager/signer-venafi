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

In the Microsoft Windows certificate template management console, you will need to create three new CA templates:

* kubernetes-client - with extended usages: clientAuth
* kubernetes-server - with extended usages: serverAuth
* kubernetes-peer - with extended usages: clientAuth, serverAuth


In the Microsoft Windows certificate authority management console, you will need to add those three new templates to an existing CA certificate.

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
tpp_zone = TLS/SSL\Certificates\Kubernetes\Cluster1
```
We will use this later to send certificates to TPP for signing.

Export the Venafi TPP CA certificate as a PEM encoded text file,
using the Windows Management Console on the server hosting TPP,
and save it as `ca.venafi.crt`.

## Demo Script

The steps below have been wrapped in a script.
See `./demo.sh`.

[![asciicast](https://asciinema.org/a/jJhINutJiEr8xdOCnebzLyTTO.svg)](https://asciinema.org/a/jJhINutJiEr8xdOCnebzLyTTO)

You can run the script by running `make demo-kubelet-signer` from the root of this repository.

## Generating Certificate Signing Requests

Run `kubeadm alpha certs generate-csr` to generate private keys and CSR files for all the certificates required by the Kubernetes control-plane.
For example:

```!sh
kubeadm alpha certs generate-csr \
           --config kubeadm.conf \
           --kubeconfig-dir etc_kubernetes/pki \
           --cert-dir etc_kubernetes/pki
```

Where `kubeadm.conf` contains:

```
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
apiServer:
  certSANs:
  - "kind-control-plane"
  - "127.0.0.1"
controlPlaneEndpoint: "kind-control-plane"
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
nodeRegistration:
  name: "kind-control-plane"
```

This will create the following directories and files:

```
etc_kubernetes/
└── pki
    ├── admin.conf
    ├── admin.conf.csr
    ├── apiserver.csr
    ├── apiserver-etcd-client.csr
    ├── apiserver-etcd-client.key
    ├── apiserver.key
    ├── apiserver-kubelet-client.csr
    ├── apiserver-kubelet-client.key
    ├── controller-manager.conf
    ├── controller-manager.conf.csr
    ├── etcd
    │   ├── healthcheck-client.csr
    │   ├── healthcheck-client.key
    │   ├── peer.csr
    │   ├── peer.key
    │   ├── server.csr
    │   └── server.key
    ├── front-proxy-client.csr
    ├── front-proxy-client.key
    ├── kubelet.conf
    ├── kubelet.conf.csr
    ├── scheduler.conf
    └── scheduler.conf.csr
```

You can examine the CSR files using the `openssl` command, as follows:

```!shell
$ openssl req -noout -text -in etc_kubernetes/pki/apiserver.csr
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

## Signing Certificates with Venafi TPP

Now you need to send all those CSR files to your Venafi TPP server to be signed.
Each CSR needs to be submitted to the TPP policy sub-folder corresponding to that certificate's required [Extended Key Usage] attributes.
You can do this using the `vcert enroll` sub-command, as follows:

```!sh
vcert enroll -z "${VCERT_ZONE}\server" --config ${VCERT_INI} --nickname apiserver --csr file:kubernetes/pki/apiserver.csr --cert-file kubernetes/pki/apiserver.crt
```

This will submit the CSR file to TPP and poll until the signed certificate is downloaded.

Here is a snippet of bash which can be used to sign all the certificates:

```!sh
function vcert_enroll() {
    while read nickname csr; do
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
        vcert enroll -z "${VCERT_ZONE}\\${policy}"  --config ${VCERT_INI} --nickname "${nickname}" --csr "file:${csr}" --cert-file "${cert}" >/dev/null
    done
}

find ${KUBERNETES_DIR} -name '*.csr' | \
    while read path; do
        nickname=$(basename $path .csr)
        if echo $path | fgrep '/etcd/' >/dev/null; then
            nickname="etcd-${nickname}"
        fi
        echo $nickname $path
    done | \
        vcert_enroll
```

You can examine the signed certificates using `openssl`, as follows:

```!sh
$ openssl x509 -noout -text -in etc_kubernetes/pki/apiserver.crt
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            2f:00:00:02:27:82:75:58:d2:40:d1:94:c3:00:00:00:00:02:27
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: DC = com, DC = venafidemo, CN = venafidemo-TPP-CA
        Validity
            Not Before: Jun  3 10:10:21 2020 GMT
            Not After : Jun  3 10:10:21 2021 GMT
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
        X509v3 extensions:
            X509v3 Subject Alternative Name:
                DNS:drax, DNS:kubernetes, DNS:kubernetes.default, DNS:kubernetes.default.svc, DNS:kubernetes.default.svc.cluster.local, DNS:kind-control-plane, DNS:kube-apiserver, IP Address:10.96.0.1, IP Address:192.168.0.11, IP Addr
ess:127.0.0.1
            X509v3 Subject Key Identifier:
                D8:99:F5:2B:D3:7A:41:E9:EC:5C:CA:27:41:F1:C6:74:E0:85:94:F3
            X509v3 Authority Key Identifier:
                keyid:83:75:7A:54:58:18:B8:22:1D:28:77:BE:ED:E5:29:3F:D8:A1:F5:FE

            X509v3 CRL Distribution Points:

                Full Name:
                  URI:ldap:///CN=venafidemo-TPP-CA,CN=tpp,CN=CDP,CN=Public%20Key%20Services,CN=Services,CN=Configuration,DC=venafidemo,DC=com?certificateRevocationList?base?objectClass=cRLDistributionPoint

            Authority Information Access:
                CA Issuers - URI:ldap:///CN=venafidemo-TPP-CA,CN=AIA,CN=Public%20Key%20Services,CN=Services,CN=Configuration,DC=venafidemo,DC=com?cACertificate?base?objectClass=certificationAuthority

            X509v3 Key Usage: critical
                Digital Signature, Key Encipherment
            1.3.6.1.4.1.311.21.7:
                0/.'+.....7.........l...<...:...S.f...6......d...
            X509v3 Extended Key Usage:
                TLS Web Server Authentication
            1.3.6.1.4.1.311.21.10:
                0.0
..+.......
    Signature Algorithm: sha256WithRSAEncryption
         00:a0:64:99:93:c2:7d:cf:a0:c8:ae:2b:6b:66:64:a8:3a:c4:
         6c:75:43:5b:27:67:de:42:94:ed:cd:8a:13:02:e7:43:65:21:
         77:30:e3:d1:ce:df:97:0a:f4:3e:03:31:6f:35:50:23:28:04:
         3a:93:f8:cd:c7:59:b5:77:56:75:50:87:82:8e:60:6b:75:f1:
         cc:e2:72:fc:3c:7d:29:ee:93:d4:a9:c6:a4:cd:62:b7:66:5d:
         44:09:63:97:3d:46:5a:7d:f5:63:c2:e4:d0:e4:f7:b8:db:9d:
         70:e0:8a:94:13:d5:59:1c:c3:bd:0c:b3:9e:e1:a7:99:65:9f:
         13:71:df:78:f2:e7:1d:c6:81:ef:ef:f5:af:99:fd:57:a9:e4:
         a9:ac:8f:6f:76:a4:1f:8a:d1:7c:21:49:fa:6b:c1:12:84:3f:
         97:1b:27:34:d2:1e:3b:71:36:03:76:53:e3:ac:98:f8:14:81:
         96:80:3b:de:77:d7:37:32:f6:d5:4f:52:b2:8c:a1:72:f8:fd:
         fb:96:cc:c9:95:4d:f6:bc:6f:53:6c:75:cc:f7:20:98:71:71:
         f2:0b:50:40:f3:5a:e8:1d:a8:99:37:7d:0f:df:d5:64:3f:0e:
         5c:aa:53:d4:a7:b4:71:2a:22:6e:79:4a:c2:7d:1a:2c:0b:2e:
         3a:97:c0:74

```

Notice the "Issuer" details at the top, showing that the certificate is signed by the Venafi CA.
And notice the [Extended Key Usage] details at the end, showing that this certificate is for server authentication only.

## Wrapping Up

### Create the CA certificate files

Copy the `ca.venafi.crt` (which you downloaded from TPP earlier) file to `kubernetes/pki/ca.crt`
Also save copies of the that file as: `kubernetes/pki/front-proxy-ca.crt` and `kubernetes/pki/etcd/ca.crt`.
This is because we are using the same CA for signing the `Etcd` and [Aggregated API server] certificates.

**NOTE: We use a single CA in this demo for simplicity. It is not recommended to use the same CA for all three.**

### Kubeconfig Certificates

The `kubelet`, `kube-controller-manager` and `kube-scheduler` all need `kubeconfig` files with embedded client certificates,
which allow them to connect to the `kube-apiserver`.

These `kubeconfig` files will have been generated by `kubeadm alpha certs generate-csr` above, but you now need to update those files with the 

Modify the CA data in each of the KUBECONFIG files:

```
venafi_ca_data=$(base64 -w 0 < "kubernetes/pki/ca.crt")
find "kubernetes/" -name '*.conf'  | \
    xargs -n 1 -I {} -- \
          kubectl --kubeconfig={} config set clusters.kubernetes.certificate-authority-data "${venafi_ca_data}"

```

Add the signed client certificate data to each file:

```
find "${KUBERNETES_DIR}" -name '*.conf'  | \
    while read path; do
        context=$(kubectl --kubeconfig "${path}" config current-context)
        kubectl --kubeconfig=${path} config set users.${context%%@kubernetes}.client-certificate-data "$(base64 -w 0 < "${path}.crt")"
    done
```

Note that the CA and certificate PEM data is base64 encoded (without line-breaks).

### Create a Service Account Token Encryption Key Pair

We also have to pre-create the `sa.key` and `sa.pub` key pair, which is used by the API server to encrypt authentication tokens used by intra-cluster clients, via Service Accounts.

```
kubeadm init phase certs sa --cert-dir "kubernetes/pki"
```

### Configure Kind to Mount the Pre-generated Certificates and KUBECONFIG files

We are almost ready to start the Kind cluster, but first we need to create a Kind configuration file.
This configuration will cause Kind to mount our pre-generated certificates and KUBECONFIG files into the Kind control-plane Docker container at the default Kubernetes directory paths.
Create a file called `kind.conf.yaml` with the following content:

```
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.18.2@sha256:7b27a6d0f2517ff88ba444025beae41491b016bc6af573ba467b70c5e8e0d85f
  extraMounts:
    - containerPath: /etc/kubernetes/pki
      hostPath: ${KUBERNETES_DIR}/pki
      readOnly: true
    - containerPath: /etc/kubernetes/admin.conf
      hostPath: ${KUBERNETES_DIR}/admin.conf
      readOnly: true
    - containerPath: /etc/kubernetes/controller-manager.conf
      hostPath: ${KUBERNETES_DIR}/controller-manager.conf
      readOnly: true
    - containerPath: /etc/kubernetes/kubelet.conf
      hostPath: ${KUBERNETES_DIR}/kubelet.conf
      readOnly: false
    - containerPath: /etc/kubernetes/scheduler.conf
      hostPath: ${KUBERNETES_DIR}/scheduler.conf
      readOnly: true
- role: worker
  image: kindest/node:v1.18.2@sha256:7b27a6d0f2517ff88ba444025beae41491b016bc6af573ba467b70c5e8e0d85f
  extraMounts:
    - containerPath: /etc/kubernetes/pki/ca.crt
      hostPath: ${KUBERNETES_DIR}/pki/ca.crt
      readOnly: true
```

Where `${KUBERNETES_DIR}` is absolute the path to `kubernetes/` which contains all the KUBECONFIG files and the `pki/` sub-directory.

Notice that these are all mounted `readonly` except `kubelet.conf`, which needs to be mutable because one of the `kubeadm init` steps that will be performed inside the Docker container,
is to reconfigure the Kubelet to use a dynamically generated KUBECONFIG at `/var/lib/kubernetes` where its embedded certificate can be regularly rotated.


## Start the Cluster

Finally, with all the certificates and KUBECONFIG files ready, we can start the cluster using our custom config file:

```
kind create cluster --retain --config kind.conf.yaml
```

Notice that we use the `--retain` flag, so that if Kind fails it will leave behind the Docker containers so that we can investigate the problem.

## Deploy signer-venafi on the control plane node

Kind will first start the control-plane node using keys and certificates provisioned earlier.
The worker node, uses dynamically generated keys and certificates.
Deploy the `signer-venafi` controller to the control-plane node:

```
until ${KIND} get kubeconfig > kube.config 2>/dev/null; do
    sleep 1
done

export KUBECONFIG="${PWD}/kube.config"

until kubectl get nodes; do
    sleep 1
done

make -C ${ROOT_DIR} docker-build kind-load deploy-kubelet-signer
kubectl -n signer-venafi-system rollout status deployment signer-venafi-controller-manager
kubectl -n signer-venafi-system logs --follow deploy/signer-venafi-controller-manager manager
```

The signer will watch the API server for `CertificateSigningRequest` resources for the kubelet and sign them using the Venafi API.
The worker node Kubelet should then be able to join the cluster and the `kind` command will finish.

## Discussion

### kubeadm should have a pluggable automated signing mechanism

To simplify this whole process of using `kubeadm init` with an external CA.

### TPP should have an API for dynamically configuring policy folders

It is difficult to set up TPP with all the necessary CA templates and policy folders, but as far as I can see it is not possible to automate the policy creation with the TPP API.

### Venafi Cloud should support ECDSA certificates

All the Kubernetes tools are hardcoded to create ECDSA keys and certificates, but these are not supported by Venafi Cloud.


## Links

* [Kind](https://kind.sigs.k8s.io)
* [Kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm/)
* [Venafi Trust Protection Platform](https://www.venafi.com/platform/trust-protection-platform)
* [External CA mode](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#external-ca-mode)
* [Extended Key Usage](https://tools.ietf.org/html/rfc5280#section-4.2.1.12)
* [Vcert](https://github.com/Venafi/vcert)
