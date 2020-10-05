package dashboard_client

import (
	"testing"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

var (
	policyNamespacedName = "default/testPolicy"
)

func TestPol_All(t *testing.T) {
	t.SkipNow()
	c := getClient()
	pols, err := c.SecurityPolicy().All()
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, pol := range pols {
		t.Logf("policy ID: %s, aName: %s, ratelimit per: %d per %d, accessRights api name: %s", pol.ID, pol.Name, pol.Rate, pol.Per, pol.AccessRights["41433797848f41a558c1573d3e55a410"].APIName)
	}
}

func TestPol_GetOne(t *testing.T) {
	t.SkipNow()
	c := getClient()

	newPol := createPolicy()
	_, err := c.SecurityPolicy().Create(newPol, policyNamespacedName)
	if err != nil && err.Error() != "policy id collision detected" {
		t.Fatal(err.Error())
	}

	pol, err := c.SecurityPolicy().Get(policyNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	if pol == nil {
		t.Fatal("Policy lookup failed.")
	}

	//cleanup
	err = c.SecurityPolicy().Delete(policyNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, pol not deleted.")
	}
}

func TestPol_Create(t *testing.T) {
	t.SkipNow()
	c := getClient()
	pols, err := c.SecurityPolicy().All()
	if err != nil {
		t.Fatal(err.Error())
	}

	numPols := len(pols)
	newPol := createPolicy()

	_, err = c.SecurityPolicy().Create(newPol, policyNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	newPols, err := c.SecurityPolicy().All()
	if numPols+1 != len(newPols) {
		t.Fatal("Should have 1 more policy")
	}

	//cleanup
	err = c.SecurityPolicy().Delete(policyNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, pol not deleted.")
	}
}

func TestPol_CreateIncludesUniqueTag(t *testing.T) {
	t.SkipNow()
	c := getClient()

	newPol := createPolicy()
	newPol.Tags = append(newPol.Tags, "hello-world", GetPolicyK8SName(policyNamespacedName))
	_, err := c.SecurityPolicy().Create(newPol, policyNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	pol, err := c.SecurityPolicy().Get(policyNamespacedName)
	if pol == nil {
		t.Fatal("Couldn't find policy")
	}

	if pol.Tags[0] != "hello-world" {
		t.Fatal("Deleted an old tag, whoops!")
	}

	if pol.Tags[1] != GetPolicyK8SName(policyNamespacedName) {
		t.Fatal("Didn't add the tag!")
	}

	//cleanup
	err = c.SecurityPolicy().Delete(policyNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, pol not deleted.")
	}
}

func TestPol_FailsWhenCreatingExistingPolicyID(t *testing.T) {
	t.SkipNow()
	c := getClient()

	newPol := createPolicy()
	_, err := c.SecurityPolicy().Create(newPol, policyNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	newPolTwo := createPolicy()
	_, err = c.SecurityPolicy().Create(newPolTwo, policyNamespacedName)
	if err == nil {
		// error out
		t.Fatal("Should've thrown an error!")
	}

	//cleanup
	err = c.SecurityPolicy().Delete(policyNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, pol not deleted.")
	}
}

func TestPol_Update(t *testing.T) {
	t.SkipNow()
	c := getClient()

	newPol := createPolicy()
	newPol.Tags = append(newPol.Tags, "hello-world", GetPolicyK8SName(policyNamespacedName))

	_, err := c.SecurityPolicy().Create(newPol, policyNamespacedName)
	if err != nil && err.Error() != "policy id collision detected" {
		t.Fatal(err.Error())
	}

	newRate := 11

	newPol.Rate = int64(newRate)
	err = c.SecurityPolicy().Update(newPol, policyNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	fetchedPol, err := c.SecurityPolicy().Get(policyNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	if fetchedPol.Rate != int64(newRate) {
		t.Fatal("Did not update the Rate Limit")
	}

	if fetchedPol.Tags[0] != "hello-world" {
		t.Fatal("Deleted an old tag, whoops!")
	}

	if fetchedPol.Tags[1] != GetPolicyK8SName(policyNamespacedName) {
		t.Fatal("Didn't add the tag!")
	}

	//cleanup
	err = c.SecurityPolicy().Delete(policyNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, pol not deleted.")
	}
}

func TestPol_DeleteNonexistentPolicy(t *testing.T) {
	t.SkipNow()
	c := getClient()

	_ = c.HotReload()

	err := c.SecurityPolicy().Delete("fake-pol-id")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func createPolicy() *v1.SecurityPolicySpec {
	newPol := &v1.SecurityPolicySpec{}
	newPol.Name = "my new pol"
	newPol.Rate = 10
	newPol.Per = 60

	newPol.ID = "myid"
	newPol.Active = true
	newPol.AccessRights = make(map[string]v1.AccessDefinition)
	newPol.AccessRights["NEED_A_SENSIBLE_CUSTOM_NAME_HERE"] = v1.AccessDefinition{
		APIName: "my Api",
		APIID:   "NEED_A_SENSIBLE_CUSTOM_NAME_HERE",
	}

	return newPol
}
