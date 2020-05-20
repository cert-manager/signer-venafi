package controllers

import (
	"context"
	"encoding/pem"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	capi "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

const sampleCertificate = `
-----BEGIN CERTIFICATE-----
MIID8TCCAtmgAwIBAgIUJA+ikzthgsVHR7XZhLZfLOIyR3EwDQYJKoZIhvcNAQEL
BQAwaDELMAkGA1UEBhMCVVMxDzANBgNVBAgTBk9yZWdvbjERMA8GA1UEBxMIUG9y
dGxhbmQxEzARBgNVBAoTCkt1YmVybmV0ZXMxCzAJBgNVBAsTAkNBMRMwEQYDVQQD
EwpLdWJlcm5ldGVzMB4XDTIwMDUxNTEyNTEwMFoXDTIxMDUxNTEyNTEwMFowfDEL
MAkGA1UEBhMCVVMxDzANBgNVBAgTBk9yZWdvbjERMA8GA1UEBxMIUG9ydGxhbmQx
FzAVBgNVBAoTDnN5c3RlbTptYXN0ZXJzMSAwHgYDVQQLExdLdWJlcm5ldGVzIFRo
ZSBIYXJkIFdheTEOMAwGA1UEAxMFYWRtaW4wggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQDJQ3WG50I8jq6UwiOe15NJNuPDTR53Gb4qbH8xIese8gZABldA
V98KOEd2JTkOuIpn59tfn3yraEDaG0fxbrQhZbpdwxFCBM+p3Hpm9jmWsZHBc1n0
Ieox8NATJ3tL28lkhWQDEN+K8qcqeyGcjM72KgVek+KBn0ynofoUmqUKHnySwF2X
lztIeiNywafQFQLWtaLDNmtHRHc9qnBE0NYStNvzkkaXFqljmZb+9m9QrBY8s1ME
V7rRmMD2294TvmbrwSkIEvmE7eobcnLKTdfXYb0KHq77FRsujWxbjG96x69z3mZN
uAn5XYHWvz+2GPewyxw7K6Tqroin9dOQxAStAgMBAAGjfzB9MA4GA1UdDwEB/wQE
AwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIw
ADAdBgNVHQ4EFgQUS+LL8ZkVySqfwActnjTVmWsOg7AwHwYDVR0jBBgwFoAU0ooN
d2A37YPf4WUu6cJDTTVf3OswDQYJKoZIhvcNAQELBQADggEBAF77Cd1ysnxmvQx3
3bMNsSPRbZ/KC/x/fteQujr8uTfeONLDofHiQBjw0UG5wEozw/WhbEjETmqPUqJA
ZcsTAm7TsQmKr+Rd6C1oBc98p8ptS5McTsxqk9a65bFrHABiznlbPIjrfFRBhP66
P56uwR0nC+PF0b3hbzgDhzZrUN9UoNAySuwLMfP6KZp3DSTC7yIEP70vFSTtmFOf
MOw+1TNpdMX2OokTLQsB+p4w8Eyjm9h0TMIv8yDzCSjipMd0Xp3MfgZ2Q2Ia54uG
gVTrBNGPua4NeGuQFvKDgpgBow6wTviIAm9SC7jo2kI670zaMuzvAraUej8tC/I5
BSvRUW8=
-----END CERTIFICATE-----`

var _ = Describe("CertificateSigningRequest Reconciler", func() {
	It("Signs a CSR with matching signerName", func() {
		By("Creating a sample CSR")
		ctx := context.Background()
		csr := &capi.CertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "default",
			},
			Spec: capi.CertificateSigningRequestSpec{
				// TODO: KubeBuilder 2.3.1 uses k8s 1.16 which does not have signerName
				// test fails because SignerName becomes nil after conversion
				SignerName: &sampleSignerName,
				Request:    []byte(sampleCSR),
				Usages: []capi.KeyUsage{
					"digital signature",
					"key encipherment",
					"server auth",
				},
			},
		}
		Expect(k8sClient.Create(ctx, csr)).To(Succeed())
		defer func() {
			By("Deleting the sample CSR")
			Expect(k8sClient.Delete(ctx, csr)).To(Succeed())
		}()

		By("Fetching the CSR back from the API server")
		key := client.ObjectKey{Namespace: csr.Namespace, Name: csr.Name}
		var actualCSR capi.CertificateSigningRequest
		Expect(k8sClient.Get(ctx, key, &actualCSR)).To(Succeed())

		time.Sleep(5)

		By("Approving the sample CSR")
		actualCSR.Status.Conditions = append(
			actualCSR.Status.Conditions,
			capi.CertificateSigningRequestCondition{
				Type:    capi.CertificateApproved,
				Reason:  "TestApprove",
				Message: "Approved for use in test",
			},
		)
		_, err := clientset.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(ctx, &actualCSR, metav1.UpdateOptions{})
		Expect(err).To(Succeed())

		By("Waiting for the CSR to be signed")
		Eventually(func() ([]byte, error) {
			err := k8sClient.Get(ctx, key, &actualCSR)
			return actualCSR.Status.Certificate, err
		}, 5).ShouldNot(BeNil())

		By("Checking that the CSR certificate content is a PEM encoded CERTIFICATE")
		block, rest := pem.Decode(actualCSR.Status.Certificate)
		Expect(block.Type).To(Equal("CERTIFICATEX"))
		Expect(rest).To(BeEmpty())
	})
})
