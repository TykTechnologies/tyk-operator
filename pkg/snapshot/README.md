# Snapshot (WIP)

The snapshot package provides a CLI tool that allows dashboard users to export their 
existing Tyk APIs and Security Policies into CRD YAML format that can be used by Tyk Operator. 

It can help you to migrate existing APIs and Policies to Kubernetes environment.

> **Notes:** After the migration, please be reminded that you should stop using the Dashboard UI to manage the migrated APIs and Policies. Please see [Using Tyk Operator to enable GitOps with Tyk](https://tyk.io/docs/getting-started/key-concepts/gitops-with-tyk/) for more information.

This tool is provided as PoC only and bear some limitations and restrictions. Please use with caution.


---

[Prerequisites](#prerequisites) | [Installation](#installation) | [Preparation](#preparation) | [Usage](#usage) | [Limitations](#limitations)

---


## Prerequisites

1. Access to `tyk-operator` repository.
2. [go](https://go.dev/doc/install)
3. Credentials to connect Tyk Dashboard. Please visit [Tyk Docs](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#tyk-self-managed-hybrid) for details.

## Installation

In order to use snapshot tool, checkout "feat/importer-cli" branch from Tyk Operator repository on your local machine.

1. Clone Tyk Operator repository
```bash
git clone https://github.com/TykTechnologies/tyk-operator.git
cd tyk-operator/
git checkout feat/importer-cli
```

2. Build Tyk Operator
```bash
go build
```

## Preparation
1. Use API Category to group the APIs you want to export

By default, snapshot tool will export all APIs from Tyk dashboard with category `operator`.  You can also categorize your APIs by teams, environments, or any way you want. You can specify which API category to be exported in later steps.

![apis](./img/apis.png)

2. Specify the Kubernetes resource names in Config Data

By default, the created ApiDefinition resource will have the name `replace-me-{i}`. It is not the best name for a kubernetes resource and should be manually updated. You can store this meta-data in Config Data for each API first before exporting, so that you do not need to manually update the output file afterwards.

![config-data](./img/config-data.png)

## Usage
```bash
tyk-operator exports APIs and Security Policies from your Tyk Dashboard to Custom Resource that can be used with Tyk Operator

Export API Definitions:
   --snapshot <output_file>
    	Pull a snapshot of ApiDefinitions from Tyk Dashboard and output as CR
  --all
    	Dump all APIs
  --category <category_name>
    	Dump APIs from specified category (default "operator")

Export Security Policies:
  --policy <output_file>
    	Pull a snapshot of SecurityPolicies from Tyk Dashboard and output as CR
```



### Setting required environment variables

Store the Tyk Dashboard connection parameters in environment variables before running `tyk-operator`, e.g.

```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> \
./tyk-operator --snapshot <OUTPUT_FILE>
```

where
- `<TYK_URL>`: Management URL of your Tyk Dashboard.
- `<TYK_AUTH>`: Operator user API Key.
- `<TYK_ORG>`: Operator user ORG ID.

> For more details on how to obtain the URL and credentials, please visit [Tyk Docs](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#tyk-self-managed-hybrid).



### Exporting API Definitions

#### Specify Category to export

By default, tyk-operator exports ApiDefinitions defined in `#operator` category. You can change default category via `--category` flag.

```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> ./tyk-operator --snapshot output.yaml --category k8s
```

You can also dump all ApiDefinitions from Dashboard via `--all` flag.

```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> ./tyk-operator --snapshot output.yaml --all
```

#### Output CR

`tyk-operator` CLI creates an output file specified via `--snapshot` flag. Each ApiDefinition CR metadata has a default name  `replace-me-{i}` where `{i}` increases by each ApiDefinition.

For example,
```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: replace-me-0 # Default name for ApiDefinition CRs
  namespace: default # Default namespace for ApiDefinition CRs
spec:
  ...
```

In order to specify CR metadata, you can use Config Data. For specified ApiDefinitions,
snapshot CLI generates ApiDefinition CRs based on Config Data of that specific ApiDefinition.

```json
{
  "k8sName": "metadata-name",
  "k8sNs": "metadata-namespace"
}
```

The CLI checks for `k8sName` and `k8sNs` fields of each ApiDefinition's Config Data 
to generate metadata of the output CR. If these fields exist, the CLI uses the 
values specified in these fields. Otherwise, it uses default values (`replace-me-` 
for name and `default` for namespace) for them.

Assume we have the following ApiDefinitions, two of which are categorized as `#testing` 
and created on our Dashboard.

![apis](./img/apis.png)

If we would like to specify metadata of the `test-api-5`, we can update Config Data
of the ApiDefinition, as follows.

![config-data](./img/config-data.png)

So, the generated output for this environment will look as follows;
```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> ./tyk-operator --snapshot output.yaml --category testing
```
```yaml
# output.yaml

apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  creationTimestamp: null
  name: httpbin-api-5   # obtained from Config Data's "k8sName" field.
  namespace: staging    # obtained from Config Data's "k8sNs" field. 
spec:
  name: 'test-api-5 #testing'
  ...
---
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  creationTimestamp: null
  name: replace-me-1    # Since Config Data does not include "k8sName", default name is used.
  namespace: default    # Since Config Data does not include "k8sKey", default namespace is used.
spec:
  name: 'test-api-3 #testing'
  ...
```

### Exporting Security Policies

You can export your SecurityPolicy objects by specifying `--policy` flag.
```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> ./tyk-operator --snapshot output.yaml --policy policies.yaml
```
SecurityPolicy CRs will be saved into a file specified in `--policy` command.

_**Warning:**_ All ApiDefinitions that SecurityPolicy access must exist in Kubernetes.
Otherwise, when you try to apply the policies, SecurityPolicy controller logs an error since it cannot find corresponding ApiDefinition resource in the environment.

### Applying the API and Policy Custom Resources

1. Apply the API Definition CRs first

2. After that, you can apply the Security Policies.

3. The APIs and Policies are now managed by Tyk Operator!

## Limitations
- Not all features are supported by Operator. Non-supported features would be
_lost_ during the conversion. 

> Please visit [ApiDefinition](https://github.com/TykTechnologies/tyk-operator/blob/master/docs/api_definitions.md) and [Policies](https://github.com/TykTechnologies/tyk-operator/blob/master/docs/policies.md) documentations to see supported features.

- The CLI tool will include all fields from your API RAW Definitions (i.e. also the empty and default fields). You can manually clean up those fields if you're sure 
- Please remember that this is a PoC for exporting ApiDefinitions to k8s resources. 
First, try on your testing environment.
