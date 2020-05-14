package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	capi "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	It("reconciles", func() {
		csr := &capi.CertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test1-",
				Namespace:    "default",
			},
			Spec: capi.CertificateSigningRequestSpec{
				Request: []byte(sampleCSR),
				Usages: []capi.KeyUsage{
					"digital signature",
					"key encipherment",
					"server auth",
				},
			},
		}
		ctx := context.Background()
		// Create the Cluster object and expect the Reconcile and Deployment to be created
		Expect(k8sClient.Create(ctx, csr)).ToNot(HaveOccurred())
		// key := client.ObjectKey{Namespace: csr.Namespace, Name: csr.Name}
		defer func() {
			err := k8sClient.Delete(ctx, csr)
			Expect(err).NotTo(HaveOccurred())
		}()

		// Make sure the Cluster exists.
		// Eventually(func() bool {
		//	if err := testEnv.Get(ctx, key, instance); err != nil {
		//		return false
		//	}
		//	return len(instance.Finalizers) > 0
		// }, timeout).Should(BeTrue())
	})
})
