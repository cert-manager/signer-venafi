package filter

import (
	"fmt"

	capi "k8s.io/api/certificates/v1beta1"

	capihelper "github.com/cert-manager/signer-venafi/internal/api"
)

// Filter allows unsuitable CSR resources to be filtered out before we attempt
// to process and sign them.
// This is abstracted so that the filtering logic can be unit-tested separately
// from the Reconcile function.
type Filter interface {
	// Check returns an error if the CSR is unsuitable for processing.
	// E.g. if it is marked for deletion or if it has not been approved.
	Check(capi.CertificateSigningRequest) error
}

type CSRFilter struct {
	SignerName string
}

var _ Filter = &CSRFilter{}

func (o *CSRFilter) Check(csr capi.CertificateSigningRequest) error {
	switch {
	case !csr.DeletionTimestamp.IsZero():
		return fmt.Errorf("CSR has been deleted")
	case csr.Spec.SignerName == nil:
		return fmt.Errorf("CSR does not have a signer name")
	case *csr.Spec.SignerName != o.SignerName:
		return fmt.Errorf("CSR signer name does not match")
	case !capihelper.IsCertificateRequestApproved(&csr):
		return fmt.Errorf("CSR is not approved")
	case csr.Status.Certificate != nil:
		return fmt.Errorf("CSR has already been signed")
	}
	return nil
}
