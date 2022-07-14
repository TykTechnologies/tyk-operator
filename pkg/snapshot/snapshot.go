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

func PrintSnapshot(ctx context.Context, fileName string) {
	apiDefSpecList, err := klient.Universal.Api().List(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	f, err := os.Create(fileName)
	defer f.Close()

	bw := bufio.NewWriter(f)

	for i, v := range apiDefSpecList.Apis {

		e := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		})

		apiDef := tykv1alpha1.ApiDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ApiDefinition",
				APIVersion: "tyk.tyk.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("REPLACE_ME_%d", i),
			},
			Spec: tykv1alpha1.APIDefinitionSpec{},
		}
		apiDef.Spec.APIDefinitionSpec = *v

		e.Encode(&apiDef, bw)

		bw.Write([]byte("\n---\n\n"))
	}

	bw.Flush()
}
