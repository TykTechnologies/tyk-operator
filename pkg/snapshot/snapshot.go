package snapshot

import (
	"bufio"
	"context"
	jsonEncoding "encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/dashboard"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// names stores each ApiDefiniton's ID as a key and .metadata.name field of corresponding ApiDefinition.
var names map[string]string

const (
	NameKey      = "k8sName"
	NamespaceKey = "k8sNs"
	DefaultName  = "replace-me"
	DefaultNs    = "default"
)

// PrintSnapshot outputs a snapshot of the Dashboard as a CR.
func PrintSnapshot(
	ctx context.Context,
	env *environmet.Env,
	apiDefinitionsFile,
	policiesFile,
	category string,
	dumpAll bool,
) error {
	apiDefSpecList, err := klient.Universal.Api().List(ctx)
	if err != nil {
		return err
	}

	var policiesList []tykv1alpha1.SecurityPolicySpec
	if policiesFile != "" {
		policiesList, err = fetchPolicies(env)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(apiDefinitionsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriter(f)
	e := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	createApiDef := func(metaName, metaNs string) tykv1alpha1.ApiDefinition {
		return tykv1alpha1.ApiDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ApiDefinition",
				APIVersion: "tyk.tyk.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      metaName,
				Namespace: metaNs,
			},
			Spec: tykv1alpha1.APIDefinitionSpec{},
		}
	}

	names = make(map[string]string)

	if dumpAll {
		for i, v := range apiDefSpecList.Apis {
			// Parse Config Data of the ApiDefinition created on Dashboard.
			name, ns := parseConfigData(v, fmt.Sprintf("%s-%d", DefaultName, i))

			// create an ApiDefinition object.
			apiDef := createApiDef(name, ns)
			apiDef.Spec.APIDefinitionSpec = *v

			names[apiDef.Spec.APIID] = apiDef.ObjectMeta.Name

			if err := e.Encode(&apiDef, bw); err != nil {
				return err
			}

			if _, err := bw.WriteString("\n---\n\n"); err != nil {
				return err
			}
		}

		if policiesFile != "" {
			policyFile, err := os.Create(policiesFile)
			if err != nil {
				return err
			}
			defer policyFile.Close()

			pw := bufio.NewWriter(policyFile)
			if err := exportPolicies(policiesList, pw, e); err != nil {
				return err
			}
		}

		return bw.Flush()
	}

	category = strings.TrimSpace(category)
	if !strings.HasPrefix(category, "#") {
		category = fmt.Sprintf("#%s", category)
	}

	i := 0

	for _, v := range apiDefSpecList.Apis {
		if contains := strings.Contains(v.Name, category); !contains {
			continue
		}

		// Parse Config Data of the ApiDefinition created on Dashboard.
		name, ns := parseConfigData(v, fmt.Sprintf("%s-%d", DefaultName, i))

		// create an ApiDefinition object.
		apiDef := createApiDef(name, ns)
		apiDef.Spec.APIDefinitionSpec = *v

		names[apiDef.Spec.APIID] = apiDef.ObjectMeta.Name
		i++

		if err := e.Encode(&apiDef, bw); err != nil {
			return err
		}

		if _, err := bw.WriteString("\n---\n\n"); err != nil {
			return err
		}
	}

	if policiesFile != "" {
		policyFile, err := os.Create(policiesFile)
		if err != nil {
			return err
		}
		defer policyFile.Close()

		pw := bufio.NewWriter(policyFile)
		if err := exportPolicies(policiesList, pw, e); err != nil {
			return err
		}
	}

	return bw.Flush()
}

// fetchPolicies lists all SecurityPolicy objects form dashboard.
func fetchPolicies(e *environmet.Env) ([]tykv1alpha1.SecurityPolicySpec, error) {
	url := client.JoinURL(e.URL, "/api/portal/policies?p=-2")
	method := http.MethodGet

	hc := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", e.Auth)

	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	policiesResponse := &dashboard.PoliciesResponse{}

	err = jsonEncoding.NewDecoder(res.Body).Decode(&policiesResponse)
	if err != nil {
		return nil, err
	}

	return policiesResponse.Policies, nil
}

// parseConfigData parses given ApiDefinitionSpec's ConfigData field. It checks existence of NameKey and NamespaceKey
// keys in the ConfigData map. Returns their values if keys exist. Otherwise, returns default values for name and namespace.
func parseConfigData(apiDefSpec *model.APIDefinitionSpec, defName string) (name, namespace string) {
	if apiDefSpec.ConfigData == nil {
		return defName, DefaultNs
	}

	// Parse name
	val, ok := apiDefSpec.ConfigData.Object[NameKey]
	if !ok {
		return defName, DefaultNs
	}

	name, ok = val.(string)
	if !ok {
		return defName, DefaultNs
	}

	// Warn if .metadata.name includes an empty character because it violates k8s spec rules.
	name = strings.TrimSpace(name)
	if strings.Contains(name, " ") {
		fmt.Printf(
			"WARNING: Please ensure that API identified by %s does not include empty space in its ConfigData[%s].\n",
			apiDefSpec.APIID,
			NameKey,
		)
	}

	// Parse namespace
	val, ok = apiDefSpec.ConfigData.Object[NamespaceKey]
	if !ok {
		return name, DefaultNs
	}

	namespace, ok = val.(string)
	if !ok {
		return name, DefaultNs
	}

	// Warn if .metadata.namespace includes an empty character because it violates k8s spec rules.
	namespace = strings.TrimSpace(namespace)
	if strings.Contains(namespace, " ") {
		fmt.Printf(
			"WARNING: Please ensure that API identified by %s does not include empty space in its ConfigData[%s].\n",
			apiDefSpec.APIID,
			NamespaceKey,
		)
	}

	return
}

// exportPolicies writes all policies to the given buffer in a YAML format.
func exportPolicies(policiesList []tykv1alpha1.SecurityPolicySpec, w *bufio.Writer, e *json.Serializer) error {
	for i := 0; i < len(policiesList); i++ {
		pol := tykv1alpha1.SecurityPolicy{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SecurityPolicy",
				APIVersion: "tyk.tyk.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("security-policy-%d", i)},
			Spec:       tykv1alpha1.SecurityPolicySpec{},
		}

		pol.Spec = policiesList[i]
		pol.Spec.ID = policiesList[i].MID

		for i := 0; i < len(pol.Spec.AccessRightsArray); i++ {
			apiID := pol.Spec.AccessRightsArray[i].APIID

			name := names[apiID]
			if name == "" {
				fmt.Printf("WARNING: Please ensure that API identified by %s exists in k8s environment.\n", apiID)
			}

			p, ok := pol.Spec.AccessRights[apiID]
			if ok {
				p.Name = name
				p.Namespace = "default"

				pol.Spec.AccessRights[apiID] = p
			}

			pol.Spec.AccessRightsArray[i].Name = name
			pol.Spec.AccessRightsArray[i].Namespace = "default"
		}

		if err := e.Encode(&pol, w); err != nil {
			return err
		}

		if _, err := w.WriteString("\n---\n\n"); err != nil {
			return err
		}
	}

	return w.Flush()
}
