package snapshot

import (
	"bufio"
	"context"
	"fmt"
	"os"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

const (
	k8sKey = "k8sName"
)

func PrintSnapshot(ctx context.Context, fileName string, dumpAll bool) error {
	apiDefSpecList, err := klient.Universal.Api().List(ctx)
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

	if dumpAll {
		for i, v := range apiDefSpecList.Apis {
			apiDef := createApiDef(fmt.Sprintf("%s-%d", apiDefSpecName, i))
			apiDef.Spec.APIDefinitionSpec = *v

			if err := e.Encode(&apiDef, bw); err != nil {
				return err
			}

			if _, err := bw.WriteString("\n---\n\n"); err != nil {
				return err
			}
		}

		return bw.Flush()
	}

	for _, v := range apiDefSpecList.Apis {
		if v.ConfigData != nil {
			val, ok := v.ConfigData.Object[k8sKey]
			if ok {
				if s, ok := val.(string); ok {
					apiDefSpecName = s
				}

				apiDef := createApiDef(apiDefSpecName)
				apiDef.Spec.APIDefinitionSpec = *v

				if err := e.Encode(&apiDef, bw); err != nil {
					return err
				}

				if _, err := bw.WriteString("\n---\n\n"); err != nil {
					return err
				}
			}
		}
	}

	return bw.Flush()
}
