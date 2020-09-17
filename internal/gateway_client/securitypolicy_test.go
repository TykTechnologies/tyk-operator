package gateway_client

import (
	"testing"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
)

func TestPol_All(t *testing.T) {
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

	c := getClient()

	newPol := createPolicy()
	_, err := c.SecurityPolicy().Create(newPol)
	if err != nil && err.Error() != "policy id collision detected" {
		t.Fatal(err.Error())
	}

	_ = c.HotReload()

	pol, err := c.SecurityPolicy().Get(newPol.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if pol.ID != newPol.ID {
		t.Fatal("Policy lookup failed.")
	}

}

func TestPol_Create(t *testing.T) {
	c := getClient()
	_ = c.HotReload()
	pols, err := c.SecurityPolicy().All()
	if err != nil {
		t.Fatal(err.Error())
	}

	numPols := len(pols)
	newPol := createPolicy()
	polId, err := c.SecurityPolicy().Create(newPol)

	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("polId: %s", polId)

	_ = c.HotReload()
	newPols, err := c.SecurityPolicy().All()
	if numPols+1 != len(newPols) {
		t.Fatal("Should have 1 more policy")
	}
}

func TestPol_FailsWhenCreatingExistingPolicyID(t *testing.T) {
	c := getClient()
	_ = c.HotReload()

	newPol := createPolicy()
	_, err := c.SecurityPolicy().Create(newPol)
	if err != nil {
		t.Fatal(err.Error())
	}

	_ = c.HotReload()

	newPolTwo := createPolicy()
	_, err = c.SecurityPolicy().Create(newPolTwo)
	if err == nil {
		// error out
		t.Fatal("Should've thrown an error!")
	}
}

func TestPol_Update(t *testing.T) {
	c := getClient()

	newPol := createPolicy()
	_, err := c.SecurityPolicy().Create(newPol)
	if err != nil {
		t.Fatal(err.Error())
	}

	newPol.Rate = 123
	err = c.SecurityPolicy().Update(newPol)
	if err != nil {
		t.Fatal(err.Error())
	}

}

func TestPol_Delete(t *testing.T) {
	c := getClient()
	err := c.HotReload()
	if err != nil {
		t.Fatal(err.Error())
	}

	polId, err := c.SecurityPolicy().Create(createPolicy())
	if err != nil {
		t.Fatal(err.Error())
	}

	_ = c.HotReload()

	err = c.SecurityPolicy().Delete(polId)
	if err != nil {
		t.Fatal(err.Error())
	}

	_ = c.HotReload()
	pols, err := c.SecurityPolicy().All()
	for index := range pols {
		t.Log("index: ", index)
		if pols[index].ID == polId {
			t.Fatal("Should have deleted this policy")
		}
	}
}

func TestPol_DeleteNonexistentPolicy(t *testing.T) {
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
	newPol.Rate = 50
	newPol.Per = 123
	newPol.OrgID = "testorg"
	newPol.ID = "myid"

	return newPol
}
