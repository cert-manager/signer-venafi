# Example Signer

Deploy `signer-venafi` as a signer for `CertificateSigningRequest` resources with `signer-name=example.com/foo`.
See `rbac.yaml` to see how we configure API server permissions to only allow the signer to sign CSR resources with that signer-name.

```
make kind-create-cluster docker-build kind-load deploy-example-signer
kubectl -n signer-venafi-system logs deploy/signer-venafi-controller-manager manager --follow

kubectl apply -f sample-csr.yaml
kubectl certificate approve sample-csr
```

```
kubectl get csr sample-csr
NAME         AGE   SIGNERNAME        REQUESTOR          CONDITION
sample-csr   36s   example.com/foo   kubernetes-admin   Approved,Issued
```
