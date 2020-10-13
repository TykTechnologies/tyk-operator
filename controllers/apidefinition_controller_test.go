/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ApiDefinition controller", func() {
	Skip("integration test skipped")

	// Define utility constants for object names.
	const (
		ApiDefinitionName      = "httpbin"
		ApiDefinitionNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Second * 1
	)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here & it's pretty well battle tested.
	Context("HttpBin API Definition", func() {
		It("Should create successfully", func() {
			key := types.NamespacedName{
				Name:      ApiDefinitionName,
				Namespace: ApiDefinitionNamespace,
			}

			created := &v1alpha1.ApiDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.APIDefinitionSpec{
					Name:             "httpbin",
					UseKeylessAccess: true,
					Protocol:         "http",
					Active:           true,
					Proxy: v1alpha1.Proxy{
						TargetURL:       "http://httpbin.org",
						StripListenPath: true,
					},
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			//By("expecting status.apiid to be set")
			//Eventually(func() bool {
			//	found := &v1alpha1.ApiDefinition{}
			//	k8sClient.Get(context.Background(), key, found)
			//	return found.Status.ApiID != ""
			//}, timeout, interval).Should(BeTrue())

			//By("expecting gateway to have loaded the api definition and proxy")
			//Eventually(func() bool {
			//	res, err := http.Get("http://localhost:8080/httpbin/get")
			//	if err != nil {
			//		return false
			//	}
			//	defer res.Body.Close()
			//	if res.StatusCode != http.StatusOK {
			//		return false
			//	}
			//
			//	return true
			//}, timeout, interval).Should(BeTrue())

			// Update enable auth
			updated := &v1alpha1.ApiDefinition{}
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			updated.Spec.UseKeylessAccess = false
			updated.Spec.UseStandardAuth = true
			updated.Spec.AuthConfigs["authToken"] = v1alpha1.AuthConfig{
				AuthHeaderName: "Authorization",
			}
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			//By("expecting unauthorized status code")
			//Eventually(func() bool {
			//	res, err := http.Get("http://localhost:8080/get")
			//	if err != nil {
			//		return false
			//	}
			//	defer res.Body.Close()
			//	if res.StatusCode != http.StatusOK {
			//		return false
			//	}
			//
			//	return true
			//}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				found := &v1alpha1.ApiDefinition{}
				k8sClient.Get(context.Background(), key, found)
				return k8sClient.Delete(context.Background(), found)
			}, timeout, interval).Should(Succeed())

			//Eventually(func() int {
			//	res, err := http.Get("http://localhost:8080/get")
			//	if err != nil {
			//		return 0
			//	}
			//	defer res.Body.Close()
			//	return res.StatusCode
			//}, timeout, interval).Should(Equal(http.StatusNotFound))
		})
	})
})
