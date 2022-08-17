package snapshot

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var names map[string]string

func PrintSnapshot(ctx context.Context, fileName, category string, dumpAll bool) error {
	apiDefSpecList, err := klient.Universal.Api().List(ctx)
	if err != nil {
		return err
	}

	policiesList, err := klient.Universal.Portal().Policy().All(ctx)
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
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

	createApiDef := func(specName string) tykv1alpha1.ApiDefinition {
		return tykv1alpha1.ApiDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ApiDefinition",
				APIVersion: "tyk.tyk.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: specName,
			},
			Spec: tykv1alpha1.APIDefinitionSpec{},
		}
	}

	apiDefSpecName := "replace-me"
	names = make(map[string]string)

	if dumpAll {
		for i, v := range apiDefSpecList.Apis {
			apiDef := createApiDef(fmt.Sprintf("%s-%d", apiDefSpecName, i))
			apiDef.Spec.APIDefinitionSpec = *v
			names[apiDef.Spec.APIID] = apiDef.ObjectMeta.Name

			if err := e.Encode(&apiDef, bw); err != nil {
				return err
			}

			if _, err := bw.WriteString("\n---\n\n"); err != nil {
				return err
			}
		}

		if err := exportPolicies(policiesList, bw, e); err != nil {
			return err
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

		apiDef := createApiDef(fmt.Sprintf("%s-%d", apiDefSpecName, i))
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

	if err := exportPolicies(policiesList, bw, e); err != nil {
		return err
	}

	return bw.Flush()
}

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
