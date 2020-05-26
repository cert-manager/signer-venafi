package filter_test

import (
	"testing"

	capi "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/cert-manager/signer-venafi/internal/filter"
)

const sampleSignerName = "example.com/sample-signer-name"

// Sample Certificate request and certificate are generate according to instructions at:
// https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/master/docs/04-certificate-authority.md#provisioning-a-ca-and-generating-tls-certificates
const sampleCSR = `
-----BEGIN CERTIFICATE REQUEST-----
MIICwTCCAakCAQAwfDELMAkGA1UEBhMCVVMxDzANBgNVBAgTBk9yZWdvbjERMA8G
A1UEBxMIUG9ydGxhbmQxFzAVBgNVBAoTDnN5c3RlbTptYXN0ZXJzMSAwHgYDVQQL
ExdLdWJlcm5ldGVzIFRoZSBIYXJkIFdheTEOMAwGA1UEAxMFYWRtaW4wggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDJQ3WG50I8jq6UwiOe15NJNuPDTR53
Gb4qbH8xIese8gZABldAV98KOEd2JTkOuIpn59tfn3yraEDaG0fxbrQhZbpdwxFC
BM+p3Hpm9jmWsZHBc1n0Ieox8NATJ3tL28lkhWQDEN+K8qcqeyGcjM72KgVek+KB
n0ynofoUmqUKHnySwF2XlztIeiNywafQFQLWtaLDNmtHRHc9qnBE0NYStNvzkkaX
FqljmZb+9m9QrBY8s1MEV7rRmMD2294TvmbrwSkIEvmE7eobcnLKTdfXYb0KHq77
FRsujWxbjG96x69z3mZNuAn5XYHWvz+2GPewyxw7K6Tqroin9dOQxAStAgMBAAGg
ADANBgkqhkiG9w0BAQsFAAOCAQEAjsdH/IgtMTiF7zXAlVTZvrT4rxRLPJZ7E7m2
0PDHloRxK9nGywpmfXlLXDlJ2UL0i/Gipa01deujqhLwnq2LKuHfRn16fAMaHE9r
qioviGdEr/HLiXTZ087/cuLMu+CxyVrB5KvTptXVFAWcHlVjbUcFvmRnQPYPAVEX
WU54pq67c8CNy/b0JoCi/khmfbnalYvhYgQT9hhodkQeaq2/28LTtbJwXJ1mbQbC
kH/YwZEoKrJnLO0PWP0/emiNMxJYp1cPeQDsILMnJOjaR/WakCncGID3XbQO6LRw
OKbMbQNLoXS2f6qrS1Iqv4xxvHdDncH4zdhJiLdRqUJrSjPgMQ==
-----END CERTIFICATE REQUEST-----
`

func TestCSRFilter_Check(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*capi.CertificateSigningRequest)
		wantErr bool
	}{
		{
			name:    "Success",
			mutate:  func(csr *capi.CertificateSigningRequest) {},
			wantErr: false,
		},
		{
			name: "ErrorDeleted",
			mutate: func(csr *capi.CertificateSigningRequest) {
				t := metav1.Now()
				csr.SetDeletionTimestamp(&t)
			},
			wantErr: true,
		},
		{
			name: "ErrorNilSigner",
			mutate: func(csr *capi.CertificateSigningRequest) {
				csr.Spec.SignerName = nil
			},
			wantErr: true,
		},
		{
			name: "ErrorWrongSigner",
			mutate: func(csr *capi.CertificateSigningRequest) {
				*csr.Spec.SignerName += "XYZ"
			},
			wantErr: true,
		},
		{
			name: "ErrorNotApproved",
			mutate: func(csr *capi.CertificateSigningRequest) {
				csr.Status.Conditions = nil
			},
			wantErr: true,
		},
		{
			name: "ErrorAlreadySigned",
			mutate: func(csr *capi.CertificateSigningRequest) {
				csr.Status.Certificate = []byte{}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csr := capi.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
				},
				Spec: capi.CertificateSigningRequestSpec{
					SignerName: pointer.StringPtr(sampleSignerName),
					Request:    []byte(sampleCSR),
					Usages: []capi.KeyUsage{
						"digital signature",
						"key encipherment",
						"server auth",
					},
				},
				Status: capi.CertificateSigningRequestStatus{
					Conditions: []capi.CertificateSigningRequestCondition{
						{
							Type: capi.CertificateApproved,
						},
					},
				},
			}
			tt.mutate(&csr)
			o := &filter.CSRFilter{
				SignerName: sampleSignerName,
			}
			if err := o.Check(csr); (err != nil) != tt.wantErr {
				t.Errorf("CSRFilter.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
