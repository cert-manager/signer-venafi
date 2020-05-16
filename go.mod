module github.com/cert-manager/signer-venafi

go 1.14

require (
	github.com/Venafi/vcert v3.18.4+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/jetstack/cert-manager v0.15.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/wallrj/vcert v3.18.4+incompatible
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	gopkg.in/ini.v1 v1.56.0 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/controller-runtime v0.6.0
)
