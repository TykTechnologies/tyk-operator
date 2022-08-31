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

var (
	// names stores each ApiDefiniton's ID as a key and .metadata.name field of corresponding ApiDefinition.
	names map[string]string

	// namesSpaces stores each ApiDefiniton's ID as a key and .metadata.namespace field of corresponding ApiDefinition.
	nameSpaces map[string]string

	idName map[string]*model.APIDefinitionSpec
)

type Group struct {
	Policy         tykv1alpha1.SecurityPolicySpec
	APIDefinitions model.APIDefinitionSpecList
}

type Groups []Group

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
	dumpAll,
	group bool,
) error {
	apiDefSpecList, err := klient.Universal.Api().List(ctx)
	if err != nil {
		return err
	}

	var policiesList []tykv1alpha1.SecurityPolicySpec
	if policiesFile != "" || group {
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
	nameSpaces = make(map[string]string)

	// Output file will contain ApiDefinitions grouped by SecurityPolicies.
	if group {
		idName = make(map[string]*model.APIDefinitionSpec)
		for i := 0; i < len(apiDefSpecList.Apis); i++ {
			idName[apiDefSpecList.Apis[i].APIID] = apiDefSpecList.Apis[i]
		}

		groups := groupPolicies(policiesList)
		for i := 0; i < len(groups); i++ {
			g := groups[i]

			groupedFile, err := os.Create(fmt.Sprintf("grouped-pol-%d.yaml", i))
			if err != nil {
				return err
			}

			pw := bufio.NewWriter(groupedFile)

			if err := writePolicies([]tykv1alpha1.SecurityPolicySpec{g.Policy}, pw, e); err != nil {
				groupedFile.Close()
				return err
			}

			for ii, v := range g.APIDefinitions.Apis {
				// Parse Config Data of the ApiDefinition created on Dashboard.
				name, ns := parseConfigData(v, fmt.Sprintf("%s-%d", DefaultName, ii))

				// create an ApiDefinition object.
				apiDef := createApiDef(name, ns)
				apiDef.Spec.APIDefinitionSpec = *v

				if err := e.Encode(&apiDef, pw); err != nil {
					groupedFile.Close()
					return err
				}

				if _, err := pw.WriteString("\n---\n"); err != nil {
					groupedFile.Close()
					return err
				}
			}

			pw.Flush()
			groupedFile.Close()
		}

		return nil
	}

	// Output file will contain all ApiDefinitions without checking any category.
	if dumpAll {
		for i, v := range apiDefSpecList.Apis {
			// Parse Config Data of the ApiDefinition created on Dashboard.
			name, ns := parseConfigData(v, fmt.Sprintf("%s-%d", DefaultName, i))

			// create an ApiDefinition object.
			apiDef := createApiDef(name, ns)
			apiDef.Spec.APIDefinitionSpec = *v

			// store metadata of this ApiDefinition in memory.
			storeMetadata(apiDef.Spec.APIID, apiDef.ObjectMeta.Name, apiDef.ObjectMeta.Namespace)

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
			if err := writePolicies(policiesList, pw, e); err != nil {
				return err
			}
		}

		return bw.Flush()
	}

	// Output file will contain ApiDefinition based on specified category.
	category = strings.TrimSpace(category)
	if !strings.HasPrefix(category, "#") {
		category = fmt.Sprintf("#%s", category)
	}
	fmt.Printf("Looking for ApiDefinitions in %s category.\n", category)

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

		storeMetadata(apiDef.Spec.APIID, apiDef.ObjectMeta.Name, apiDef.ObjectMeta.Namespace)
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
		if err := writePolicies(policiesList, pw, e); err != nil {
			return err
		}
	}

	return bw.Flush()
}

func storeMetadata(key, name, namespace string) {
	names[key] = name
	nameSpaces[key] = namespace
}

func getMetadata(key string) (name, namespace string) {
	return names[key], nameSpaces[key]
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

// writePolicies writes all policies to the given buffer in a YAML format.
func writePolicies(policiesList []tykv1alpha1.SecurityPolicySpec, w *bufio.Writer, e *json.Serializer) error {
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

			name, namespace := getMetadata(apiID)
			if name == "" || namespace == "" {
				fmt.Printf("WARNING: Please ensure that API identified by %s exists in k8s environment.\n", apiID)
			}

			p, ok := pol.Spec.AccessRights[apiID]
			if ok {
				p.Name = name
				p.Namespace = namespace

				pol.Spec.AccessRights[apiID] = p
			}

			pol.Spec.AccessRightsArray[i].Name = name
			pol.Spec.AccessRightsArray[i].Namespace = namespace
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

func groupPolicies(policiesList []tykv1alpha1.SecurityPolicySpec) Groups {
	groups := Groups{}

	for i := 0; i < len(policiesList); i++ {
		g := Group{Policy: policiesList[i]}
		for _, ar := range policiesList[i].AccessRightsArray {
			g.APIDefinitions.Apis = append(g.APIDefinitions.Apis, idName[ar.APIID])
		}

		groups = append(groups, g)
	}

	return groups
}
