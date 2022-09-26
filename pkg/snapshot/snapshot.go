package snapshot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
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

	// idAPIs stores ApiDefinitions based on its ID.
	idAPIs map[string]*model.APIDefinitionSpec

	// ErrNonexistentKey represents an error if the given key does not exist in the object.
	ErrNonexistentKey = errors.New("key does not exist in the Config Data")

	// ErrNonStringVal represents an error if the underlying value of interface{} is not string type.
	ErrNonStringVal = errors.New("failed to convert interface{} to string")
)

type Group struct {
	Policy         tykv1alpha1.SecurityPolicySpec
	APIDefinitions model.APIDefinitionSpecList
}

type Groups []Group

const (
	NameKey      = "k8sName"
	NamespaceKey = "k8sNamespace"
	DefaultName  = "REPLACE_ME"
	DefaultNs    = ""
)

// PrintSnapshot outputs a snapshot of the Dashboard as a CR.
func PrintSnapshot(
	ctx context.Context,
	env *environmet.Env,
	apiDefinitionsFile,
	policiesFile,
	category string,
	group bool,
) error {
	apiDefSpecList, err := klient.Universal.Api().List(ctx)
	if err != nil {
		return err
	}

	var policiesList []tykv1alpha1.SecurityPolicySpec
	if policiesFile != "" || group {
		policiesList, err = klient.Universal.Portal().Policy().All(ctx)
		if err != nil {
			return err
		}
	}

	names = make(map[string]string)
	nameSpaces = make(map[string]string)

	e := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	exportApiDefs := func(i int, w io.Writer, v *model.APIDefinitionSpec) error {
		// Parse Config Data of the ApiDefinition created on Dashboard.
		name, ns, err := parseConfigData(v, fmt.Sprintf("%s_%d", DefaultName, i))
		if err != nil {
			// Instead of terminating the function with error, just do not export ApiDefinitions with invalid ConfigData
			// fields, instead inform users with the following log message.
			fmt.Printf("WARNING: failed to parse API %v due to malformed ConfigData, err: %v\n", v.APIID, err)
			return nil
		}

		// create an ApiDefinition object.
		apiDef := createApiDef(name, ns)
		apiDef.Spec.APIDefinitionSpec = *v

		storeMetadata(apiDef.Spec.APIID, apiDef.ObjectMeta.Name, apiDef.ObjectMeta.Namespace)

		if err := e.Encode(&apiDef, w); err != nil {
			return err
		}

		if _, err := w.Write([]byte("\n---\n\n")); err != nil {
			return err
		}

		return nil
	}

	// Output file will contain ApiDefinitions grouped by SecurityPolicies.
	if group {
		idAPIs = make(map[string]*model.APIDefinitionSpec)
		for i := 0; i < len(apiDefSpecList.Apis); i++ {
			idAPIs[apiDefSpecList.Apis[i].APIID] = apiDefSpecList.Apis[i]
		}

		groups := groupPolicies(policiesList)
		for i := 0; i < len(groups); i++ {
			g := groups[i]

			groupedFile, err := os.Create(fmt.Sprintf("grouped-output-%d.yaml", i))
			if err != nil {
				return err
			}

			pw := bufio.NewWriter(groupedFile)

			for ii, v := range g.APIDefinitions.Apis {
				if err := exportApiDefs(ii, pw, v); err != nil {
					return err
				}
			}

			if err := writePolicies([]tykv1alpha1.SecurityPolicySpec{g.Policy}, pw, e); err != nil {
				groupedFile.Close()
				return err
			}

			pw.Flush()
			groupedFile.Close()
		}

		return nil
	}

	f, err := os.Create(apiDefinitionsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriter(f)

	exportPolicies := func() error {
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

		return nil
	}

	// Output file will contain ApiDefinition based on specified category.
	if category != "" {
		category = strings.TrimSpace(category)
		if !strings.HasPrefix(category, "#") {
			category = fmt.Sprintf("#%s", category)
		}

		fmt.Printf("Looking for ApiDefinitions in %s category.\n", category)

		for i, v := range apiDefSpecList.Apis {
			if contains := strings.Contains(v.Name, category); !contains {
				continue
			}

			if err := exportApiDefs(i, bw, v); err != nil {
				return err
			}
		}

		if err := exportPolicies(); err != nil {
			return err
		}

		return bw.Flush()
	}

	// Output file will contain all ApiDefinitions without checking any category.
	for i, v := range apiDefSpecList.Apis {
		if err := exportApiDefs(i, bw, v); err != nil {
			return err
		}
	}

	if err := exportPolicies(); err != nil {
		return err
	}

	return bw.Flush()
}

func createApiDef(metaName, metaNs string) tykv1alpha1.ApiDefinition {
	meta := metav1.ObjectMeta{Name: metaName}
	if metaNs != "" {
		meta.Namespace = metaNs
	}

	return tykv1alpha1.ApiDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApiDefinition",
			APIVersion: "tyk.tyk.io/v1alpha1",
		},
		ObjectMeta: meta,
		Spec:       tykv1alpha1.APIDefinitionSpec{},
	}
}

func storeMetadata(key, name, namespace string) {
	names[key] = name
	nameSpaces[key] = namespace
}

func getMetadata(key string) (name, namespace string) {
	return names[key], nameSpaces[key]
}

// val returns given key's value from given map in string format.
func val(obj map[string]interface{}, key string) (string, error) {
	v, ok := obj[key]
	if !ok {
		return "", ErrNonexistentKey
	}

	strVal, ok := v.(string)
	if !ok {
		return "", ErrNonStringVal
	}

	return strings.TrimSpace(strVal), nil
}

var (
	ErrNonExistentConfigData = errors.New("failed to find ConfigData: non existent")
)

// parseConfigData parses given ApiDefinitionSpec's ConfigData field. It checks existence of NameKey and NamespaceKey
// keys in the ConfigData map. Returns their values if keys exist. Otherwise, returns default values for name and namespace.
func parseConfigData(apiDefSpec *model.APIDefinitionSpec, defName string) (name, namespace string, err error) {
	if apiDefSpec.ConfigData == nil {
		return defName, DefaultNs, ErrNonExistentConfigData
	}

	// Parse name
	name, err = val(apiDefSpec.ConfigData.Object, NameKey)
	if err != nil {
		return defName, DefaultNs, fmt.Errorf("failed to parse k8s name from ConfigData, err: %v", err)
	}

	namespace, _ = val(apiDefSpec.ConfigData.Object, NamespaceKey)

	// Warn if .metadata includes an empty character because it violates k8s spec rules.
	for _, v := range []string{name, namespace} {
		if strings.Contains(v, " ") {
			fmt.Printf(
				"WARNING: Please ensure that API identified by %s does not include empty space in its ConfigData[%s].\n",
				apiDefSpec.APIID,
				NamespaceKey,
			)
		}
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
			g.APIDefinitions.Apis = append(g.APIDefinitions.Apis, idAPIs[ar.APIID])
		}

		groups = append(groups, g)
	}

	return groups
}
