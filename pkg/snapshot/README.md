# Snapshot

snapshot package provides a CLI tool that allows dashboard users to export their 
ApiDefinitions in CRD YAML format that can be used by Tyk Operator. So that it 
facilitates migrations from existing ApiDefinitions to Kubernetes environment.

## Usage

In order to use snapshot utility, you should have access to Tyk Operator repository
on your local machine.

1. Clone Tyk Operator repository
```bash
git clone https://github.com/TykTechnologies/tyk-operator.git
cd tyk-operator/
```

2. Build Tyk Operator
```bash
go build
```

3. Run snapshot
```bash
TYK_MODE=pro TYK_URL=<TYK_URL> TYK_AUTH=<TYK_AUTH> TYK_ORG=<TYK_ORG> tyk-operator --snapshot output.yaml
```
where
- `<TYK_URL>`: Management URL of your Tyk Dashboard.
- `<TYK_AUTH>`: Operator user API Key.
- `<TYK_ORG>`: Operator user ORG ID.

> For more details, please visit [Tyk Docs](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#tyk-self-managed-hybrid).

This command exports all of your ApiDefinitions from Dashboard to an output file
called `output.yaml`.

Output file includes all of your ApiDefinitions in CRD YAML format. The only required
change is .metadata.name field. By default, snapshot creates ApiDefinition CRs 
called `REPLACE_ME_0`, `REPLACE_ME_1` and so on. You need to change these fields
according to your own preference.

For example, to change all ApiDefinition CR names to `tyk-operator-api`,
```bash
sed 's/REPLACE_ME_/operator-api-/g' output.yaml > apis.yaml
```

Then, you can simply create this new `apis.yaml` file, as follows:
```bash
kubectl apply -f apis.yaml
```

> **Note:** Not all features are supported by Operator. Hence, some configurations would be 
‘lost’ during the conversion. Please visit [ApiDefinition documentation](https://github.com/TykTechnologies/tyk-operator/blob/master/docs/api_definitions.md)
to see supported features.
