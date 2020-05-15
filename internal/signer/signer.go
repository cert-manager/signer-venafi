package signer

import (
	capi "k8s.io/api/certificates/v1beta1"
)

type Signer interface {
	Sign(capi.CertificateSigningRequest) ([]byte, error)
}
