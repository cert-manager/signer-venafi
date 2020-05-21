/*
Copyright 2020 The Cert-Manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	capi "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capihelper "github.com/cert-manager/signer-venafi/internal/api"
	"github.com/cert-manager/signer-venafi/internal/signer"
)

// CertificateSigningRequestReconciler reconciles a CertificateSigningRequest object
type CertificateSigningRequestReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	Signer     signer.Signer
	SignerName string
}

// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=get;update;patch

func (r *CertificateSigningRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithName("Reconcile").WithValues("certificatesigningrequest", req.NamespacedName)
	ctx := context.Background()

	var csr capi.CertificateSigningRequest
	if err := r.Client.Get(ctx, req.NamespacedName, &csr); err != nil {
		if client.IgnoreNotFound(err) == nil {
			log.Info("CSR not found. Ignoring.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("error getting CSR: %v", err)
	}

	switch {
	case csr.Spec.SignerName == nil:
		log.V(1).Info("CSR does not have a signer name. Ignoring.")
	case *csr.Spec.SignerName != r.SignerName:
		log.V(1).Info("CSR signer name does not match Reconciler signer name. Ignoring.", "signer-name", csr.Spec.SignerName)
	case !capihelper.IsCertificateRequestApproved(&csr):
		log.V(1).Info("CSR is not approved, Ignoring.")
	case csr.Status.Certificate != nil:
		log.V(1).Info("CSR has already been signed. Ignoring.")
	case string(csr.Status.Certificate) != "":
		log.V(1).Info("CSR has already been signed. Ignoring.")
	case csr.Annotations["pickup-id"] == "":
		log.V(1).Info("Signing")

		pickupID, err := r.Signer.Sign(csr)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error signing: %v", err)
		}

		original := csr.DeepCopy()
		metav1.SetMetaDataAnnotation(&csr.ObjectMeta, "pickup-id", pickupID)

		if reflect.DeepEqual(original, csr) {
			return ctrl.Result{}, nil
		}

		patch := client.MergeFrom(original)
		if err := r.Client.Status().Patch(ctx, &csr, patch); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching CSR: %v", err)
		}
	default:
		log.V(1).Info("Picking up")

		pickupID := csr.Annotations["pickup-id"]

		certificate, err := r.Signer.Pickup(pickupID)
		if err != nil {
			if errors.Is(err, signer.ErrTemporary) {
				log.V(1).Info("Temporary error picking up certificate", "err", err)
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			return ctrl.Result{}, fmt.Errorf("error signing: %v", err)
		}

		original := csr.DeepCopy()
		delete(csr.Annotations, "pickup-id")
		csr.Status.Certificate = certificate

		if reflect.DeepEqual(original, csr) {
			return ctrl.Result{}, nil
		}

		patch := client.MergeFrom(original)
		if err := r.Client.Status().Patch(ctx, &csr, patch); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching CSR: %v", err)
		}

	}

	return ctrl.Result{}, nil
}

func (r *CertificateSigningRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capi.CertificateSigningRequest{}).
		Complete(r)
}
