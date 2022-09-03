# Client Mutual TLS

Tyk supports mTLS between gateway and the client. You can read more about it [here](https://tyk.io/docs/nightly/basic-config-and-security/security/mutual-tls/client-mtls)

> *Note:*  Operator only support `Static` client mTLS flow.

mTLS Auth can be enabled by setting `use_mutual_tls_auth` to true.

You can provide list of client certificates in two ways:
- *Using exisiting Certificates:*

  If you have already uploaded certificate on Tyk Gateway, you can specify certificate IDs using this field.

  You can find sample manifests [here](./../../config/samples/mtls/client/httpbin_client_mtls_using_certids.yaml)

- *Using secrets:*

  You can store certificate in secret and provide it reference in `client_certificate_refs` field.
  Operator will upload this certificate to Tyk and get it's certificate ID. 
  > *Note:* Secrets and API should be created in same namespace.

  You can find sample manifests [here](./../../config/samples/mtls/client/httpbin_client_mtls_using_secret.yaml)
