# Demo: Bootstrapping a Kubernetes cluster using Signer-Venafi

* Start a Kind cluster with csrsigning controller disabled.
* Launch the signer-venafi operator, outside the cluster,
  to sign the bootstrap CSR for the worker node.

The demo is recorded as follows:
```
asciinema rec --command 'make demo-kubelet-signer'  --title "Bootstrapping a Kubernetes cluster using Signer-Venafi"
```

## Requirements

1. A `<REPOSITORY>/vcert.ini` file with URL and credentials for a running TPP server.
2. TPP Policy folder linked to a Windows CA template with:
   1. Allowed usages: "signing", "key encipherment", "server auth", "client auth"
   2. Minimum key size: 256
