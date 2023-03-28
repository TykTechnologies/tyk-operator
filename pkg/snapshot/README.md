# Snapshot (PoC)

Tyk Operator includes a snapshot package which provides a CLI tool that allows users to export their 
existing Tyk APIs and Security Policies into CRD YAML format that can be used by Tyk Operator. 

It can help you to migrate existing APIs and Policies to Kubernetes environment.

> **Notes:** After the migration, please be reminded that you should stop using 
the Dashboard UI to manage the migrated APIs and Policies. Please see 
[Using Tyk Operator to enable GitOps with Tyk](https://tyk.io/docs/getting-started/key-concepts/gitops-with-tyk/) 
for more information.

This tool is provided as PoC only and bear some limitations and restrictions. 
Please use with caution.

---

[Prerequisites](#prerequisites) | [Preparation](#preparation) | [Installation](#installation) | [Usage](#usage) | [Limitations](#limitations)

---

## Prerequisites

1. Docker or Tyk Operator binaries
2. Credentials to connect Tyk Dashboard or Gateway. Please visit [Tyk Docs](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator) for details.

## Preparation
Prepare the API metadata before exporting.

1. **Specify the Kubernetes resource names in `Config Data` (Required)**

Before exporting the APIs, you must specify the k8s metadata name to be used for each APIDefinition resource. 
It can be saved in `config_data`. You can optionally provide the namespace too.

NOTE: `snapshot` tool will skip exporting any APIs without "k8sName" specified in `config_data`.

Example `config_data` for your APIs:
```json
{
  "k8sName": "httpbin-api-5",
  "k8sNamespace": "staging"
}
```

Corresponding resource will look like this:
```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: httpbin-api-5
  namespace: staging
```

To edit Config Data through API Manager, 
1) Go to API Manager
2) Select the API you want to export
3) Go to "Advanced Options" tab
4) Scroll down to "Config Data" and put in the k8s metadata name and namespace (optional).
![config-data](./img/config-data.png)


2. **Use API Category to group the APIs you want to export (Optional)**

By default, snapshot tool will export **all** APIs that have proper `config_data`
configured. If you do not want to export all APIs, you can use `--category` flag
to select which API category you want to export. 

You can categorize your APIs by teams, environments, or any way 
you want. In the example below, you can see that some APIs are categorized as `#testing` 
or `#production`. 

To set API Category through API Manager, 
1) Go to API Manager
2) Select the API you want to export
3) Under "Core Settings" tab, scroll down to "API Settings".
4) Type the Category name in the input box or select from dropdown list.
5) You will see the Categories of each API on APIs page.
![apis](./img/apis.png)

## Installation
From `v0.14.0`, snapshot tool is available via Docker and Tyk Operator binaries.

### Docker
To install a particular version of tyk-operator via docker image please run the command bellow with the appropriate version you want to use. All available versions could be found on the Tyk Operator Docker Hub page here: https://hub.docker.com/r/tykio/tyk-operator/tags

```bash
docker pull tykio/tyk-operator:{tag}
```

Then the docker image tyk-operator can be used in the following way:

```bash
docker run -it --rm --env-file=.env -v "$(pwd)":/dist tykio/tyk-operator:{tag} [FLAGS]
```

where `.env` file includes Tyk credentials, as follows:

```
TYK_ORG=${TYK_ORG}
TYK_AUTH=${TYK_AUTH}
TYK_URL=${TYK_URL}
TYK_MODE=${TYK_MODE}
```

### Tyk Operator binary

Please visit [GitHub Releases](https://github.com/TykTechnologies/tyk-operator/releases) of Tyk Operator
and install Tyk Operator binary based on your platform.

### Setting required environment variables

In order snapshot tool to connect your Tyk installation, store the Tyk Dashboard 
or Gateway connection parameters in environment variables before running 
`snapshot`.

- `${TYK_MODE}`: `ce` for Tyk Open Source mode and `pro` for for Tyk Self Managed mode.
- `${TYK_URL}`: Management URL of your Tyk Dashboard or Gateway.
- `${TYK_AUTH}`: Operator user API Key.
- `${TYK_ORG}`: Operator user ORG ID.

> For more details on how to obtain the URL and credentials, please visit [Tyk Docs](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#step-3-configuring-tyk-operator).

## Usage
```bash
tyk-operator connects to a Tyk installation, pull a snapshot of APIs and Security Policies from there and output
as Custom Resource that can be used with Tyk Operator.
  
Export API Definitions to a file:
  --apidef <output_file>
    	Pull a snapshot of ApiDefinitions from Tyk and output as CR. All exported ApiDefinitions are included in
       `output_file`.

  --category <category_name>
    	Only export APIs from specified category

Export Security Policies to a file:
  --policy <output_file>
    	Pull a snapshot of all SecurityPolicies from Tyk and output as CR. All exported SecurityPolicies are 
      included in `output_file`.

Export API Definitions and Security Policies into separate files:
  --separate 
        Pull a snapshot of ApiDefinitions and SecurityPolicies from Tyk and output as CR.
        Each ApiDefinition and SecurityPolicy CR will be output into separate files.
        It can be used with --category flag to export only APIs from a category.
```

<!--
### Exporting API Definitions

#### Specify Category to export

By default, `snapshot` tool exports all ApiDefinitions created on the Tyk Dashboard
or Gateway without considering their categories. 

You can specify a category to fetch via `--category` flag, as follows:
```bash
docker run -it --rm --env-file=.env -v "$(pwd)":/dist tykio/tyk-operator:{tag} --apidef output.yaml --category k8s
```
The command above fetches all ApiDefinitions in `#k8s` category.

#### Output CR

`snapshot` tool creates output files specified via `--apidef` flag for ApiDefinitions
and `--policy` for SecurityPolicies. 

In order to specify CR metadata, you can use `Config Data`. For specified ApiDefinitions,
snapshot tool generates ApiDefinition CRs based on `Config Data` of that specific 
ApiDefinition.

```json
{
  "k8sName": "metadata-name",
  "k8sNamespace": "metadata-namespace"
}
```

For example,
```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: production-api  # .metadata.name is obtained through Config Data's 'k8sName' field.
  namespace: production # .metadata.namespace is obtained through Config Data's 'k8sNamespace' field.
spec:
  ...
```

The `snapshot` tool checks for `k8sName` and `k8sNamespace` fields of each
ApiDefinition's Config Data to generate metadata of the output CR. The only required
key for `Config Data` is `k8sName` which specifies your CR's `.metadata.name` field.

> If `k8sNamespace` is not specified, it can be specified via `kubectl apply` as follows:
```bash
kubectl apply -f ${OUTPUT_FILE} -n ${NAMESPACE}
```

<hr/>

Assume we have the following ApiDefinitions, two of which are categorized as `#testing` 
and created on our Dashboard.

![Created APIs on Tyk Dashboard](./img/apis.png)

If we would like to specify metadata of the `test-api-5`, we can update `Config Data`
of the ApiDefinition as follows.

![Config Data feature of ApiDefinition objects](./img/config-data.png)

So, the generated output for this environment will look as follows;
```bash
docker run -it --rm --env-file=.env -v "$(pwd)":/dist tykio/tyk-operator:{tag} --apidef output.yaml --category testing
```
```yaml
# output.yaml

apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  creationTimestamp: null
  name: httpbin-api-5   # obtained from Config Data's "k8sName" field.
  namespace: staging    # obtained from Config Data's "k8sNamespace" field. 
spec:
  name: 'test-api-5 #testing'
  ...
```

**Note:** Since `test-api-3 #testing` API has no valid `Config Data` configured,
which means it lacks of a Config Data with `k8sName` key, the output file does not
contain ApiDefinition Custom Resource for `test-api-3 #testing`.

### Exporting Security Policies

You can export your SecurityPolicy objects by specifying `--policy` flag.
```bash
docker run -it --rm --env-file=.env -v "$(pwd)":/dist tykio/tyk-operator:{tag} --policy policies.yaml
```
SecurityPolicy CRs will be saved into a file specified in `--policy` command.
-->
_**Warning:**_ All ApiDefinitions that SecurityPolicy access must exist in Kubernetes.
Otherwise, when you try to apply the policies, SecurityPolicy controller logs an error 
since it cannot find corresponding ApiDefinition resource in the environment.

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
