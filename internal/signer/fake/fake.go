package fake

import (
	capi "k8s.io/api/certificates/v1beta1"

	"github.com/cert-manager/signer-venafi/internal/signer"
)

type Signer struct {
	Certificate []byte
}

var _ signer.Signer = &Signer{}

func (o *Signer) Sign(csr capi.CertificateSigningRequest) ([]byte, error) {
	return o.Certificate, nil
}
