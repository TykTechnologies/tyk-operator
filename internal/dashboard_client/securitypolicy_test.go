package dashboard_client

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
	pols, err := c.SecurityPolicy().All()
	if err != nil {
		t.Fatal(err.Error())
	}

	numPols := len(pols)
	newPol := createPolicy()
	_, err = c.SecurityPolicy().Create(newPol)
	if err != nil {
		t.Fatal(err.Error())
	}

	newPols, err := c.SecurityPolicy().All()
	if numPols+1 != len(newPols) {
		t.Fatal("Should have 1 more policy")
	}
}

func TestPol_FailsWhenCreatingExistingPolicyID(t *testing.T) {
	c := getClient()

	newPol := createPolicy()
	_, err := c.SecurityPolicy().Create(newPol)
	if err != nil {
		t.Fatal(err.Error())
	}

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
	if err != nil && err.Error() != "policy id collision detected" {
		t.Fatal(err.Error())
	}

	newRate := 11

	newPol.Rate = int64(newRate)
	err = c.SecurityPolicy().Update(newPol)
	if err != nil {
		t.Fatal(err.Error())
	}

	fetchedPol, err := c.SecurityPolicy().Get(newPol.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if fetchedPol.Rate != int64(newRate) {
		t.Fatal("Did not update the Rate Limit")
	}

}

func TestPol_Delete(t *testing.T) {
	c := getClient()

	_, err := c.SecurityPolicy().Create(createPolicy())
	if err != nil {
		t.Fatal(err.Error())
	}

	//err = c.SecurityPolicy.Delete(polId)
	//if err != nil {
	//	t.Fatal(err.Error())
	//}

	//pols, err := c.SecurityPolicy.SecurityPolicy.All()
	//for index := range pols {
	//	t.Log("index: ", index)
	//	if pols[index].ID == polId {
	//		t.Fatal("Should have deleted this policy")
	//	}
	//}
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
	newPol.Rate = 10
	newPol.Per = 60
	newPol.OrgID = "5e9d9544a1dcd60001d0ed20"
	newPol.ID = "myid"
	newPol.Active = true
	newPol.AccessRights = make(map[string]v1.AccessDefinition)
	newPol.AccessRights["NEED_A_SENSIBLE_CUSTOM_NAME_HERE"] = v1.AccessDefinition{
		APIName: "my Api",
		APIID:   "NEED_A_SENSIBLE_CUSTOM_NAME_HERE",
	}

	return newPol
}
