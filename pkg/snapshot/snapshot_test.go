package snapshot

import (
	"fmt"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/TykTechnologies/tyk-operator/api/model"

	"github.com/matryer/is"
)

var (
	existingName         = "existing-filename"
	existingFullFilename = fmt.Sprintf("%s.yaml", existingName)

	nonexistentName         = "nonexistent-filename"
	nonexistentFullFilename = fmt.Sprintf("%s.yaml", nonexistentName)
)

func TestGenerateFilename(t *testing.T) {
	tmpDir := t.TempDir()
	eval := is.New(t)

	origDir, err := os.Getwd()
	eval.NoErr(err)

	err = os.Chdir(tmpDir)
	eval.NoErr(err)

	defer os.Chdir(origDir) //nolint:errcheck

	_, err = os.Create(existingFullFilename)
	eval.NoErr(err)

	testCases := []struct {
		filename       string
		expectedOutput string
	}{
		{
			filename:       existingName,
			expectedOutput: fmt.Sprintf("%s-1.yaml", existingName),
		},
		{
			filename:       nonexistentName,
			expectedOutput: nonexistentFullFilename,
		},
		{
			filename:       "",
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		fn, err := generateFilename(tc.filename)
		eval.NoErr(err)
		eval.Equal(fn, tc.expectedOutput)
	}
}

func TestCreateAPIDef(t *testing.T) {
	const tempNs = "temp-ns"
	apiDef := createApiDef(existingName, "")
	eval := is.New(t)

	eval.Equal(apiDef.ObjectMeta.Name, existingName)
	eval.Equal(apiDef.ObjectMeta.Namespace, "")

	eval.Equal(apiDef.TypeMeta.Kind, ApiDefinitionKind)
	eval.Equal(apiDef.TypeMeta.APIVersion, ApiVersion)

	apiDef = createApiDef(existingName, tempNs)
	eval.Equal(apiDef.ObjectMeta.Name, existingName)
	eval.Equal(apiDef.ObjectMeta.Namespace, tempNs)

	eval.Equal(apiDef.TypeMeta.Kind, ApiDefinitionKind)
	eval.Equal(apiDef.TypeMeta.APIVersion, ApiVersion)
}

func TestParseConfigData(t *testing.T) {
	eval := is.New(t)

	const tmpNs = "namespace"

	testCases := []struct {
		configData *model.MapStringInterfaceType
		name       string
		namespace  string
		err        error
	}{
		{
			configData: nil,
			name:       existingName,
			namespace:  "",
			err:        ErrNonExistentConfigData,
		},
		{
			configData: &model.MapStringInterfaceType{
				Unstructured: unstructured.Unstructured{
					Object: map[string]interface{}{
						"k8sName": existingName,
					},
				},
			},
			name: existingName,
			err:  nil,
		},
		{
			configData: &model.MapStringInterfaceType{
				Unstructured: unstructured.Unstructured{
					Object: map[string]interface{}{
						"k8sName":      existingName,
						"k8sNamespace": tmpNs,
					},
				},
			},
			name:      existingName,
			namespace: tmpNs,
			err:       nil,
		},
		{
			configData: &model.MapStringInterfaceType{
				Unstructured: unstructured.Unstructured{
					Object: map[string]interface{}{
						"k8sName":    existingName,
						"anotherKey": tmpNs,
					},
				},
			},
			name:      existingName,
			namespace: "",
			err:       nil,
		},
		{
			configData: &model.MapStringInterfaceType{
				Unstructured: unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			err: ErrInvalidConfigData,
		},
	}

	for _, tc := range testCases {
		apiDefSpec := &model.APIDefinitionSpec{ConfigData: tc.configData}

		n, ns, err := parseConfigData(apiDefSpec, existingName)

		eval.Equal(err, tc.err)
		eval.Equal(n, existingName)
		eval.Equal(ns, tc.namespace)
	}
}
