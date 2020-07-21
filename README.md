# signer-venafi
Experimental Venafi based signer for Kubernetes 1.18 CSR API https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/20190607-certificates-api.md#signers


## Demos

* [Example Signer](docs/demos/example-signer/README.md): demonstrates the simplest possible deployment, where the signer will sign CSRs having the signer name `example.com/foo`.
* [Bootstrapping a Kubernetes Cluster using Kubeadm and signer-venafi](docs/demos/kubelet-signer/README.md): demonstrates how to bootstrap a Kubernetes using "Kubeadm External CA Mode" to create the control-plane certificates and `signer-venafi` to sign the dynamically generated Kubelet certificates.

## Test

To run tests using in-memory fake Signer and fake vcert client.

```
make test
```

Or to run the Signer tests against a real Venafi TPP instance,
create a vcert.ini file and supply the path to that file as an environment variable for, as follows:

```
cat <<EOF > vcert.tpp.ini
tpp_url = https://tpp.example.com/vedsdk
tpp_user = <tppusername>
tpp_password = <tpppassword>
tpp_zone = TLS/SSL\Certificates\For\Example
EOF

VCERT_CONFIG_FILE=$PWD/vcert.tpp.ini make test
```
