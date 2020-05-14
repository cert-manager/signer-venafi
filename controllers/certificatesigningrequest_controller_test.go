package controllers

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	capi "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestXYZ(t *testing.T) {
	t.Log("Hello world")
	if 1 == 1 {
		t.Errorf("1 != 1")
	}
}

// const sampleCSR =

var _ = Describe("CertificateSigningRequest Reconciler", func() {
	It("reconciles", func() {
		csr := &capi.CertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test1-",
				Namespace:    "default",
			},
			Spec: capi.CertificateSigningRequestSpec{
				Request: []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJQllqQ0NBUWdDQVFBd01ERXVNQ3dHQTFVRUF4TWxiWGt0Y0c5a0xtMTVMVzVoYldWemNHRmpaUzV3YjJRdQpZMngxYzNSbGNpNXNiMk5oYkRCWk1CTUdCeXFHU000OUFnRUdDQ3FHU000OUF3RUhBMElBQkJ5eGRucGNvU1l6ClJZS2VKanJYVVNaUzFScU5aQjFYbG5XemVpWDA1M21iVEtGRVV3UENaV3duUFdoMGNFbkRTUkdFL3o4ZnU2Z1kKNXJyNzhOR1lGQUtnZGpCMEJna3Foa2lHOXcwQkNRNHhaekJsTUdNR0ExVWRFUVJjTUZxQ0pXMTVMWE4yWXk1dAplUzF1WVcxbGMzQmhZMlV1YzNaakxtTnNkWE4wWlhJdWJHOWpZV3lDSlcxNUxYQnZaQzV0ZVMxdVlXMWxjM0JoClkyVXVjRzlrTG1Oc2RYTjBaWEl1Ykc5allXeUhCTUFBQWhpSEJBb0FJZ0l3Q2dZSUtvWkl6ajBFQXdJRFNBQXcKUlFJZ2NyWHo3YXFkZnRrdXh6MzlQV3RjeDBKMkpGTE9EL3hzY2gvWUtGRVFPWFVDSVFDVURQVXpJK25jTjF1TgoySHVqd0FNaWd5RlJEd0ZIOVBJTUtrZDl0KytNa0E9PQotLS0tLUVORCBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0K"),
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
