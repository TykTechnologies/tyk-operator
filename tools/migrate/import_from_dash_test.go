package migrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type ApisResponse struct {
	Apis  []DashboardApi `json:"apis"`
	Pages int            `json:"pages"`
}

type DashboardApi struct {
	CreatedAt     string               `json:"created_at,omitempty"`
	ApiDefinition v1.APIDefinitionSpec `json:"api_definition"`
}

func TestBuild(t *testing.T) {
	for i := 0; i < 1; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			a, err := readAPI(0)
			if err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			err = Build(&buf, "default", a, []v1.SecurityPolicySpec{})
			if err != nil {
				t.Fatal(err)
			}
			r, err := readResource(i)
			if err != nil {
				t.Fatal(err)
			}
			ioutil.WriteFile(fmt.Sprintf("samples/%d_apis.yaml", i), buf.Bytes(), 0600)
			if !bytes.Equal(buf.Bytes(), r) {
				t.Fatal("Incorrect resource generated")
			}
		})
	}
}

func readResource(n int) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("samples/%d_apis.yaml", n))
}

func readAPI(n int) ([]v1.APIDefinitionSpec, error) {
	f, err := os.Open(fmt.Sprintf("samples/%d_apis.json", n))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var o ApisResponse
	err = json.NewDecoder(f).Decode(&o)
	if err != nil {
		return nil, err
	}
	var a []v1.APIDefinitionSpec
	for _, v := range o.Apis {
		a = append(a, v.ApiDefinition)
	}
	return a, nil
}
