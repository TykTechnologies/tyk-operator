# Changelog

## [Unreleased](https://github.com/TykTechnologies/tyk-operator/tree/HEAD)
[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.10.0...HEAD)

## [v0.10.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.10.0)
[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.9.0...v0.10.0)

**Helm chart**
- Changed default version of operator tag from latest to latest stable release

**Added**
- PoC support of GraphQL Federation on Tyk Operator. It is still WIP.
- Added support of [Global Rate-Limiting](https://github.com/TykTechnologies/tyk-operator/blob/master/config/samples/httpbin_global_rate_limit.yaml) at the API Level.

**Documentation**
- Verified support of Host based routing
- Added GoLand IDE integration

**Changed**
- Makefile: `release` target now replaces operator tag version with the release version in `helm/values.yaml` file.

## [v0.9.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.9.0)
[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.8.2...v0.9.0)

**Breaking Changes:**
- `do_not_track` field's default value is changed from `true` to `false`, to make default behaviour inline with Tyk Dashboard/Gateway.
Therefore, Analytics for API will be enabled by default from this version. A user have to explicitly disable it by setting `do_not_track` field value to `true`.

**Added:**
- Added an [example YAML manifest](./config/samples/httpbin_endpoint_tracking.yaml) for Endpoint Tracking.
- Added Support of Auth Headers while creating GraphQL ProxyOnly API 
- Added [Certificate Pinning](https://tyk.io/docs/security/certificate-pinning/) support in the Tyk Operator [#405](https://github.com/TykTechnologies/tyk-operator/issues/405)
- Added Upstream mTLS gateway parameters that references a secret that contains the upstream certificate

**Documentation**
- Added documentation for GraphQL ProxyOnly API
- Added documentation and examples for using manually uploaded certificates for upstream mTLS

**Fixed:**
- Fixed a bug in which ApiDefinition CRDs were wrongly mutated

**Changed:**
- Installation: Preloading of images during is turned off by default. It can turned on by setting `TYK_OPERATOR_PRELOAD_IMAGES` to true.

## [v0.8.2](https://github.com/TykTechnologies/tyk-operator/tree/v0.8.2) (2022-03-14)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.8.1...v0.8.2)

**Added:**

- Add a description on how to modify Tyk Operator Configuration [#362](https://github.com/TykTechnologies/tyk-operator/issues/362)
- Support GraphQL documentation on developer portal catalogue [\#358](https://github.com/TykTechnologies/tyk-operator/issues/358)
- Expose `config_data` field to the custom resources [\#356](https://github.com/TykTechnologies/tyk-operator/issues/356)
- Document current support of prefix and exact path matching [\#209](https://github.com/TykTechnologies/tyk-operator/issues/209)
- Document allow_list, block_list and ignore_list middleware [\#92](https://github.com/TykTechnologies/tyk-operator/issues/92)
- Added JSON Schema Validation support [\#59](https://github.com/TykTechnologies/tyk-operator/issues/59)
- Document current UDG support 
- Added Getting-Started docs


**Fixed:**

- Deleting an OperatorContext now results in error/requeue if it is referenced by other resources [\#347](https://github.com/TykTechnologies/tyk-operator/issues/347)
- Improved the documentation regarding the version of cert-manager [\#388](https://github.com/TykTechnologies/tyk-operator/issues/388)

**Closed issues:**

- Tyk gateway ingress doesn't work [\#365](https://github.com/TykTechnologies/tyk-operator/issues/365)
- Aws nlb [\#364](https://github.com/TykTechnologies/tyk-operator/issues/364)
- How can I access to tyk dashboard? Missing documentation. [\#363](https://github.com/TykTechnologies/tyk-operator/issues/363)

## [v0.8.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.8.1) (2021-10-25)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.8.0...v0.8.1)

**Added:**

- Support of different local registry specification to source kube-rbac-proxy [\#361](https://github.com/TykTechnologies/tyk-operator/issues/361)

## [v0.8.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.8.0) (2021-10-07)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.2...v0.8.0)

**Added:**

- Package and publish helm charts [\#321](https://github.com/TykTechnologies/tyk-operator/issues/321)

**Changed:**

- Moved to stable ingress `networking.k8s.io/v1` [\#366](https://github.com/TykTechnologies/tyk-operator/issues/366)

**Helm Chart Changes:**

- Can now set `rbac.resources` in values.yaml that will set resources for kube-rbac-proxy container  [\#359](https://github.com/TykTechnologies/tyk-operator/issues/359)


**Fixed:**

- Non existent contextRef should return error from reconciler & requeue [\#346](https://github.com/TykTechnologies/tyk-operator/issues/346)
- expose NodePort on kind cluster to bind  admin API  and the gateway in ci [\#330](https://github.com/TykTechnologies/tyk-operator/issues/330)


**Closed issues:**

- 404 not found [\#351](https://github.com/TykTechnologies/tyk-operator/issues/351)
- Error when adding a graphQL API  [\#315](https://github.com/TykTechnologies/tyk-operator/issues/315)

## [v0.7.2](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.2) (2021-08-16)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.1...v0.7.2)

**Fixed:**


- Security policies not created using Operator Context [\#344](https://github.com/TykTechnologies/tyk-operator/issues/344)

**Closed issues:**

- Issue with Kubernetes Operator / API Definition [\#264](https://github.com/TykTechnologies/tyk-operator/issues/264)

## [v0.7.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.1) (2021-07-29)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.0...v0.7.1)

**Fixed:**

- Operator Context permission [\#340](https://github.com/TykTechnologies/tyk-operator/issues/340)


## [v0.7.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.0) (2021-07-13)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.6.1...v0.7.0)

**Changed:**
- Upgrade operator-sdk to v1.3.0 [\#220](https://github.com/TykTechnologies/tyk-operator/issues/220)
- Switch from vendoring to Go modules [\#197](https://github.com/TykTechnologies/tyk-operator/issues/197)
- Upgraded crds and webhooks to v1 [\#124](https://github.com/TykTechnologies/tyk-operator/issues/124)


**Added:**

- Support the custom\_middleware\_bundle field [\#306](https://github.com/TykTechnologies/tyk-operator/issues/306)
- Allow operator to target different gateways [\#322](https://github.com/TykTechnologies/tyk-operator/issues/322)
- Add documentation on how to upgrade the operator with helm [\#293](https://github.com/TykTechnologies/tyk-operator/issues/293)
- Support for updating the API Catalog [\#266](https://github.com/TykTechnologies/tyk-operator/issues/266)
- build and publish latest docker tag upon merge to master [\#251](https://github.com/TykTechnologies/tyk-operator/issues/251)

**Fixed:**

- Version apiextensions.k8s.io/v1beta1:CustomResourceDefinition is deprecated [\#273](https://github.com/TykTechnologies/tyk-operator/issues/273)
- defect: provide a request size limit middleware example [\#282](https://github.com/TykTechnologies/tyk-operator/issues/282)
- Security Policy "configured" even without changes [\#250](https://github.com/TykTechnologies/tyk-operator/issues/250)

**Closed issues:**

- track: gateway hello endpoint not always available [\#205](https://github.com/TykTechnologies/tyk-operator/issues/205)
- chore: Make examples versioned and more accessible [\#313](https://github.com/TykTechnologies/tyk-operator/issues/313)

## [v0.6.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.6.1) (2021-04-22)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.6.0...v0.6.1)

**Added:**

- Ingress gateway tls port should be customisable [\#213](https://github.com/TykTechnologies/tyk-operator/issues/213)

**Fixed:**

- Deleting an api definition should fail if a security policy grants access to it [\#286](https://github.com/TykTechnologies/tyk-operator/issues/286)
- TLS ingress should not try to open 443 on the Tyk container [\#284](https://github.com/TykTechnologies/tyk-operator/issues/284)

## [v0.6.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.6.0) (2021-04-09)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.5.0...v0.6.0)

**Added:**

- Use official tyk helm charts for CI [\#268](https://github.com/TykTechnologies/tyk-operator/issues/268)
- Feature: URL rewrite to internal APIs [\#133](https://github.com/TykTechnologies/tyk-operator/issues/133)

## [v0.5.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.5.0) (2021-02-22)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.4.1...v0.5.0)

**Added:**

- Expose and publish session\_lifetime to the ApiDefinition object [\#254](https://github.com/TykTechnologies/tyk-operator/issues/254)
- Docs, examples, validation for JWT auth [\#247](https://github.com/TykTechnologies/tyk-operator/issues/247)
- Apidef: introduce method tranform middlewares to the API Definition Object [\#93](https://github.com/TykTechnologies/tyk-operator/issues/93)

**Fixed:**

- Makefile is currently tightly coupled to local kind cluster development [\#203](https://github.com/TykTechnologies/tyk-operator/issues/203)
- JWT default policy key creation issue [\#257](https://github.com/TykTechnologies/tyk-operator/issues/257)
- Helm chart hardcoded resources [\#241](https://github.com/TykTechnologies/tyk-operator/issues/241)


## [v0.4.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.4.1) (2021-01-06)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.4.0...v0.4.1)

**Added:**

- Use controllerutil for security policy reconciliation [\#214](https://github.com/TykTechnologies/tyk-operator/issues/214)
- Ingress without template should create simple keyless api definition [\#212](https://github.com/TykTechnologies/tyk-operator/issues/212)
- Document and support hostname wildcards [\#210](https://github.com/TykTechnologies/tyk-operator/issues/210)
- Modify template api definition should trigger update in ingress resource [\#208](https://github.com/TykTechnologies/tyk-operator/issues/208)
- Ingress controller listens to custom ingress classes [\#217](https://github.com/TykTechnologies/tyk-operator/issues/217)

**Fixed:**

- Deleting API leaves artifacts inside the organization document [\#71](https://github.com/TykTechnologies/tyk-operator/issues/71)
- API ids should be deterministic [\#231](https://github.com/TykTechnologies/tyk-operator/issues/231)
- Creating 2 ingress resources results in an api definition being deleted [\#229](https://github.com/TykTechnologies/tyk-operator/issues/229)
- Delete template api definition should block if ingress resources depend upon it [\#226](https://github.com/TykTechnologies/tyk-operator/issues/226)


## [v0.4.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.4.0) (2020-12-18)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.3.0...v0.4.0)

**Added:**

- Update tests and ensure everything runs in ci [\#185](https://github.com/TykTechnologies/tyk-operator/issues/185)
- Migrate defaulting logic from ApiDefinition reconciler to mutating webhook [\#179](https://github.com/TykTechnologies/tyk-operator/issues/179)
- API: enable detailed recording [\#177](https://github.com/TykTechnologies/tyk-operator/issues/177)
- Feature: Defaulting webhooks for security policies [\#134](https://github.com/TykTechnologies/tyk-operator/issues/134)
- Synchronise certs stored in K8s secrets into the Tyk API Manager [\#105](https://github.com/TykTechnologies/tyk-operator/issues/105)
- Support ingress resources [\#89](https://github.com/TykTechnologies/tyk-operator/issues/89)
- Test environment [\#116](https://github.com/TykTechnologies/tyk-operator/issues/116)

**Fixed:**

- API resource is created even when there is error with universal client [\#186](https://github.com/TykTechnologies/tyk-operator/issues/186)
- Security policies are not idempotent [\#182](https://github.com/TykTechnologies/tyk-operator/issues/182)

**Closed:**

- Track: issues with Policies created through Operator in Tyk Dashboard [\#114](https://github.com/TykTechnologies/tyk-operator/issues/114)



## [v0.3.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.3.0) (2020-11-27)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.2.0...v0.3.0)

**Added:**

- Update documentation for development environment [\#176](https://github.com/TykTechnologies/tyk-operator/issues/176)
- Docs: How to configure JetBrains & VS code intellisense & validation plugins [\#171](https://github.com/TykTechnologies/tyk-operator/issues/171)
- make helm should be interoperable with Mac & Linux [\#170](https://github.com/TykTechnologies/tyk-operator/issues/170)
- ci: update CI to build the operator & install it via helm [\#165](https://github.com/TykTechnologies/tyk-operator/issues/165)

**Closed issues:**

- Permissions issue with events [\#166](https://github.com/TykTechnologies/tyk-operator/issues/166)

## [v0.2.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.2.0) (2020-11-17)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.1.0...v0.2.0)

**Added:**

- APIDef: enable CORS configuration on an API Definition object [\#158](https://github.com/TykTechnologies/tyk-operator/issues/158)
- Feature: Continuous delivery github actions [\#152](https://github.com/TykTechnologies/tyk-operator/issues/152)
- Docs: Active flag in ApiDefinition resource is Pro feature [\#151](https://github.com/TykTechnologies/tyk-operator/issues/151)
- ci: Deploy Tyk Pro for CI integration testing [\#68](https://github.com/TykTechnologies/tyk-operator/issues/68)


**Closed issues:**

- Research: How to package & deploy the Tyk Operator [\#91](https://github.com/TykTechnologies/tyk-operator/issues/91)

## [v0.1.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.1.0) (2020-11-05)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/374344334c847a2cc3444ec11297674fd236dc8d...v0.1.0)

**Added:**

- Feature: Support gRPC plugins [\#149](https://github.com/TykTechnologies/tyk-operator/issues/149)
- APIDef: Introduce udg support [\#98](https://github.com/TykTechnologies/tyk-operator/issues/98)
- APIDef: Introduce GraphQL proxy support [\#95](https://github.com/TykTechnologies/tyk-operator/issues/95)

**Fixed:**

- Unable to deploy operator inside cluster when webhooks enabled [\#90](https://github.com/TykTechnologies/tyk-operator/issues/90)

**Closed issues:**

- Store Mongo IDs for Tyk Pro objects in CR status field [\#81](https://github.com/TykTechnologies/tyk-operator/issues/81)
- Referencing objects that already exist [\#79](https://github.com/TykTechnologies/tyk-operator/issues/79)
- Store dependencies in ConfigMap [\#78](https://github.com/TykTechnologies/tyk-operator/issues/78)
- fr: Webhooks [\#62](https://github.com/TykTechnologies/tyk-operator/issues/62)



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
