package helper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// createK8sClient creates a k8s client to be used in mock clients.
func createK8sClient() (ctrl.Client, error) {
	conf := config.GetConfigOrDie()
	scheme := runtime.NewScheme()

	cl, err := ctrl.New(conf, ctrl.Options{Scheme: scheme})
	if err != nil {
		return nil, nil
	}

	err = v1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, nil
	}

	return cl, err
}

// UpdateSecurityPolicyAnnotations updates each SecurityPolicy objects with "mock_test":"tyk" label.
// It increments Update Count counter by one.
func UpdateSecurityPolicyAnnotations(ctx context.Context) error {
	cl, err := createK8sClient()
	if err != nil {
		return err
	}

	labels := map[string]string{"mock_test": "tyk"}
	policies := v1alpha1.SecurityPolicyList{}

	err = cl.List(ctx, &policies, ctrl.MatchingLabels(labels))
	if err != nil {
		fmt.Printf("failed to list ApiDefinitions, err: %v", err)
		return err
	}

	for _, item := range policies.Items {
		err = updateAnnotations(ctx, cl, &item)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateApiDefinitionAnnotations updates each ApiDefinition objects' with "mock_test":"tyk" label.
// It increments Update Count counter by one.
func UpdateApiDefinitionAnnotations(ctx context.Context) error {
	cl, err := createK8sClient()
	if err != nil {
		return err
	}

	labels := map[string]string{"mock_test": "tyk"}
	apiDefList := v1alpha1.ApiDefinitionList{}

	err = cl.List(ctx, &apiDefList, ctrl.MatchingLabels(labels))
	if err != nil {
		fmt.Printf("failed to list ApiDefinitions, err: %v", err)
		return err
	}

	for _, item := range apiDefList.Items {
		err = updateAnnotations(ctx, cl, &item)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateAnnotations increments the number of Update Request counter stored in given objects' annotations.
func updateAnnotations(ctx context.Context, cl ctrl.Client, object ctrl.Object) error {
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		annotations["mock_test"] = strconv.Itoa(0)
	}

	currUpdateCount, err := strconv.Atoi(annotations["mock_test"])
	if err != nil {
		fmt.Println("failed to convert 'mock_test' annotation to integer, err: ", err)
		return err
	}

	currUpdateCount++
	annotations["mock_test"] = strconv.Itoa(currUpdateCount)
	object.SetAnnotations(annotations)

	err = cl.Update(ctx, object)
	if err != nil {
		fmt.Println("cannot update object, err: ", err)
		return err
	}

	return nil
}
