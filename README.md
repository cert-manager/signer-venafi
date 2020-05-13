# signer-venafi
Experimental Venafi based signer for Kubernetes 1.18 CSR API https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/20190607-certificates-api.md#signers


## Build and Test

```
kind create cluster
make docker-build kind-load deploy
kubectl -n signer-venafi-system logs deploy/signer-venafi-controller-manager manager --follow
```
