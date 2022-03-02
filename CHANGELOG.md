# Changelog

## [Unreleased](https://github.com/TykTechnologies/tyk-operator/tree/HEAD)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.8.1...HEAD)

**Added:**

- tyk-operator-conf??? [\#362](https://github.com/TykTechnologies/tyk-operator/issues/362)
- Support GraphQL documentation on developer portal catalogue [\#358](https://github.com/TykTechnologies/tyk-operator/issues/358)
- apidef: expose config\_data field to the custom resources [\#356](https://github.com/TykTechnologies/tyk-operator/issues/356)
- document and improve support for prefix and exact path matching [\#209](https://github.com/TykTechnologies/tyk-operator/issues/209)
- apidef: introduce allow\_list, block\_list and ignore\_list [\#92](https://github.com/TykTechnologies/tyk-operator/issues/92)
- mw: JSON Schema Validation support [\#59](https://github.com/TykTechnologies/tyk-operator/issues/59)

**Fixed:**

- deleting an OperatorContext error/requeue if it is referenced by other resources [\#347](https://github.com/TykTechnologies/tyk-operator/issues/347)

**Closed issues:**

- tyk gateway ingress doesn't work [\#365](https://github.com/TykTechnologies/tyk-operator/issues/365)
- aws nlb [\#364](https://github.com/TykTechnologies/tyk-operator/issues/364)
- how can I access to tyk dashboard? Missing documentation. [\#363](https://github.com/TykTechnologies/tyk-operator/issues/363)

## [v0.8.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.8.1) (2021-10-25)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.8.0...v0.8.1)

**Added:**

- Be able to specify different local registry to source kube-rbac-proxy [\#361](https://github.com/TykTechnologies/tyk-operator/issues/361)

## [v0.8.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.8.0) (2021-10-07)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.2...v0.8.0)

**Fixed:**

- 404 not found [\#351](https://github.com/TykTechnologies/tyk-operator/issues/351)
- Error when adding a graphQL API  [\#315](https://github.com/TykTechnologies/tyk-operator/issues/315)
- Failed to sync template - no matches for kind "Ingress" in version "networking.k8s.io/v1beta1" \(K8s v1.22.2\) [\#366](https://github.com/TykTechnologies/tyk-operator/issues/366)
- Operator Pod Fails to come up when Resource Quotas are used in a namespace.  [\#359](https://github.com/TykTechnologies/tyk-operator/issues/359)
- non existent contextRef should return error from reconciler & requeue [\#346](https://github.com/TykTechnologies/tyk-operator/issues/346)

**Closed issues:**

- expose NodePort on kind cluster to bind  admin API  and the gateway in ci [\#330](https://github.com/TykTechnologies/tyk-operator/issues/330)
- Package and publish helm charts [\#321](https://github.com/TykTechnologies/tyk-operator/issues/321)

## [v0.7.2](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.2) (2021-08-16)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.1...v0.7.2)

**Fixed:**

- Operator context not picked up by the operator [\#340](https://github.com/TykTechnologies/tyk-operator/issues/340)
- Security policies not created using Operator Context [\#344](https://github.com/TykTechnologies/tyk-operator/issues/344)

**Closed issues:**

- Issue with Kubernetes Operator / API Definition [\#264](https://github.com/TykTechnologies/tyk-operator/issues/264)

## [v0.7.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.1) (2021-07-29)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.0...v0.7.1)

## [v0.7.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.0) (2021-07-13)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.6.1...v0.7.0)

**Added:**

- chore: Make examples versioned and more accessible [\#313](https://github.com/TykTechnologies/tyk-operator/issues/313)
- Support the custom\_middleware\_bundle field [\#306](https://github.com/TykTechnologies/tyk-operator/issues/306)
- Upgrade operator-sdk to v1.3.0 [\#220](https://github.com/TykTechnologies/tyk-operator/issues/220)
- Switch from vendoring to Go modules [\#197](https://github.com/TykTechnologies/tyk-operator/issues/197)
- feat: upgrade crds and webhooks to v1 [\#124](https://github.com/TykTechnologies/tyk-operator/issues/124)
- CRUD a Tyk Organization decleratively [\#64](https://github.com/TykTechnologies/tyk-operator/issues/64)
- wh: configure webhooks to catch events that occur at API level [\#63](https://github.com/TykTechnologies/tyk-operator/issues/63)
- allow operator to target different gateways [\#322](https://github.com/TykTechnologies/tyk-operator/issues/322)
- Add documentation on how to upgrade the operator with helm [\#293](https://github.com/TykTechnologies/tyk-operator/issues/293)
- Support for updating the API Catalog [\#266](https://github.com/TykTechnologies/tyk-operator/issues/266)

**Fixed:**

- Version apiextensions.k8s.io/v1beta1:CustomResourceDefinition is deprecated [\#273](https://github.com/TykTechnologies/tyk-operator/issues/273)
- track: gateway hello endpoint not always available [\#205](https://github.com/TykTechnologies/tyk-operator/issues/205)
- defect: provide a request size limit middleware example [\#282](https://github.com/TykTechnologies/tyk-operator/issues/282)
- Security Policy "configured" even without changes [\#250](https://github.com/TykTechnologies/tyk-operator/issues/250)

**Closed issues:**

- build and publish latest docker tag upon merge to master [\#251](https://github.com/TykTechnologies/tyk-operator/issues/251)

## [v0.6.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.6.1) (2021-04-22)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.6.0...v0.6.1)

**Added:**

- Does the scope capability in JWT works [\#265](https://github.com/TykTechnologies/tyk-operator/issues/265)
- ingress gateway tls port should be customisable [\#213](https://github.com/TykTechnologies/tyk-operator/issues/213)

**Fixed:**

- deleting an api definition should fail if a security policy grants access to it [\#286](https://github.com/TykTechnologies/tyk-operator/issues/286)
- TLS ingress should not try to open 443 on the Tyk container [\#284](https://github.com/TykTechnologies/tyk-operator/issues/284)

## [v0.6.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.6.0) (2021-04-09)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.5.0...v0.6.0)

**Added:**

- Import existing API definitions and convert to CRDs [\#258](https://github.com/TykTechnologies/tyk-operator/issues/258)
- Use official tyk helm charts for CI [\#268](https://github.com/TykTechnologies/tyk-operator/issues/268)
- feat: url rewrite to internal APIs [\#133](https://github.com/TykTechnologies/tyk-operator/issues/133)

## [v0.5.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.5.0) (2021-02-22)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.4.1...v0.5.0)

**Added:**

- policies \(feat req\): API level limits [\#66](https://github.com/TykTechnologies/tyk-operator/issues/66)
- expose and publish session\_lifetime to the ApiDefinition object [\#254](https://github.com/TykTechnologies/tyk-operator/issues/254)
- docs, examples, validation for JWT auth [\#247](https://github.com/TykTechnologies/tyk-operator/issues/247)
- apidef: introduce method tranform middlewares to the API Definition Object [\#93](https://github.com/TykTechnologies/tyk-operator/issues/93)

**Fixed:**

- JWT default policy key creation issue [\#257](https://github.com/TykTechnologies/tyk-operator/issues/257)
- helm chart hardcoded resources [\#241](https://github.com/TykTechnologies/tyk-operator/issues/241)

**Closed issues:**

- makefile is currently tightly coupled to local kind cluster development [\#203](https://github.com/TykTechnologies/tyk-operator/issues/203)

## [v0.4.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.4.1) (2021-01-06)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.4.0...v0.4.1)

**Added:**

- use controllerutil for security policy reconciliation [\#214](https://github.com/TykTechnologies/tyk-operator/issues/214)
- ingress without template should create simple keyless api definition [\#212](https://github.com/TykTechnologies/tyk-operator/issues/212)
- document and support hostname wildcards [\#210](https://github.com/TykTechnologies/tyk-operator/issues/210)
- modify template api definition should trigger update in ingress resource [\#208](https://github.com/TykTechnologies/tyk-operator/issues/208)

**Fixed:**

- track: deleting API leaves artifacts inside the organization document [\#71](https://github.com/TykTechnologies/tyk-operator/issues/71)
- api ids should be deterministic [\#231](https://github.com/TykTechnologies/tyk-operator/issues/231)
- creating 2 ingress resources results in an api definition being deleted [\#229](https://github.com/TykTechnologies/tyk-operator/issues/229)
- delete template api definition should block if ingress resources depend upon it [\#226](https://github.com/TykTechnologies/tyk-operator/issues/226)
- Security Policy cannot be published to the developer portal [\#223](https://github.com/TykTechnologies/tyk-operator/issues/223)

**Closed issues:**

- ingress controller should be able to listen to custom ingress classes [\#217](https://github.com/TykTechnologies/tyk-operator/issues/217)

## [v0.4.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.4.0) (2020-12-18)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.3.0...v0.4.0)

**Added:**

- feat: cut a release [\#191](https://github.com/TykTechnologies/tyk-operator/issues/191)
- Update tests and ensure everything runs in ci [\#185](https://github.com/TykTechnologies/tyk-operator/issues/185)
- migrate defaulting logic from ApiDefinition reconciler to mutating webhook [\#179](https://github.com/TykTechnologies/tyk-operator/issues/179)
- api: enable detailed recording [\#177](https://github.com/TykTechnologies/tyk-operator/issues/177)
- feat: defaulting webhooks for security policies [\#134](https://github.com/TykTechnologies/tyk-operator/issues/134)
- synchronise certs stored in K8s secrets into the Tyk API Manager [\#105](https://github.com/TykTechnologies/tyk-operator/issues/105)
- research: support ingress resources [\#89](https://github.com/TykTechnologies/tyk-operator/issues/89)

**Fixed:**

- api resource is created even when there is error with universal client [\#186](https://github.com/TykTechnologies/tyk-operator/issues/186)
- track: issues with Policies created through Operator in Tyk Dashboard [\#114](https://github.com/TykTechnologies/tyk-operator/issues/114)
- bug: security policies are not idempotent [\#182](https://github.com/TykTechnologies/tyk-operator/issues/182)

**Closed issues:**

- Test environment [\#116](https://github.com/TykTechnologies/tyk-operator/issues/116)

## [v0.3.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.3.0) (2020-11-27)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.2.0...v0.3.0)

**Added:**

- Update documentation for development environment [\#176](https://github.com/TykTechnologies/tyk-operator/issues/176)
- docs: how to configure JetBrains & VS code intellisense & validation plugins [\#171](https://github.com/TykTechnologies/tyk-operator/issues/171)
- make helm should be interoperable with Mac & Linux [\#170](https://github.com/TykTechnologies/tyk-operator/issues/170)

**Fixed:**

- bug: permissions issue with events [\#166](https://github.com/TykTechnologies/tyk-operator/issues/166)

**Closed issues:**

- ci: update CI to build the operator & install it via helm [\#165](https://github.com/TykTechnologies/tyk-operator/issues/165)

## [v0.2.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.2.0) (2020-11-17)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.1.0...v0.2.0)

**Added:**

- apidef: enable CORS configuration on an API Definition object [\#158](https://github.com/TykTechnologies/tyk-operator/issues/158)
- feat: Continuous delivery github actions [\#152](https://github.com/TykTechnologies/tyk-operator/issues/152)
- docs: active flag in ApiDefinition resource is Pro feature [\#151](https://github.com/TykTechnologies/tyk-operator/issues/151)
- ci: deploy Tyk Pro for CI integration testing [\#68](https://github.com/TykTechnologies/tyk-operator/issues/68)

**Fixed:**

- bug: deleting all apis in the gateway should return 404 not found. [\#148](https://github.com/TykTechnologies/tyk-operator/issues/148)

**Closed issues:**

- research: how to package & deploy the Tyk Operator [\#91](https://github.com/TykTechnologies/tyk-operator/issues/91)

## [v0.1.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.1.0) (2020-11-05)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/374344334c847a2cc3444ec11297674fd236dc8d...v0.1.0)

**Added:**

- feat: support gRPC plugins [\#149](https://github.com/TykTechnologies/tyk-operator/issues/149)
- apidef: introduce udg support [\#98](https://github.com/TykTechnologies/tyk-operator/issues/98)
- apidef: introduce GraphQL proxy support [\#95](https://github.com/TykTechnologies/tyk-operator/issues/95)

**Fixed:**

- bug: unable to deploy operator inside cluster when webhooks enabled [\#90](https://github.com/TykTechnologies/tyk-operator/issues/90)

**Closed issues:**

- Store Mongo IDs for Tyk Pro objects in CR status field [\#81](https://github.com/TykTechnologies/tyk-operator/issues/81)
- Referencing objects that already exist [\#79](https://github.com/TykTechnologies/tyk-operator/issues/79)
- Store dependencies in ConfigMap [\#78](https://github.com/TykTechnologies/tyk-operator/issues/78)
- fr: Webhooks [\#62](https://github.com/TykTechnologies/tyk-operator/issues/62)



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
