# signer-venafi
Experimental Venafi based signer for Kubernetes 1.18 CSR API https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/20190607-certificates-api.md#signers


## Build and Deploy

```
kind create cluster
make docker-build kind-load deploy
kubectl -n signer-venafi-system logs deploy/signer-venafi-controller-manager manager --follow
```


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
