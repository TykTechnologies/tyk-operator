package snapshot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var (
	// names stores each ApiDefinition's ID as a key and .metadata.name field of corresponding ApiDefinition.
	names map[string]string

	// namesSpaces stores each ApiDefinition's ID as a key and .metadata.namespace field of corresponding ApiDefinition.
	nameSpaces map[string]string

	// ErrNonExistentKey represents an error if the given key does not exist in the object.
	ErrNonExistentKey = errors.New("key does not exist in the Config Data")

	// ErrNonStringVal represents an error if the underlying value of interface{} is not string type.
	ErrNonStringVal = errors.New("failed to convert interface{} to string")

	// ErrNonExistentConfigData represents an error if the given ApiDefinition includes empty (nil) ConfigData field.
	ErrNonExistentConfigData = errors.New("failed to parse ConfigData: non existent")

	// ErrInvalidConfigData represents an error if the given ConfigData does not include the required key 'k8sName'.
	ErrInvalidConfigData = errors.New("failed to parse 'k8sName' field in ConfigData")
)

const (
	NameKey      = "k8sName"
	NamespaceKey = "k8sNamespace"
	DefaultName  = "REPLACE_ME"
	DefaultNs    = ""
)

// PrintSnapshot outputs a snapshot of the Dashboard as a CR.
func PrintSnapshot(ctx context.Context, apiDefinitionsFile, policiesFile, category string, separate bool) error {
	apiDefSpecList, err := klient.Universal.Api().List(ctx)
	if err != nil {
		return err
	}

	var policiesList []tykv1alpha1.SecurityPolicySpec
	if policiesFile != "" || separate {
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
			fmt.Printf("WARNING: failed to parse API %v due to malformed ConfigData, err: %v\n", v.APIID, err)
			return err
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

	exportPolicies := func(i int, userPolicy tykv1alpha1.SecurityPolicySpec) error {
		if policiesFile != "" {
			policyFile, err := os.Create(policiesFile)
			if err != nil {
				return err
			}
			defer policyFile.Close()

			pw := bufio.NewWriter(policyFile)
			if err := writePolicy(i, userPolicy, pw, e); err != nil {
				return err
			}
		}

		return nil
	}

	if separate {
		for i, apiDefSpec := range apiDefSpecList.Apis {
			name, ns, err := parseConfigData(apiDefSpec, "")
			if err != nil {
				fmt.Printf("WARNING: failed to parse API %v due to malformed ConfigData, err: %v\n",
					apiDefSpec.APIID,
					err,
				)
				return err
			}

			filename := fmt.Sprintf("%s-%s", ns, name)
			if ns == "" {
				filename = name
			}

			fullFilename, err := generateFilename(filename)
			if err != nil {
				return err
			}

			f, err := os.Create(fullFilename)
			if err != nil {
				return err
			}

			bw := bufio.NewWriter(f)

			if err := exportApiDefs(i, bw, apiDefSpec); err != nil && !errors.Is(err, ErrInvalidConfigData) {
				f.Close()
				return err
			}

			f.Close()
		}

		for i := 0; i < len(policiesList); i++ {
			policiesFile = fmt.Sprintf("%s-%s.yaml", "policy", policiesList[i].MID)
			if err != nil {
				return err
			}

			if err := exportPolicies(i, policiesList[i]); err != nil {
				return err
			}
		}

		return nil
	}

	f, err := os.Create(apiDefinitionsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriter(f)

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

			if err := exportApiDefs(i, bw, v); err != nil && !errors.Is(err, ErrInvalidConfigData) {
				return err
			}
		}

		for i := 0; i < len(policiesList); i++ {
			if err := exportPolicies(i, policiesList[i]); err != nil {
				return err
			}
		}

		return bw.Flush()
	}

	// Output file will contain all ApiDefinitions without checking any category.
	for i, v := range apiDefSpecList.Apis {
		if err := exportApiDefs(i, bw, v); err != nil && !errors.Is(err, ErrInvalidConfigData) {
			return err
		}
	}

	for i := 0; i < len(policiesList); i++ {
		if err := exportPolicies(i, policiesList[i]); err != nil {
			return err
		}
	}

	return bw.Flush()
}

func generateFilename(filename string) (string, error) {
	counter := 0

	_, err := os.Stat(fmt.Sprintf("%s.yaml", filename))
	if err == nil {
		fullFilename := fmt.Sprintf("%s-*.yaml", filename)
		matches, err := filepath.Glob(fullFilename)
		if err != nil {
			return "", err
		}

		// if there is no matches, it means that we do not have any files with `-*.yaml` suffix. So, this file
		// must be the first copy considering <filename>.yaml already exists.
		if len(matches) == 0 {
			counter = 1
		}

		for _, match := range matches {
			tmpFullFilename := strings.Trim(match, ".yaml")

			splitted := strings.Split(tmpFullFilename, "-")
			if len(splitted) == 0 {
				return "", fmt.Errorf(
					"unexpected filename while creating separated output, filename: %v",
					tmpFullFilename,
				)
			}

			lastCounter, err := strconv.Atoi(splitted[len(splitted)-1])
			if err != nil {
				return "", err
			}

			if counter <= lastCounter {
				counter = lastCounter + 1
			}
		}
	}

	fullFilename := fmt.Sprintf("%s.yaml", filename)
	if counter != 0 {
		fullFilename = fmt.Sprintf("%s-%d.yaml", filename, counter)
	}

	return fullFilename, nil
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
		return "", ErrNonExistentKey
	}

	strVal, ok := v.(string)
	if !ok {
		return "", ErrNonStringVal
	}

	return strings.TrimSpace(strVal), nil
}

// parseConfigData parses given ApiDefinitionSpec's ConfigData field. It checks existence of NameKey and NamespaceKey
// keys in the ConfigData map. Returns their values if keys exist. Otherwise, returns default values for name and namespace.
// Returns error in case of missing NameKey in the Config Data.
func parseConfigData(apiDefSpec *model.APIDefinitionSpec, defName string) (name, namespace string, err error) {
	if apiDefSpec.ConfigData == nil {
		return defName, DefaultNs, ErrNonExistentConfigData
	}

	// Parse name
	name, err = val(apiDefSpec.ConfigData.Object, NameKey)
	if err != nil {
		return defName, DefaultNs, ErrInvalidConfigData
	}

	namespace, _ = val(apiDefSpec.ConfigData.Object, NamespaceKey) //nolint:errcheck

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

func writePolicy(idx int, userPolicy tykv1alpha1.SecurityPolicySpec, w *bufio.Writer, e *json.Serializer) error {
	pol := tykv1alpha1.SecurityPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SecurityPolicy",
			APIVersion: "tyk.tyk.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("security-policy-%d", idx)},
		Spec:       tykv1alpha1.SecurityPolicySpec{},
	}

	pol.Spec = userPolicy
	pol.Spec.ID = userPolicy.MID

	for i := 0; i < len(pol.Spec.AccessRightsArray); i++ {
		apiID := pol.Spec.AccessRightsArray[i].APIID

		name, namespace := getMetadata(apiID)
		if name == "" {
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

	return w.Flush()
}
