# Snapshot

snapshot package provides a CLI tool that allows dashboard users to export their 
ApiDefinitions in CRD YAML format that can be used by Tyk Operator. So that it 
facilitates migrations from existing ApiDefinitions to Kubernetes environment.

## Installation

In order to use snapshot utility, you should have access to Tyk Operator repository
on your local machine.

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

3. Run snapshot

```bash
Usage of ./tyk-operator:
  -all
        Dump all APIs
  -category string
        Dump APIs from specified category. (default "operator")
```

- By default, snapshot only dumps ApiDefinitions grouped by a category. The category
can be configured through `--category` flag. The default category is `operator`.
For example, the command below only dumps ApiDefinitions within `k8s` category.
```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> ./tyk-operator --snapshot output.yaml --category k8s
```

- To dump all of your ApiDefinitions, you can pass the `--all` flag.
For example, the command below dumps all ApiDefinitions.
```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> ./tyk-operator --snapshot output.yaml --all
```

## Usage

```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> ./tyk-operator --snapshot output.yaml
```
where

- `<TYK_URL>`: Management URL of your Tyk Dashboard.
- `<TYK_AUTH>`: Operator user API Key.
- `<TYK_ORG>`: Operator user ORG ID.

> For more details, please visit [Tyk Docs](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#tyk-self-managed-hybrid).

This command exports ApiDefinitions grouped by `#operator` category from Dashboard 
to an output file called `output.yaml`.

Output file includes all the ApiDefinitions from `operator` group in CRD YAML format.

You may want to update metadata name of CRs. By default, snapshot creates CRs named 
`replace-me-{i}` where `{i}` increased by each ApiDefinition. You can update 
metadata either manually or by using simple tools.

For example, to change all ApiDefinition CR names to `tyk-operator-api`:
```bash
sed 's/replace-me-/operator-api-/g' output.yaml > apis.yaml
```

Then, you can simply create this new `apis.yaml` file, as follows:
```bash
kubectl apply -f apis.yaml
```

> **Note:** Not all features are supported by Operator. Hence, some configurations would be 
‘lost’ during the conversion. Please visit [ApiDefinition documentation](https://github.com/TykTechnologies/tyk-operator/blob/master/docs/api_definitions.md)
to see supported features.
