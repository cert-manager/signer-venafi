package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	capi "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const sampleCSR = `
-----BEGIN CERTIFICATE REQUEST-----
MIIBYjCCAQgCAQAwMDEuMCwGA1UEAxMlbXktcG9kLm15LW5hbWVzcGFjZS5wb2Qu
Y2x1c3Rlci5sb2NhbDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABByxdnpcoSYz
RYKeJjrXUSZS1RqNZB1XlnWzeiX053mbTKFEUwPCZWwnPWh0cEnDSRGE/z8fu6gY
5rr78NGYFAKgdjB0BgkqhkiG9w0BCQ4xZzBlMGMGA1UdEQRcMFqCJW15LXN2Yy5t
eS1uYW1lc3BhY2Uuc3ZjLmNsdXN0ZXIubG9jYWyCJW15LXBvZC5teS1uYW1lc3Bh
Y2UucG9kLmNsdXN0ZXIubG9jYWyHBMAAAhiHBAoAIgIwCgYIKoZIzj0EAwIDSAAw
RQIgcrXz7aqdftkuxz39PWtcx0J2JFLOD/xsch/YKFEQOXUCIQCUDPUzI+ncN1uN
2HujwAMigyFRDwFH9PIMKkd9t++MkA==
-----END CERTIFICATE REQUEST-----
`

var _ = Describe("CertificateSigningRequest Reconciler", func() {
	It("Signs a CSR with matching signerName", func() {
		ctx := context.Background()
		csr := &capi.CertificateSigningRequest{
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
		}
		Expect(k8sClient.Create(ctx, csr)).To(Succeed())
		defer func() {
			Expect(k8sClient.Delete(ctx, csr)).To(Succeed())
		}()

		key := client.ObjectKey{Namespace: csr.Namespace, Name: csr.Name}

		var actualCSR capi.CertificateSigningRequest
		Eventually(func() ([]byte, error) {
			err := k8sClient.Get(ctx, key, &actualCSR)
			return actualCSR.Status.Certificate, err
		}, 5).ShouldNot(BeNil())
	})
})
