package dashboard_client

import (
	"testing"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"

	_ "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

var (
	webhookNamespacedName = "default/webhooktest"
)

func TestWebhook_All(t *testing.T) {
	t.SkipNow()
	c := getClient()
	webhooks, err := c.Webhook().All()
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, webhook := range webhooks {
		t.Logf("webhook id: %s, webhook name: %s, webhook target: %s, webhook method: %s", webhook.ID, webhook.Name, webhook.TargetPath, webhook.Method)
	}
}

func TestWebhook_GetOne_Create_Delete(t *testing.T) {
	t.SkipNow()
	c := getClient()

	newHook := createWebhook()
	err := c.Webhook().Create(webhookNamespacedName, newHook)
	if err != nil {
		t.Fatal(err.Error())
	}

	hook, err := c.Webhook().Get(webhookNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	if hook == nil {
		t.Fatal("Webhook lookup failed.")
	}

	//cleanup
	err = c.Webhook().Delete(webhookNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, webhook not deleted.")
	}
}

func TestWebhook_FailsWhenCreatingExistingWebhookID(t *testing.T) {
	t.SkipNow()
	c := getClient()

	webhook := createWebhook()
	err := c.Webhook().Create(webhookNamespacedName, webhook)
	if err != nil {
		t.Fatal(err.Error())
	}

	webhookTwo := createWebhook()
	err = c.Webhook().Create(webhookNamespacedName, webhookTwo)
	if err == nil {
		// error out
		t.Fatal("Should've thrown an error!")
	}

	//cleanup
	err = c.Webhook().Delete(webhookNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, pol not deleted.")
	}
}

func TestWebhook_Update(t *testing.T) {
	t.SkipNow()
	c := getClient()

	newWebhook := createWebhook()
	err := c.Webhook().Create(webhookNamespacedName, newWebhook)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Do the update
	methodName := v1.WebhookMethod("DELETE")
	newWebhook.Method = methodName
	err = c.Webhook().Update(webhookNamespacedName, newWebhook)
	if err != nil {
		t.Fatal(err.Error())
	}

	fetchedHook, err := c.Webhook().Get(webhookNamespacedName)
	if err != nil {
		t.Fatal(err.Error())
	}

	if fetchedHook.Method != methodName {
		t.Fatal("Did not update the method name")
	}

	//cleanup
	err = c.Webhook().Delete(webhookNamespacedName)
	if err != nil {
		// error out
		t.Fatal("Error cleanup up test, pol not deleted.")
	}
}

func TestWebhook_DeleteNonexistentWebhook(t *testing.T) {
	t.SkipNow()
	c := getClient()

	err := c.Webhook().Delete("fake-web-id")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func createWebhook() *v1.WebhookSpec {
	newHook := &v1.WebhookSpec{}

	newHook.ID = "hello world"
	newHook.Name = "my hook"
	newHook.Method = "GET"
	newHook.TargetPath = "http://foo"

	return newHook
}
