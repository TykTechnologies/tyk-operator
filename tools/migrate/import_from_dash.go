package migrate

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
)

var scheme = runtime.NewScheme()

func init() {
	v1alpha1.AddToScheme(scheme)
}

func Build(o io.Writer, ns string, apis []v1alpha1.APIDefinitionSpec, policies []v1alpha1.SecurityPolicySpec) error {
	var buf bytes.Buffer
	se := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{
		Yaml: true,
	})
	name := func(n string) string {
		return types.NamespacedName{
			Namespace: ns,
			Name:      n,
		}.String()
	}

	// mapping of ID to a resource name
	names := map[string]string{}

	for i := 0; i < len(apis); i++ {
		names[apis[i].ID] = toName(apis[i].Name)
	}

	for i := 0; i < len(policies); i++ {
		names[policies[i].ID] = toName(policies[i].Name)
	}
	for i := 0; i < len(apis); i++ {
		a := &apis[i]
		for j := 0; j < len(a.JWTDefaultPolicies); j++ {
			a.JWTDefaultPolicies[j] = name(a.JWTDefaultPolicies[j])
		}
		for k, v := range a.JWTScopeToPolicyMapping {
			a.JWTScopeToPolicyMapping[k] = names[v]
		}
	}
	for i := 0; i < len(policies); i++ {
		p := &policies[i]
		for j := 0; j < len(p.AccessRightsArray); j++ {
			a := p.AccessRightsArray[j]
			a.Namespace = ns
			a.Name = names[a.APIID]

			// these will be updated by the operator
			a.APIID = ""
			a.APIName = ""
		}
		// clear dash specific stuff
		p.AccessRights = nil
	}

	for i := 0; i < len(apis); i++ {
		a := &apis[i]
		resource := &v1alpha1.ApiDefinition{}
		resource.Name = names[a.ID]
		resource.Kind = "ApiDefinition"
		resource.APIVersion = v1alpha1.GroupVersion.String()
		resource.Namespace = ns

		// clear dash specific stuff
		a.ID = ""
		a.APIID = ""

		resource.Spec = copyAPI(a)

		if err := se.Encode(resource, &buf); err != nil {
			return err
		}
	}
	for i := 0; i < len(policies); i++ {
		a := &policies[i]
		resource := &v1alpha1.SecurityPolicy{}
		resource.Name = names[a.ID]
		resource.Namespace = ns
		resource.Spec = copyPolicy(a)
		if err := se.Encode(resource, &buf); err != nil {
			return err
		}
	}
	return cleanup(o, &buf)
}

func toName(n string) string {
	n = strings.ToLower(n)
	s := bufio.NewScanner(strings.NewReader(n))
	s.Split(bufio.ScanWords)
	var words []string
	for s.Scan() {
		words = append(words, s.Text())
	}
	fmt.Println(words)
	return strings.Join(words, "-")
}

func cleanup(o io.Writer, buf *bytes.Buffer) error {
	remove := []string{
		`  creationTimestamp: null
`,
		`status:
  api_id: ""`,
	}
	b := buf.Bytes()
	for _, v := range remove {
		b = bytes.ReplaceAll(b, []byte(v), []byte(""))
	}
	_, err := o.Write(b)
	return err
}
