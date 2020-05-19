package signer

import (
	"errors"

	capi "k8s.io/api/certificates/v1beta1"
)

var ErrTemporary = errors.New("Temporary Error")

type Signer interface {
	Sign(capi.CertificateSigningRequest) (string, error)
	Pickup(string) ([]byte, error)
}
