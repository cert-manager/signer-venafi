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

package main

import (
	"flag"
	"os"

	capi "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Venafi/vcert"
	"github.com/cert-manager/signer-venafi/controllers"
	"github.com/cert-manager/signer-venafi/internal/signer/venafi"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = capi.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		leaderElectionID     string
		debugLogging         bool
		signerName           string
		vcertConfigPath      string
	)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&leaderElectionID, "leader-election-id", "signer-venafi-leader-election",
		"The name of the configmap used to coordinate leader election between controller-managers.")
	flag.BoolVar(&debugLogging, "debug-logging", true, "Enable debug logging.")
	flag.StringVar(&signerName, "signer-name", "example.com/foo", "Only sign CSR with this .spec.signerName.")
	flag.StringVar(&vcertConfigPath, "vcert-config", "vcert.ini", "Vcert INI file path.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(debugLogging)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	vcertConfig := &vcert.Config{
		ConfigFile: vcertConfigPath,
	}
	err = vcertConfig.LoadFromFile()
	if err != nil {
		setupLog.Error(err, "unable load vcert config file", "vcert-config-path", vcertConfigPath)
		os.Exit(1)
	}

	vcertClient, err := vcert.NewClient(vcertConfig)
	if err != nil {
		setupLog.Error(err, "unable initialize vcert client", "vcert-config-path", vcertConfigPath)
		os.Exit(1)
	}

	signer := &venafi.Signer{
		Client: vcertClient,
		Log:    ctrl.Log.WithName("signer").WithName("venafi").WithName("Signer"),
	}

	if err = (&controllers.CertificateSigningRequestReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("CertificateSigningRequestReconciler"),
		Scheme:     mgr.GetScheme(),
		Signer:     signer,
		SignerName: signerName,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertificateSigningRequestReconciler")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
