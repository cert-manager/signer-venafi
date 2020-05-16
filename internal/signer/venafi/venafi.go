package venafi

import (
	"fmt"
	"time"

	"github.com/Venafi/vcert/pkg/certificate"
	"github.com/Venafi/vcert/pkg/endpoint"
	"github.com/go-logr/logr"
	"github.com/jetstack/cert-manager/pkg/util/pki"
	capi "k8s.io/api/certificates/v1beta1"

	"github.com/cert-manager/signer-venafi/internal/signer"
)

type Signer struct {
	Client endpoint.Connector
	Log    logr.Logger
}

var _ signer.Signer = &Signer{}

func (o *Signer) Sign(csr capi.CertificateSigningRequest) ([]byte, error) {
	log := o.Log.WithName("Sign")

	log.V(1).Info("Generating template from CSR")
	tmpl, err := pki.GenerateTemplateFromCSRPEM(csr.Spec.Request, time.Hour*24, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate template from CSR PEM: %v", err)
	}

	log.V(1).Info("Generating vreq")
	vreq := certificate.NewRequest(tmpl)
	vreq.CsrOrigin = certificate.UserProvidedCSR
	vreq.CSR = csr.Spec.Request

	log.V(1).Info("Requesting certificate")
	pickupID, err := o.Client.RequestCertificate(vreq, "foo/bar")
	if err != nil {
		return nil, fmt.Errorf("failed to request certificate: %v", err)
	}

	log.V(1).Info("Retrieving certificate", "pickup-id", pickupID)
	certs, err := o.Client.RetrieveCertificate(&certificate.Request{PickupID: pickupID})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve certificate: %v", err)
	}

	return []byte(certs.Certificate), nil
}
