package fake

import (
	"github.com/cert-manager/signer-venafi/internal/signer"
	capi "k8s.io/api/certificates/v1beta1"
)

type Signer struct {
	Certificate []byte
}

var _ signer.Signer = &Signer{}

func (o *Signer) Sign(csr capi.CertificateSigningRequest) ([]byte, error) {
	return o.Certificate, nil
}