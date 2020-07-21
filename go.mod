module github.com/cert-manager/signer-venafi

go 1.14

require (
	github.com/Venafi/vcert v3.18.4+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/jetstack/cert-manager v0.15.2
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.10.0
	gopkg.in/ini.v1 v1.56.0 // indirect
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6 // indirect
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29 // indirect
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	sigs.k8s.io/controller-runtime v0.6.1
)
