# Universal Data Graph (UDG)

> At the moment, Tyk Operator does not support the **V2 GraphQL Engine** introduced in 3.2 Tyk release.

The Universal Data Graph (UDG) lets you combine multiple APIs into one universal interface. With the help of GraphQL you’re able to access multiple APIs with a single query.

You can publish your REST and GraphQL based APIs as GraphQL using Tyk's Universal Data Graph and Tyk Operator.

## Universal Data Graph using Tyk Operator

The following example demonstrates how we can use UDG in the Tyk Operator through the example YAML file created for UDG.
For demonstration purposes, we will use two APIs that are publicly available.
- https://countries.trevorblades.com: is a GraphQL API that returns information about countries.
- https://restcountries.com/v2/alpha: is a REST API that returns information about the given country.

We will combine these two different kinds of APIs into one universal interface with the Universal Data Graph using the Tyk Operator.

Before diving into how we can use UDG in the operator, let's go through some of the concepts that the Universal Data Graph (UDG) introduces.
For more detailed information about the Universal Data Graph concepts, please refer to the official [Tyk Documentation](https://tyk.io/docs/universal-data-graph/udg-concepts/).

### Concepts

1. [Data Sources](https://tyk.io/docs/tyk-stack/universal-data-graph/concepts/datasources/)

    Data sources are responsible for loading the data for a certain field and type. 
   
    For example, assume you would like to return the name and population of the city as a response of your API. 
    You can load name information from one API and population information from another API with the help of Data Sources.

2. [Field Mappings](https://tyk.io/docs/universal-data-graph/concepts/field_mappings/)

    You can enable field mappings between your fields defined in your GraphQL schema and the fields in the JSON response.

3. [Arguments](https://tyk.io/docs/universal-data-graph/concepts/arguments/)

    You can obtain arguments from the GraphQL query using templating syntax. The following example demonstrates the usage of arguments as well.

### Example

Suppose we would like to create a new GraphQL service that exposes a `countries` query responsible for returning information about countries.

```graphql
type Query {
    countries: [Country]
}
```

In order to achieve that, we will use the Countries GraphQL API, which will return the following information about countries: name and code. 
Along with the information returned by the Countries GraphQL API, we also would like to obtain the population, subregion, and alternative spellings of 
countries from the REST Countries API.

So, we can define `Country` type as follows;
```graphql
type Country {
    name: String
    code: String
    restCountry: RestCountry
}

type RestCountry {
    altSpellings: [String]
    subregion: String
    population: String
}
```

Thus, we defined our GraphQL schema. Now, we can jump into the [udg_1.yaml](../config/samples/udg_1.yaml) file to go over the configurations for the Universal Data Graph.

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: udg
spec:
  name: Universal Data Graph Example
  use_keyless: true
  protocol: http
  active: true
  proxy:
    target_url: ""
    listen_path: /udg
    strip_listen_path: true
  graphql:
    enabled: true
    execution_mode: executionEngine
    schema: |
      type Country {
        name: String
        code: String
        restCountry: RestCountry
      }

      type Query {
        countries: [Country]
      }

      type RestCountry {
        altSpellings: [String]
        subregion: String
        population: String
      }
    type_field_configurations:
      - type_name: Query
        field_name: countries
        mapping:
          disabled: false
          path: countries
        data_source:
          kind: GraphQLDataSource
          data_source_config:
            url: "https://countries.trevorblades.com"
            method: POST
            status_code_type_name_mappings: []
      - type_name: Country
        field_name: restCountry
        mapping:
          disabled: true
          path: ""
        data_source:
          kind: HTTPJSONDataSource
          data_source_config:
            url: "https://restcountries.com/v2/alpha/{{ .object.code }}"
            method: GET
            default_type_name: RestCountry
            status_code_type_name_mappings:
              - status_code: 200
    playground:
      enabled: true
      path: /playground
```

1. We start with creating an ApiDefinition called `Universal Data Graph Example` through `spec.name` field. 
2. With the UDG, we may have multiple target URLs that will be composed together to expose a single API. Therefore, we do not need to specify any target URL in the Proxy field. 
Thus, we specify an empty target URL in the `proxy.target_url` field.
3. Then, we start to configure the GraphQL engine. The details of each field can be found on [Tyk Documentation](https://tyk.io/docs/tyk-apis/tyk-gateway-api/api-definition-objects/graphql/).
   1. `spec.graphql.enabled`: We set this value to true. It means the API is a GraphQL API.
   2. `spec.graphql.execution_mode`: We set this to executionEngine. It means that we are configuring our own GraphQL API with multiple data sources.
   3. `spec.graphql.schema`: In schema, we need to specify our GraphQL schema in GraphQL SDL format. Since we designed our API above, we directly set this field to the schema that we designed.
   4. `spec.graphql.type_field_configurations`: A list of configurations used when execution_mode is executionEngine.
      - `type_name`: The type name that we are referring to.
      - `field_name`: The name of the field that the data source applies to. For example, we set `field_name` to `restCountry.` 
      Hence, the `restCountry` field will be loaded from the data source defined under `data_source` of that particular `type_name`.
      - `mapping`: We set this to false because we do not need to remap anything in the response of the data source.
      - `data_source`: Responsible for configuring data source.
        1. `kind`: Kind of the upstream. It can be one of `HTTPJSONDataSource`, `GraphQLDataSource`.
        2. `data_source_config`: Defines details of your data source.
           - `url`: The URL of the upstream data source. We can use Arguments to pass parameters or object details to upstream URL as follows;
            ```yaml
              url: "https://restcountries.com/v2/alpha/{{ .object.code }}"
            ```
           where `{{ .object.code }}` belongs to `code` field of the object which is a type of `Country`.
           - `method`: HTTP request method which the upstream server waits for the url e.g. GET, POST, UPDATE, DELETE.
4. In order to enable the GraphQL playground to test our GraphQL API, we configure `playground` field.
   1. `enabled`: If it is true, it means the playground will be exposed.
   2. `path`: The path of playground. In our example, we set `path` to `/playground`. 
   Since our proxy is configured to listen `/udg` through `proxy.listen_path` field, we can access to the playground by `<TYK GATEWAY ADDRESS>/udg/playground`.

Now, we are ready to apply [udg_1.yaml](../config/samples/udg_1.yaml).
```bash
$ kubectl apply -f config/samples/udg_1.yaml
apidefinition.tyk.tyk.io/udg created
```

In order to verify that our API is created;
```bash
$ kubectl get tykapis
NAME      DOMAIN   LISTENPATH   PROXY.TARGETURL      ENABLED
udg                /udg                              true
```

Now, you can go to the GraphQL Playground (defined under `spec.graphql.playground` of the [udg_1.yaml](../config/samples/udg_1.yaml) file) and test the Universal Data Graph.
