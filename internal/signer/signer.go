package signer

import (
	"errors"

	capi "k8s.io/api/certificates/v1beta1"
)

// ErrTemporary should be wrapped by any temporary errors returned by
// implementations of Signer so that the caller can know whether to retry.
var ErrTemporary = errors.New("Temporary Error")

type Signer interface {
	// Sign makes a request to process a certificate signing request and returns
	// a pickup ID which can be used later in Pickup
	Sign(csr capi.CertificateSigningRequest) (pickupID string, err error)
	// Pickup retrieves the signed certificate data corresponding to the
	// supplied pickup ID.
	// May return an error wrapping ErrTemporary, in which case the called
	// should retry the Pickup with the same pickupID.
	Pickup(pickupID string) (certificate []byte, err error)
}
