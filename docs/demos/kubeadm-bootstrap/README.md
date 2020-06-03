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

## Links

* [Kind](https://kind.sigs.k8s.io)
* [Kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm/)
* [Venafi Trust Protection Platform](https://www.venafi.com/platform/trust-protection-platform)
* [External CA mode](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#external-ca-mode)
* [Extended Key Usage](https://tools.ietf.org/html/rfc5280#section-4.2.1.12)
