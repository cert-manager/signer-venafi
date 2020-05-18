package api

import (
	capi "k8s.io/api/certificates/v1beta1"
)

// IsCertificateRequestApproved returns true if a certificate request has the
// "Approved" condition and no "Denied" conditions; false otherwise.
func IsCertificateRequestApproved(csr *capi.CertificateSigningRequest) bool {
	approved, denied := GetCertApprovalCondition(&csr.Status)
	return approved && !denied
}

func GetCertApprovalCondition(status *capi.CertificateSigningRequestStatus) (approved bool, denied bool) {
	for _, c := range status.Conditions {
		if c.Type == capi.CertificateApproved {
			approved = true
		}
		if c.Type == capi.CertificateDenied {
			denied = true
		}
	}
	return
}
