package venafi

import (
	"fmt"
	"strings"
	"time"

	"github.com/Venafi/vcert/pkg/certificate"
	"github.com/Venafi/vcert/pkg/endpoint"
	"github.com/go-logr/logr"
	"github.com/jetstack/cert-manager/pkg/util/pki"
	capi "k8s.io/api/certificates/v1beta1"

	"github.com/cert-manager/signer-venafi/internal/signer"
)

type Signer struct {
	ClientFactory func() (endpoint.Connector, error)
	Log           logr.Logger
}

var _ signer.Signer = &Signer{}

func (o *Signer) Sign(csr capi.CertificateSigningRequest) (string, error) {
	log := o.Log.WithName("Sign")

	log.V(1).Info("Generating template from CSR")
	tmpl, err := pki.GenerateTemplateFromCSRPEM(csr.Spec.Request, time.Hour*24, false)
	if err != nil {
		return "", fmt.Errorf("failed to generate template from CSR PEM: %v", err)
	}

	log.V(1).Info("Generating vreq")
	vreq := certificate.NewRequest(tmpl)
	vreq.CsrOrigin = certificate.UserProvidedCSR
	vreq.CSR = csr.Spec.Request

	log.V(1).Info("Requesting certificate")
	client, err := o.ClientFactory()
	if err != nil {
		return "", fmt.Errorf("failed to initialise vcert client: %s", err)
	}

	pickupID, err := client.RequestCertificate(vreq, "")
	if err != nil {
		return "", fmt.Errorf("failed to request certificate: %v", err)
	}
	return pickupID, nil
}

func (o *Signer) Pickup(pickupID string) ([]byte, error) {
	log := o.Log.WithName("Pickup")

	log.V(1).Info("Retrieving certificate", "pickup-id", pickupID)
	client, err := o.ClientFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialise vcert client: %s", err)
	}
	certs, err := client.RetrieveCertificate(&certificate.Request{PickupID: pickupID})
	if err != nil {
		if strings.Contains(err.Error(), "Issuance is pending.") {
			return nil, fmt.Errorf("%w: certificate not ready: %s", signer.ErrTemporary, err)
		}
		return nil, fmt.Errorf("failed to retrieve certificate: %v", err)
	}
	return []byte(certs.Certificate), nil
}
