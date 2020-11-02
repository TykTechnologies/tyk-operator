package universal_client

type UniversalClient interface {
	HotReload() error
	Api() UniversalApi
	SecurityPolicy() UniversalSecurityPolicy
	Webhook() UniversalWebhook
	Certificate() UniversalCertificate
}
