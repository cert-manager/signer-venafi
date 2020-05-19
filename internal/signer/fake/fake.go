package fake

import (
	capi "k8s.io/api/certificates/v1beta1"

	"github.com/cert-manager/signer-venafi/internal/signer"
)

const pickupID = "foo-bar"

type Signer struct {
	Certificate []byte
}

var _ signer.Signer = &Signer{}

func (o *Signer) Sign(csr capi.CertificateSigningRequest) (string, error) {
	return pickupID, nil
}

func (o *Signer) Pickup(pickupID string) ([]byte, error) {
	return o.Certificate, nil
}
