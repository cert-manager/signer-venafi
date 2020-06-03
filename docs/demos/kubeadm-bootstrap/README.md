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

## Links

* [Kind](https://kind.sigs.k8s.io)
* [Kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm/)
* [Venafi Trust Protection Platform](https://www.venafi.com/platform/trust-protection-platform)
* [External CA mode](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#external-ca-mode)
