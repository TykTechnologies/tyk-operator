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
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/go-logr/logr"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/dashboard_client"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/gateway_client"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
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
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	var env environmet.Env
	if err := env.Parse(); err != nil {
		setupLog.Error(err, "unable to configure Tyk Client")
		os.Exit(1)
	}
	if env.Namespace == "" {
		setupLog.Info("unable to get WatchNamespace, " +
			"the manager will watch and manage resources in all Namespaces")
	}

	options := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "91ad8c6e.tyk.io",
		Namespace:          env.Namespace,
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
	//_, err := adminClient()
	//if err != nil {
	//	setupLog.Error(err, "unable to configure admin client")
	//	os.Exit(1)
	//}

	//if err = (&controllers.OrganizationReconciler{
	//	Client: mgr.GetClient(),
	//	Log:    ctrl.Log.WithName("controllers").WithName("Organization"),
	//	Scheme: mgr.GetScheme(),
	//	//Recorder:        mgr.GetEventRecorderFor("organization-controller"),
	//	AdminDashboardCient: adminClient,
	//}).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Organization")
	//	os.Exit(1)
	//}
	a := ctrl.Log.WithName("controllers").WithName("ApiDefinition")
	if err = (&controllers.ApiDefinitionReconciler{
		Client:          mgr.GetClient(),
		Log:             a,
		Scheme:          mgr.GetScheme(),
		UniversalClient: newUniversalClient(a, env),
		Recorder:        mgr.GetEventRecorderFor("apidefinition-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ApiDefinition")
		os.Exit(1)
	}

	il := ctrl.Log.WithName("controllers").WithName("Ingress")
	if err = (&controllers.IngressReconciler{
		Client:          mgr.GetClient(),
		Log:             il,
		Scheme:          mgr.GetScheme(),
		UniversalClient: newUniversalClient(il, env),
		Recorder:        mgr.GetEventRecorderFor("ingress-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Ingress")
		os.Exit(1)
	}

	sl := ctrl.Log.WithName("controllers").WithName("SecretCert")
	if err = (&controllers.SecretCertReconciler{
		Client:          mgr.GetClient(),
		Log:             sl,
		Scheme:          mgr.GetScheme(),
		UniversalClient: newUniversalClient(sl, env),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretCert")
		os.Exit(1)
	}
	sp := ctrl.Log.WithName("controllers").WithName("SecurityPolicy")
	if err = (&controllers.SecurityPolicyReconciler{
		Client:          mgr.GetClient(),
		Log:             sp,
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("securitypolicy-controller"),
		UniversalClient: newUniversalClient(sp, env),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecurityPolicy")
		os.Exit(1)
	}

	w := ctrl.Log.WithName("controllers").WithName("SecurityPolicy")
	if err = (&controllers.WebhookReconciler{
		Client:          mgr.GetClient(),
		Log:             w,
		Scheme:          mgr.GetScheme(),
		UniversalClient: newUniversalClient(w, env),
		Recorder:        mgr.GetEventRecorderFor("webhook-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Webhook")
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

	if err = (&controllers.OperatorContextReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("OperatorContext"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OperatorContext")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func newUniversalClient(log logr.Logger, env environmet.Env) universal_client.UniversalClient {
	if env.Mode == "pro" {
		return dashboard_client.NewClient(log, env)
	}
	return gateway_client.NewClient(log, env)
}
