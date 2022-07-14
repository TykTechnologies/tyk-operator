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
	"flag"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(tykv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var configFile string
	var env environmet.Env
	var err error

	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	env.Parse()

	if env.Namespace == "" {
		setupLog.Info("unable to get WatchNamespace, " +
			"the manager will watch and manage resources in all Namespaces")
	}

	options := ctrl.Options{Scheme: scheme, Namespace: env.Namespace}

	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	// Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
	if strings.Contains(env.Namespace, ",") {
		setupLog.Info(fmt.Sprintf("manager will be watching namespace %q", env.Namespace))
		// configure cluster-scoped with MultiNamespacedCacheBuilder
		options.Namespace = ""
		options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(env.Namespace, ","))
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
	if err = (&controllers.EndpointsReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Endpoints")
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
