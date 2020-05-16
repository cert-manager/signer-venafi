package venafi

import (
	"encoding/pem"
	"testing"

	"github.com/Venafi/vcert"
	"github.com/Venafi/vcert/pkg/endpoint"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	capi "k8s.io/api/certificates/v1beta1"
)

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

func TestSigner(t *testing.T) {
	vcertClient, err := vcert.NewClient(&vcert.Config{
		ConnectorType: endpoint.ConnectorTypeFake,
	})
	require.NoError(t, err)

	s := &Signer{Client: vcertClient, Log: zapr.NewLogger(zaptest.NewLogger(t)).WithName("Signer")}

	csr := capi.CertificateSigningRequest{
		Spec: capi.CertificateSigningRequestSpec{
			Request: []byte(sampleCSR),
		},
	}
	cert, err := s.Sign(csr)
	require.NoError(t, err)

	block, rest := pem.Decode(cert)
	assert.Empty(t, rest)
	assert.Equal(t, "CERTIFICATE", block.Type)
}
