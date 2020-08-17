


operator-sdk init --domain=tyk.io --repo=github.com/TykTechnologies/tyk-operator
operator-sdk create api --group tyk --version v1 --kind Gateway --resource=true --controller=true


After modifying the *_types.go file always run the following command to update the generated code for that resource type:
```
make generate
```


Generating CRD manifests

Once the API is defined with spec/status fields and CRD validation markers, the CRD manifests can be generated and updated with the following command:

```
make manifests
```

This makefile target will invoke controller-gen to generate the CRD manifests at config/crd/bases/cache.example.com_memcacheds.yaml


Register the CRD

```
make install
```

