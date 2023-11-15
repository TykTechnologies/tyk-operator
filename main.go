/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/config"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	"github.com/TykTechnologies/tyk-operator/pkg/snapshot"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	// snapshot specific flags
	apiDefFileFlag   string
	policyFileFlag   string
	categoryFlag     string
	separateFileFlag bool
)

func init() {
	flag.StringVar(&apiDefFileFlag, "apidef", "",
		"By passing an export flag, we are telling the Operator to connect to a "+
			"Tyk installation in order to pull a snapshot of ApiDefinitions from that environment and output as CR")

	flag.BoolVar(&separateFileFlag, "separate", false, "Each ApiDefinition and Policy files will be written "+
		"into separate files.",
	)

	flag.StringVar(&categoryFlag, "category", "", "Dump APIs from specified category.")

	flag.StringVar(&policyFileFlag, "policy", "",
		"By passing an export flag, we are telling the Operator to connect to a "+
			"Tyk installation in order to pull a snapshot of SecurityPolicies from that environment and output as CR")

	runSnapshot := apiDefFileFlag != "" || policyFileFlag != "" || separateFileFlag
	if runSnapshot {
		setupLog.Info("running snapshot tool")
	} else {
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))

		utilruntime.Must(tykv1alpha1.AddToScheme(scheme))
		// +kubebuilder:scaffold:scheme
	}
}

func main() {
	var env environment.Env
	var err error

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	env.Parse()

	runSnapshot := apiDefFileFlag != "" || policyFileFlag != "" || separateFileFlag
	if runSnapshot {
		snapshotLog := ctrl.Log.WithName("snapshot").WithName("ApiDefinition")

		_, ctx, err := controllers.HttpContext(context.Background(), nil, &env, nil, snapshotLog)
		if err != nil {
			snapshotLog.Error(err, "failed to set HTTP context")
			os.Exit(1)
		}

		if err := snapshot.PrintSnapshot(ctx, apiDefFileFlag, policyFileFlag, categoryFlag, separateFileFlag); err != nil {
			snapshotLog.Error(err, "failed to run snapshot")
			os.Exit(1)
		}

		os.Exit(0)
	}

	managerOpts, err := config.LoadFromEnv()
	if err != nil {
		setupLog.Error(err, "failed to process environment variables for manager configuration")
		os.Exit(1)
	}

	options := managerOpts.ManagerOptions(scheme)

	// If any number of namespaces is specified, only watch these namespaces for caching. In order
	// to do that, caching options must be set. Otherwise, if no namespace is specified which means
	// env.Namespace is empty string (""), no need to set caching options in controller config.
	if env.Namespace != "" {
		setupLog.Info("watching resources in namespaces", env.Namespace)

		config.SetCacheOptions(env.Namespace, &options)
	} else {
		setupLog.Info("unable to get WATCH_NAMESPACE, " +
			"the manager will watch and manage resources in all namespaces")
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	a := ctrl.Log.WithName("controllers").WithName("ApiDefinition")

	if err = (&controllers.ApiDefinitionReconciler{
		Client:   mgr.GetClient(),
		Log:      a,
		Scheme:   mgr.GetScheme(),
		Env:      env,
		Recorder: mgr.GetEventRecorderFor("apidefinition-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ApiDefinition")
		os.Exit(1)
	}

	il := ctrl.Log.WithName("controllers").WithName("Ingress")

	if err = (&controllers.IngressReconciler{
		Client:   mgr.GetClient(),
		Log:      il,
		Scheme:   mgr.GetScheme(),
		Env:      env,
		Recorder: mgr.GetEventRecorderFor("ingress-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Ingress")
		os.Exit(1)
	}

	sl := ctrl.Log.WithName("controllers").WithName("SecretCert")

	if err = (&controllers.SecretCertReconciler{
		Client: mgr.GetClient(),
		Log:    sl,
		Scheme: mgr.GetScheme(),
		Env:    env,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretCert")
		os.Exit(1)
	}

	spg := ctrl.Log.WithName("controllers").WithName("SuperGraph")

	if err = (&controllers.SuperGraphReconciler{
		Client: mgr.GetClient(),
		Log:    spg,
		Scheme: mgr.GetScheme(),
		Env:    env,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SuperGraph")
		os.Exit(1)
	}

	sp := ctrl.Log.WithName("controllers").WithName("SecurityPolicy")

	if err = (&controllers.SecurityPolicyReconciler{
		Client:   mgr.GetClient(),
		Log:      sp,
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("securitypolicy-controller"),
		Env:      env,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecurityPolicy")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&tykv1alpha1.ApiDefinition{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "ApiDefinition")
			os.Exit(1)
		}

		if err = (&tykv1alpha1.SecurityPolicy{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "SecurityPolicy")
			os.Exit(1)
		}
	}

	if err = (&controllers.APIDescriptionReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("APIDescription"),
		Scheme: mgr.GetScheme(),
		Env:    env,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "APIDescription")
		os.Exit(1)
	}

	if err = (&controllers.PortalAPICatalogueReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("PortalAPICatalogue"),
		Scheme: mgr.GetScheme(),
		Env:    env,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PortalAPICatalogue")
		os.Exit(1)
	}

	if err = (&controllers.PortalConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("PortalConfig"),
		Scheme: mgr.GetScheme(),
		Env:    env,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PortalConfig")
		os.Exit(1)
	}
	if err = (&controllers.OperatorContextReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    ctrl.Log.WithName("controllers").WithName("OperatorContext"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OperatorContext")
		os.Exit(1)
	}

	if err = (&controllers.SubGraphReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    ctrl.Log.WithName("controllers").WithName("SubGraph"),
		Env:    env,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SubGraph")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
