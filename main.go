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
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/internal/dashboard_client"
	"github.com/TykTechnologies/tyk-operator/internal/gateway_client"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
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

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "91ad8c6e.tyk.io",
	})
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
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Webhook"),
		Scheme: mgr.GetScheme(),
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
	//// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func tykClient() (universal_client.UniversalClient, error) {
	mode := os.Getenv("TYK_MODE")
	insecureSkipVerify, err := strconv.ParseBool(os.Getenv("TYK_TLS_INSECURE_SKIP_VERIFY"))
	if err != nil {
		insecureSkipVerify = false
	}
	url := os.Getenv("TYK_URL")
	if url == "" {
		return nil, errors.New("missing TYK_URL")
	}
	auth := os.Getenv("TYK_AUTH")
	if auth == "" {
		return nil, errors.New("missing TYK_AUTH")
	}
	org := os.Getenv("TYK_ORG")
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
