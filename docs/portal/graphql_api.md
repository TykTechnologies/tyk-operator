## Publish GraphQL API to Portal

This example publish GraphQL API on Tyk Portal.

### 1. Create GraphQL API
GraphQL API can be deployed by setting `graphql` field of API Definitions.
     
It has following mandatory subfields
* **enabled:** Enabled indicates if GraphQL proxy should be enabled.
* **execution_mode:** The mode of a GraphQL API. There are two types: `proxyOnly` and `executionEngine`.
    - **proxyOnly:** There is one single upstream which is a GraphQL API and Tyk proxies it.
    - **executionEngine:** It lets you to configure your own GraphQL API with multiple data sources. This means that you will compose your own schema.
* **schema:** The GraphQL schema of your API is saved in this variable in SDL format.

#### Example
Here is an example which creates GraphQL API which proxies request to [https://countries.trevorblades.com/](https://countries.trevorblades.com/) and has standard auth type.

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
 name: graph-test
spec:
 name: graph-test
 graphql:
  enabled: true
  execution_mode: proxyOnly
  schema: "directive @cacheControl(maxAge: Int, scope: CacheControlScope) on FIELD_DEFINITION | OBJECT | INTERFACE\n\nenum CacheControlScope {\n  PUBLIC\n  PRIVATE\n}\n\ntype Continent {\n  code: ID!\n  name: String!\n  countries: [Country!]!\n}\n\ninput ContinentFilterInput {\n  code: StringQueryOperatorInput\n}\n\ntype Country {\n  code: ID!\n  name: String!\n  native: String!\n  phone: String!\n  continent: Continent!\n  capital: String\n  currency: String\n  languages: [Language!]!\n  emoji: String!\n  emojiU: String!\n  states: [State!]!\n}\n\ninput CountryFilterInput {\n  code: StringQueryOperatorInput\n  currency: StringQueryOperatorInput\n  continent: StringQueryOperatorInput\n}\n\ntype Language {\n  code: ID!\n  name: String\n  native: String\n  rtl: Boolean!\n}\n\ninput LanguageFilterInput {\n  code: StringQueryOperatorInput\n}\n\ntype Query {\n  continents(filter: ContinentFilterInput): [Continent!]!\n  continent(code: ID!): Continent\n  countries(filter: CountryFilterInput): [Country!]!\n  country(code: ID!): Country\n  languages(filter: LanguageFilterInput): [Language!]!\n  language(code: ID!): Language\n}\n\ntype State {\n  code: String\n  name: String!\n  country: Country!\n}\n\ninput StringQueryOperatorInput {\n  eq: String\n  ne: String\n  in: [String]\n  nin: [String]\n  regex: String\n  glob: String\n}\n\n\"\"\"The `Upload` scalar type represents a file upload.\"\"\"\nscalar Upload\n"
 protocol: http
 active: true
 proxy:
  listen_path: /graph-test
  target_url: https://countries.trevorblades.com/
  strip_listen_path: true
 use_keyless: false
 use_standard_auth: true   
```

### 2. Create Security Policy
Create a security policy for publishing the above API to the portal.

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: SecurityPolicy
metadata:
 name: graph-pol
spec:
 name: graph-pol
 active: true
 state: active
 access_rights_array:
 - name: graph-test
   namespace: default
   versions:
     - Default
```

`access_rights_array` contains details of API created in previous step.


### 3. Create API Description
Create API Description object, which will be used for creating portal catalogue resource in next step.

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: APIDescription
metadata:
 name: test-graph
spec:
 name: Test graphQL API
 policyRef:
  name: graph-pol
  namespace: default
 docs: 
  doc_type: graphql
 show: true
 version: v2
```

`policyRef` contains details of security policy created in previous step.

### 4. Create Portal Catalogue
Create a portal catalogue.

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: PortalAPICatalogue
metadata:
 name: test-graph-cat
spec:
 apis:
 - apiDescriptionRef:
    name: test-graph
    namespace: default
```

Your API is now published to your Portal!!



## Protected GraphQL Catalogue

If you have a protected API, your users won’t be able to inspect the GraphQL schema or make API calls until they add their API Key to the Headers section.



## CORS

You may have to enable the following CORS settings to allow your consumers to access the GraphQL Playground.
This could be done by setting following fields in API Definition

```yaml
CORS:
    enable: true
    allowed_headers:
      - "Authorization"
      - "Content-Type"
```
