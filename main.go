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
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/TykTechnologies/tyk-operator/pkg/dashboard_admin_client"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/dashboard_client"
	"github.com/TykTechnologies/tyk-operator/pkg/gateway_client"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	//"github.com/TykTechnologies/tyk-operator/internal/gateway_client"
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

	watchNamespace, found := getWatchNamespace()
	if !found {
		setupLog.Info("unable to get WatchNamespace, " +
			"the manager will watch and manage resources in all Namespaces")
	}

	options := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "91ad8c6e.tyk.io",
		Namespace:          watchNamespace,
	}

	// Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
	if strings.Contains(watchNamespace, ",") {
		setupLog.Info(fmt.Sprintf("manager will be watching namespace %q", watchNamespace))
		// configure cluster-scoped with MultiNamespacedCacheBuilder
		options.Namespace = ""
		options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(watchNamespace, ","))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	//if err = (&controllers.GatewayReconciler{
	//	Client: mgr.GetClient(),
	//	Log:    ctrl.Log.WithName("controllers").WithName("Gateway"),
	//	Scheme: mgr.GetScheme(),
	//}).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Gateway")
	//	os.Exit(1)
	//}

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

	tykClient, err := tykClient()
	if err != nil {
		setupLog.Error(err, "unable to configure Tyk Client")
		os.Exit(1)
	}

	if err = (&controllers.ApiDefinitionReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("ApiDefinition"),
		Scheme:          mgr.GetScheme(),
		UniversalClient: tykClient,
		Recorder:        mgr.GetEventRecorderFor("apidefinition-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ApiDefinition")
		os.Exit(1)
	}

	if err = (&controllers.SecurityPolicyReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("SecurityPolicy"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("securitypolicy-controller"),
		UniversalClient: tykClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecurityPolicy")
		os.Exit(1)
	}

	if err = (&controllers.WebhookReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("Webhook"),
		Scheme:          mgr.GetScheme(),
		UniversalClient: tykClient,
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
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, bool) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	return os.LookupEnv(watchNamespaceEnvVar)
}

func adminClient() (*dashboard_admin_client.Client, error) {
	mode := strings.TrimSpace(os.Getenv("TYK_MODE"))
	insecureSkipVerify, err := strconv.ParseBool(os.Getenv("TYK_TLS_INSECURE_SKIP_VERIFY"))
	if err != nil {
		insecureSkipVerify = false
	}
	url := strings.TrimSpace(os.Getenv("TYK_URL"))
	if url == "" {
		return nil, errors.New("missing TYK_URL")
	}
	// ADMIN AUTH NOT MANDATORY - AS WE ARE NOT MANAGING ORGS YET
	auth := strings.TrimSpace(os.Getenv("TYK_ADMIN_AUTH"))
	if auth == "" {
		return nil, nil
	}

	switch mode {
	case "pro":
		return dashboard_admin_client.NewClient(
			url,
			auth,
			insecureSkipVerify,
		), nil
	case "oss":
		{
			return nil, nil
		}
	default:
		return nil, errors.New("unknown TYK_MODE")
	}
}

func tykClient() (universal_client.UniversalClient, error) {
	mode := strings.TrimSpace(os.Getenv("TYK_MODE"))
	insecureSkipVerify, err := strconv.ParseBool(os.Getenv("TYK_TLS_INSECURE_SKIP_VERIFY"))
	if err != nil {
		insecureSkipVerify = false
	}
	url := strings.TrimSpace(os.Getenv("TYK_URL"))
	if url == "" {
		return nil, errors.New("missing TYK_URL")
	}
	auth := strings.TrimSpace(os.Getenv("TYK_AUTH"))
	if auth == "" {
		return nil, errors.New("missing TYK_AUTH")
	}
	org := strings.TrimSpace(os.Getenv("TYK_ORG"))
	if org == "" {
		return nil, errors.New("missing TYK_ORG")
	}

	switch mode {
	case "pro":
		return dashboard_client.NewClient(
			url,
			auth,
			insecureSkipVerify,
			org,
		), nil
	case "oss":
		{
			return gateway_client.NewClient(
				url,
				auth,
				insecureSkipVerify,
				org,
			), nil
		}
	default:
		return nil, errors.New("unknown TYK_MODE")
	}
}
