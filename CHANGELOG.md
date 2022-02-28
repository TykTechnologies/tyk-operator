# Changelog

## [Unreleased](https://github.com/TykTechnologies/tyk-operator/tree/HEAD)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.8.1...HEAD)

**Resolved issues:**

- Support GraphQL documentation on developer portal catalogue [\#358](https://github.com/TykTechnologies/tyk-operator/issues/358)
- apidef: expose config\_data field to the custom resources [\#356](https://github.com/TykTechnologies/tyk-operator/issues/356)
- deleting an OperatorContext error/requeue if it is referenced by other resources [\#347](https://github.com/TykTechnologies/tyk-operator/issues/347)
- document and improve support for prefix and exact path matching [\#209](https://github.com/TykTechnologies/tyk-operator/issues/209)
- apidef: introduce allow_list, block_list and ignore_list [\#92](https://github.com/TykTechnologies/tyk-operator/issues/92)
- mw: JSON Schema Validation support [\#59](https://github.com/TykTechnologies/tyk-operator/issues/59)

**Closed issues:**

- tyk gateway ingress doesn't work [\#365](https://github.com/TykTechnologies/tyk-operator/issues/365)
- aws nlb [\#364](https://github.com/TykTechnologies/tyk-operator/issues/364)
- how can I access to tyk dashboard? Missing documentation. [\#363](https://github.com/TykTechnologies/tyk-operator/issues/363)
- tyk-operator-conf??? [\#362](https://github.com/TykTechnologies/tyk-operator/issues/362)

**Added:**

- Added GitHub Actions to run linters on each PR.
- Integrate Sonar Cloud.
- Support GraphQL documentation on developer portal catalogue.
- Document support for prefix and exact path matching
- Expose `config_data` field to the custom resources
- Add instructions about tyk-operator-conf values
- JSON Schema Validation support
- introduce allow_list, block_list and ignore_list
- Documentation on current UDG support

**Fixed:**

- Deleting an OperatorContext error/requeue if it is referenced by other resources
- Fix validation webhook error for empty target urls

**Merged pull requests:**

- \[TT-4489\] Fix validation webhook error for empty target urls [\#400](https://github.com/TykTechnologies/tyk-operator/pull/400) ([buraksekili](https://github.com/buraksekili))
- Update example YAML files for Ignored and Whitelist plugins [\#399](https://github.com/TykTechnologies/tyk-operator/pull/399) ([buraksekili](https://github.com/buraksekili))
- Update version number of the supported UDG [\#397](https://github.com/TykTechnologies/tyk-operator/pull/397) ([buraksekili](https://github.com/buraksekili))
- \[TT-4410\] Expose code coverage in sonarCloud [\#395](https://github.com/TykTechnologies/tyk-operator/pull/395) ([andrei-tyk](https://github.com/andrei-tyk))
- \[TT-4489\] Documentation on current UDG support [\#394](https://github.com/TykTechnologies/tyk-operator/pull/394) ([buraksekili](https://github.com/buraksekili))
- \[TT-3700\] Add JSON Schema Validation support [\#391](https://github.com/TykTechnologies/tyk-operator/pull/391) ([buraksekili](https://github.com/buraksekili))
- \[TT-3699\] Enable white/black/ingored lists for endpoint designer via operator [\#390](https://github.com/TykTechnologies/tyk-operator/pull/390) ([andrei-tyk](https://github.com/andrei-tyk))
- add a link for each pre-requisite so newbies can go straight to the tool that is required [\#389](https://github.com/TykTechnologies/tyk-operator/pull/389) ([sredxny](https://github.com/sredxny))
- \[TT-3683\] Add a description of how to modify Tyk Operator Configuration [\#387](https://github.com/TykTechnologies/tyk-operator/pull/387) ([buraksekili](https://github.com/buraksekili))
- Update invalid YAML format [\#386](https://github.com/TykTechnologies/tyk-operator/pull/386) ([buraksekili](https://github.com/buraksekili))
- \[TT-3695\] Add details of the access\_rights\_array field in policy migration. [\#385](https://github.com/TykTechnologies/tyk-operator/pull/385) ([buraksekili](https://github.com/buraksekili))
- Show only new lint issues [\#384](https://github.com/TykTechnologies/tyk-operator/pull/384) ([komalsukhani](https://github.com/komalsukhani))
- Update markdown formatting and delete escape char while creating secret. [\#383](https://github.com/TykTechnologies/tyk-operator/pull/383) ([buraksekili](https://github.com/buraksekili))
- \[TT-3673\] Fix operator context issue [\#382](https://github.com/TykTechnologies/tyk-operator/pull/382) ([komalsukhani](https://github.com/komalsukhani))
- \[TT-4306\] Add GitHub Actions to run linters [\#381](https://github.com/TykTechnologies/tyk-operator/pull/381) ([buraksekili](https://github.com/buraksekili))
- \[TT-3678\] Add path-types feature of ingress objects [\#380](https://github.com/TykTechnologies/tyk-operator/pull/380) ([buraksekili](https://github.com/buraksekili))
- Update PR template [\#379](https://github.com/TykTechnologies/tyk-operator/pull/379) ([komalsukhani](https://github.com/komalsukhani))
- Add license check while installing pro [\#378](https://github.com/TykTechnologies/tyk-operator/pull/378) ([komalsukhani](https://github.com/komalsukhani))
- \[TT-3674\] Added support of config\_data in API Definition [\#377](https://github.com/TykTechnologies/tyk-operator/pull/377) ([komalsukhani](https://github.com/komalsukhani))
- \[TT-3798\] Add document on how to publish Graphql API to Portal [\#376](https://github.com/TykTechnologies/tyk-operator/pull/376) ([komalsukhani](https://github.com/komalsukhani))
- \[TT-3673\] Fix operator context issue [\#375](https://github.com/TykTechnologies/tyk-operator/pull/375) ([komalsukhani](https://github.com/komalsukhani))
- Makefile changes [\#373](https://github.com/TykTechnologies/tyk-operator/pull/373) ([komalsukhani](https://github.com/komalsukhani))
- Fix development documentation [\#372](https://github.com/TykTechnologies/tyk-operator/pull/372) ([komalsukhani](https://github.com/komalsukhani))
- docs: fix instructions [\#369](https://github.com/TykTechnologies/tyk-operator/pull/369) ([ermirizio](https://github.com/ermirizio))

## [v0.8.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.8.1) (2021-10-25)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.8.0...v0.8.1)

**Resolved issues:**

- Be able to specify different local registry to source kube-rbac-proxy [\#361](https://github.com/TykTechnologies/tyk-operator/issues/361)

**Merged pull requests:**

- Image spec for rbac proxy custom/private repo [\#368](https://github.com/TykTechnologies/tyk-operator/pull/368) ([cherrymu](https://github.com/cherrymu))

## [v0.8.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.8.0) (2021-10-07)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.2...v0.8.0)

**Resolved issues:**

- make target.Namespace optional [\#331](https://github.com/TykTechnologies/tyk-operator/pull/331) ([gernest](https://github.com/gernest))

**Closed issues:**

- 404 not found [\#351](https://github.com/TykTechnologies/tyk-operator/issues/351)
- Error when adding a graphQL API  [\#315](https://github.com/TykTechnologies/tyk-operator/issues/315)
- Failed to sync template - no matches for kind "Ingress" in version "networking.k8s.io/v1beta1" \(K8s v1.22.2\) [\#366](https://github.com/TykTechnologies/tyk-operator/issues/366)
- Operator Pod Fails to come up when Resource Quotas are used in a namespace.  [\#359](https://github.com/TykTechnologies/tyk-operator/issues/359)
- non existent contextRef should return error from reconciler & requeue [\#346](https://github.com/TykTechnologies/tyk-operator/issues/346)
- expose NodePort on kind cluster to bind  admin API  and the gateway in ci [\#330](https://github.com/TykTechnologies/tyk-operator/issues/330)
- Package and publish helm charts [\#321](https://github.com/TykTechnologies/tyk-operator/issues/321)

**Merged pull requests:**

- use stable ingress on sync api definition templates [\#367](https://github.com/TykTechnologies/tyk-operator/pull/367) ([gernest](https://github.com/gernest))
- Added resources field for kube-rbac-proxy container [\#360](https://github.com/TykTechnologies/tyk-operator/pull/360) ([cherrymu](https://github.com/cherrymu))
- Requeue reconciliation when we can't retrieve contextRef [\#349](https://github.com/TykTechnologies/tyk-operator/pull/349) ([gernest](https://github.com/gernest))
- mapping correct env var for organization [\#348](https://github.com/TykTechnologies/tyk-operator/pull/348) ([asoorm](https://github.com/asoorm))
- Introduce e2e testing framework [\#343](https://github.com/TykTechnologies/tyk-operator/pull/343) ([gernest](https://github.com/gernest))
- Fix operator context permission [\#341](https://github.com/TykTechnologies/tyk-operator/pull/341) ([gernest](https://github.com/gernest))
- docs: add workaround for multi tenancy without  default environment [\#337](https://github.com/TykTechnologies/tyk-operator/pull/337) ([gernest](https://github.com/gernest))
- publishing helm chart repo as gh-pages [\#335](https://github.com/TykTechnologies/tyk-operator/pull/335) ([gernest](https://github.com/gernest))
- remove gh-pages based helm repo [\#371](https://github.com/TykTechnologies/tyk-operator/pull/371) ([gernest](https://github.com/gernest))
- \[TD-370\] Publish tyk-operator to Tyk official Helm repository [\#370](https://github.com/TykTechnologies/tyk-operator/pull/370) ([vesko-tyk](https://github.com/vesko-tyk))
- Reduce docker context by adding .dockerignore [\#353](https://github.com/TykTechnologies/tyk-operator/pull/353) ([gernest](https://github.com/gernest))
- Stable ingress [\#350](https://github.com/TykTechnologies/tyk-operator/pull/350) ([gernest](https://github.com/gernest))
- Remove default context requirement [\#338](https://github.com/TykTechnologies/tyk-operator/pull/338) ([gernest](https://github.com/gernest))
- run automated tests against k8s 1.21 and upgrade older versions [\#334](https://github.com/TykTechnologies/tyk-operator/pull/334) ([asoorm](https://github.com/asoorm))

## [v0.7.2](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.2) (2021-08-16)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.1...v0.7.2)

**Closed issues:**

- Operator context not picked up by the operator [\#340](https://github.com/TykTechnologies/tyk-operator/issues/340)
- Issue with Kubernetes Operator / API Definition [\#264](https://github.com/TykTechnologies/tyk-operator/issues/264)
- Security policies not created using Operator Context [\#344](https://github.com/TykTechnologies/tyk-operator/issues/344)

## [v0.7.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.1) (2021-07-29)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.7.0...v0.7.1)

## [v0.7.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.7.0) (2021-07-13)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.6.1...v0.7.0)

**Resolved issues:**

- chore: Make examples versioned and more accessible [\#313](https://github.com/TykTechnologies/tyk-operator/issues/313)
- Support the custom\_middleware\_bundle field [\#306](https://github.com/TykTechnologies/tyk-operator/issues/306)
- Upgrade operator-sdk to v1.3.0 [\#220](https://github.com/TykTechnologies/tyk-operator/issues/220)
- Switch from vendoring to Go modules [\#197](https://github.com/TykTechnologies/tyk-operator/issues/197)
- feat: upgrade crds and webhooks to v1 [\#124](https://github.com/TykTechnologies/tyk-operator/issues/124)
- CRUD a Tyk Organization decleratively [\#64](https://github.com/TykTechnologies/tyk-operator/issues/64)
- wh: configure webhooks to catch events that occur at API level [\#63](https://github.com/TykTechnologies/tyk-operator/issues/63)
- allow operator to target different gateways [\#322](https://github.com/TykTechnologies/tyk-operator/issues/322)
- Support for updating the API Catalog [\#266](https://github.com/TykTechnologies/tyk-operator/issues/266)

**Closed issues:**

- Version apiextensions.k8s.io/v1beta1:CustomResourceDefinition is deprecated [\#273](https://github.com/TykTechnologies/tyk-operator/issues/273)
- track: gateway hello endpoint not always available [\#205](https://github.com/TykTechnologies/tyk-operator/issues/205)
- Add documentation on how to upgrade the operator with helm [\#293](https://github.com/TykTechnologies/tyk-operator/issues/293)
- defect: provide a request size limit middleware example [\#282](https://github.com/TykTechnologies/tyk-operator/issues/282)
- build and publish latest docker tag upon merge to master [\#251](https://github.com/TykTechnologies/tyk-operator/issues/251)
- Security Policy "configured" even without changes [\#250](https://github.com/TykTechnologies/tyk-operator/issues/250)

**Merged pull requests:**

- add support for portal catalogue [\#329](https://github.com/TykTechnologies/tyk-operator/pull/329) ([gernest](https://github.com/gernest))
- add comment [\#326](https://github.com/TykTechnologies/tyk-operator/pull/326) ([gernest](https://github.com/gernest))
- use same go version for tests and releases [\#325](https://github.com/TykTechnologies/tyk-operator/pull/325) ([gernest](https://github.com/gernest))
- Target multiple gateways in a single operator [\#323](https://github.com/TykTechnologies/tyk-operator/pull/323) ([gernest](https://github.com/gernest))
- consistent Universal.Api interface [\#320](https://github.com/TykTechnologies/tyk-operator/pull/320) ([gernest](https://github.com/gernest))
- add annotation for api protocol [\#319](https://github.com/TykTechnologies/tyk-operator/pull/319) ([gernest](https://github.com/gernest))
- introduce OperatorContext resource [\#318](https://github.com/TykTechnologies/tyk-operator/pull/318) ([gernest](https://github.com/gernest))
- per reconcile loop http api context [\#317](https://github.com/TykTechnologies/tyk-operator/pull/317) ([gernest](https://github.com/gernest))
- ensure SecurityPolicySpec.MID is up to date on Create [\#311](https://github.com/TykTechnologies/tyk-operator/pull/311) ([gernest](https://github.com/gernest))
- chore: preload deployment images for developing with shitty internet [\#310](https://github.com/TykTechnologies/tyk-operator/pull/310) ([gernest](https://github.com/gernest))
- chore: use context.Context in universal client [\#308](https://github.com/TykTechnologies/tyk-operator/pull/308) ([gernest](https://github.com/gernest))
- chore: remove irrelevant comment [\#305](https://github.com/TykTechnologies/tyk-operator/pull/305) ([gernest](https://github.com/gernest))
- fix nightly builds [\#304](https://github.com/TykTechnologies/tyk-operator/pull/304) ([gernest](https://github.com/gernest))
- avoid updating access\_rights\_array on SecurityPolicySpec [\#303](https://github.com/TykTechnologies/tyk-operator/pull/303) ([gernest](https://github.com/gernest))
- add request size middleware example [\#302](https://github.com/TykTechnologies/tyk-operator/pull/302) ([gernest](https://github.com/gernest))
- removing . test [\#301](https://github.com/TykTechnologies/tyk-operator/pull/301) ([asoorm](https://github.com/asoorm))
- test . [\#300](https://github.com/TykTechnologies/tyk-operator/pull/300) ([asoorm](https://github.com/asoorm))
- cleaning up main.go [\#299](https://github.com/TykTechnologies/tyk-operator/pull/299) ([asoorm](https://github.com/asoorm))
- add upgrade documentation [\#298](https://github.com/TykTechnologies/tyk-operator/pull/298) ([gernest](https://github.com/gernest))
- Build nightly release images/artifacts [\#292](https://github.com/TykTechnologies/tyk-operator/pull/292) ([gernest](https://github.com/gernest))
- proper upgrade to kubebuilder v3 [\#316](https://github.com/TykTechnologies/tyk-operator/pull/316) ([gernest](https://github.com/gernest))
- chore: remove lots of dead code [\#312](https://github.com/TykTechnologies/tyk-operator/pull/312) ([gernest](https://github.com/gernest))
- add custom\_middleware\_bundle on APIDefinitionSpec [\#309](https://github.com/TykTechnologies/tyk-operator/pull/309) ([gernest](https://github.com/gernest))
- chore/upgrade operator sdk [\#296](https://github.com/TykTechnologies/tyk-operator/pull/296) ([asoorm](https://github.com/asoorm))

## [v0.6.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.6.1) (2021-04-22)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.6.0...v0.6.1)

**Resolved issues:**

- Does the scope capability in JWT works [\#265](https://github.com/TykTechnologies/tyk-operator/issues/265)
- ingress gateway tls port should be customisable [\#213](https://github.com/TykTechnologies/tyk-operator/issues/213)

**Closed issues:**

- deleting an api definition should fail if a security policy grants access to it [\#286](https://github.com/TykTechnologies/tyk-operator/issues/286)
- TLS ingress should not try to open 443 on the Tyk container [\#284](https://github.com/TykTechnologies/tyk-operator/issues/284)

**Merged pull requests:**

- Change tag to latest [\#294](https://github.com/TykTechnologies/tyk-operator/pull/294) ([joshblakeley](https://github.com/joshblakeley))
-  Configurable ingress port [\#291](https://github.com/TykTechnologies/tyk-operator/pull/291) ([gernest](https://github.com/gernest))
- svg example in the readme [\#288](https://github.com/TykTechnologies/tyk-operator/pull/288) ([asoorm](https://github.com/asoorm))
- properly check for SecurityPolicy resource [\#287](https://github.com/TykTechnologies/tyk-operator/pull/287) ([gernest](https://github.com/gernest))
- Update installation.md [\#283](https://github.com/TykTechnologies/tyk-operator/pull/283) ([gernest](https://github.com/gernest))
- defect: mismatch b64 url encoding vs b64 raw url encoding [\#295](https://github.com/TykTechnologies/tyk-operator/pull/295) ([asoorm](https://github.com/asoorm))
- include k8s v1.20 in matrix tests [\#280](https://github.com/TykTechnologies/tyk-operator/pull/280) ([asoorm](https://github.com/asoorm))

## [v0.6.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.6.0) (2021-04-09)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.5.0...v0.6.0)

**Resolved issues:**

- Import existing API definitions and convert to CRDs [\#258](https://github.com/TykTechnologies/tyk-operator/issues/258)
- Use official tyk helm charts for CI [\#268](https://github.com/TykTechnologies/tyk-operator/issues/268)
- feat: url rewrite to internal APIs [\#133](https://github.com/TykTechnologies/tyk-operator/issues/133)

**Merged pull requests:**

- update looping documentation [\#281](https://github.com/TykTechnologies/tyk-operator/pull/281) ([gernest](https://github.com/gernest))
- simplify initialising client with TYK\_TLS\_INSECURE\_SKIP\_VERIFY [\#278](https://github.com/TykTechnologies/tyk-operator/pull/278) ([gernest](https://github.com/gernest))
- docs\(installation\): update, and clarify some points [\#270](https://github.com/TykTechnologies/tyk-operator/pull/270) ([jlucktay](https://github.com/jlucktay))
- Tyk helm chart [\#269](https://github.com/TykTechnologies/tyk-operator/pull/269) ([gernest](https://github.com/gernest))
- add support for  internal looping API [\#132](https://github.com/TykTechnologies/tyk-operator/pull/132) ([asoorm](https://github.com/asoorm))
- reducing permissions of secret controller [\#276](https://github.com/TykTechnologies/tyk-operator/pull/276) ([asoorm](https://github.com/asoorm))
- support oauth2 client credentials for api protection [\#271](https://github.com/TykTechnologies/tyk-operator/pull/271) ([asoorm](https://github.com/asoorm))
- apply TYK\_TLS\_INSECURE\_SKIP\_VERIFY [\#267](https://github.com/TykTechnologies/tyk-operator/pull/267) ([gernest](https://github.com/gernest))
- Support podAnnotations [\#262](https://github.com/TykTechnologies/tyk-operator/pull/262) ([gernest](https://github.com/gernest))

## [v0.5.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.5.0) (2021-02-22)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.4.1...v0.5.0)

**Resolved issues:**

- policies \(feat req\): API level limits [\#66](https://github.com/TykTechnologies/tyk-operator/issues/66)
- expose and publish session\_lifetime to the ApiDefinition object [\#254](https://github.com/TykTechnologies/tyk-operator/issues/254)
- docs, examples, validation for JWT auth [\#247](https://github.com/TykTechnologies/tyk-operator/issues/247)
- apidef: introduce method tranform middlewares to the API Definition Object [\#93](https://github.com/TykTechnologies/tyk-operator/issues/93)
- feat: enable JWT authentication [\#245](https://github.com/TykTechnologies/tyk-operator/pull/245) ([Jesse0Michael](https://github.com/Jesse0Michael))

**Closed issues:**

- JWT default policy key creation issue [\#257](https://github.com/TykTechnologies/tyk-operator/issues/257)
- helm chart hardcoded resources [\#241](https://github.com/TykTechnologies/tyk-operator/issues/241)
- makefile is currently tightly coupled to local kind cluster development [\#203](https://github.com/TykTechnologies/tyk-operator/issues/203)

**Merged pull requests:**

- add prerequisite for pro deployments [\#259](https://github.com/TykTechnologies/tyk-operator/pull/259) ([gernest](https://github.com/gernest))
- Create LICENSE.md [\#256](https://github.com/TykTechnologies/tyk-operator/pull/256) ([asoorm](https://github.com/asoorm))
- add session\_lifetime [\#255](https://github.com/TykTechnologies/tyk-operator/pull/255) ([gernest](https://github.com/gernest))
- Ensure dev make rules works for kind and minikube [\#253](https://github.com/TykTechnologies/tyk-operator/pull/253) ([gernest](https://github.com/gernest))
- handle float64 field [\#252](https://github.com/TykTechnologies/tyk-operator/pull/252) ([gernest](https://github.com/gernest))
- Cleanup JWT support [\#248](https://github.com/TykTechnologies/tyk-operator/pull/248) ([gernest](https://github.com/gernest))
- jwt helm [\#246](https://github.com/TykTechnologies/tyk-operator/pull/246) ([asoorm](https://github.com/asoorm))
- configurable deployment resources [\#242](https://github.com/TykTechnologies/tyk-operator/pull/242) ([gernest](https://github.com/gernest))
- support method transform middleware [\#240](https://github.com/TykTechnologies/tyk-operator/pull/240) ([gernest](https://github.com/gernest))
- Add templates - GitHub issues and PRs [\#207](https://github.com/TykTechnologies/tyk-operator/pull/207) ([jlucktay](https://github.com/jlucktay))
- feat: support ip whitelisting and blacklisting [\#238](https://github.com/TykTechnologies/tyk-operator/pull/238) ([gernest](https://github.com/gernest))

## [v0.4.1](https://github.com/TykTechnologies/tyk-operator/tree/v0.4.1) (2021-01-06)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.4.0...v0.4.1)

**Resolved issues:**

- use controllerutil for security policy reconciliation [\#214](https://github.com/TykTechnologies/tyk-operator/issues/214)
- ingress without template should create simple keyless api definition [\#212](https://github.com/TykTechnologies/tyk-operator/issues/212)
- document and support hostname wildcards [\#210](https://github.com/TykTechnologies/tyk-operator/issues/210)
- modify template api definition should trigger update in ingress resource [\#208](https://github.com/TykTechnologies/tyk-operator/issues/208)

**Closed issues:**

- track: deleting API leaves artifacts inside the organization document [\#71](https://github.com/TykTechnologies/tyk-operator/issues/71)
- api ids should be deterministic [\#231](https://github.com/TykTechnologies/tyk-operator/issues/231)
- creating 2 ingress resources results in an api definition being deleted [\#229](https://github.com/TykTechnologies/tyk-operator/issues/229)
- delete template api definition should block if ingress resources depend upon it [\#226](https://github.com/TykTechnologies/tyk-operator/issues/226)
- Security Policy cannot be published to the developer portal [\#223](https://github.com/TykTechnologies/tyk-operator/issues/223)
- ingress controller should be able to listen to custom ingress classes [\#217](https://github.com/TykTechnologies/tyk-operator/issues/217)

**Merged pull requests:**

- istio ingress bookinfo example [\#236](https://github.com/TykTechnologies/tyk-operator/pull/236) ([asoorm](https://github.com/asoorm))
- use default template for ingress [\#235](https://github.com/TykTechnologies/tyk-operator/pull/235) ([gernest](https://github.com/gernest))
- deterministic api control [\#234](https://github.com/TykTechnologies/tyk-operator/pull/234) ([gernest](https://github.com/gernest))
- \#210 sample wildcard host ingress [\#233](https://github.com/TykTechnologies/tyk-operator/pull/233) ([asoorm](https://github.com/asoorm))
- Fix regression [\#232](https://github.com/TykTechnologies/tyk-operator/pull/232) ([asoorm](https://github.com/asoorm))
- fix deletion of orphan apis [\#230](https://github.com/TykTechnologies/tyk-operator/pull/230) ([gernest](https://github.com/gernest))
- block template deletion when there are ingresses depending on it. [\#228](https://github.com/TykTechnologies/tyk-operator/pull/228) ([gernest](https://github.com/gernest))
- Ci/hash keys [\#224](https://github.com/TykTechnologies/tyk-operator/pull/224) ([asoorm](https://github.com/asoorm))
- sync apidef templates changes with the ingress controller [\#219](https://github.com/TykTechnologies/tyk-operator/pull/219) ([gernest](https://github.com/gernest))
- Use controllerutil in security policy reconciler [\#215](https://github.com/TykTechnologies/tyk-operator/pull/215) ([gernest](https://github.com/gernest))
- document ingress limitations [\#211](https://github.com/TykTechnologies/tyk-operator/pull/211) ([asoorm](https://github.com/asoorm))
- Asoorm patch 1 [\#227](https://github.com/TykTechnologies/tyk-operator/pull/227) ([asoorm](https://github.com/asoorm))

## [v0.4.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.4.0) (2020-12-18)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.3.0...v0.4.0)

**Resolved issues:**

- feat: cut a release [\#191](https://github.com/TykTechnologies/tyk-operator/issues/191)
- Update tests and ensure everything runs in ci [\#185](https://github.com/TykTechnologies/tyk-operator/issues/185)
- migrate defaulting logic from ApiDefinition reconciler to mutating webhook [\#179](https://github.com/TykTechnologies/tyk-operator/issues/179)
- api: enable detailed recording [\#177](https://github.com/TykTechnologies/tyk-operator/issues/177)
- feat: defaulting webhooks for security policies [\#134](https://github.com/TykTechnologies/tyk-operator/issues/134)
- synchronise certs stored in K8s secrets into the Tyk API Manager [\#105](https://github.com/TykTechnologies/tyk-operator/issues/105)
- research: support ingress resources [\#89](https://github.com/TykTechnologies/tyk-operator/issues/89)

**Closed issues:**

- api resource is created even when there is error with universal client [\#186](https://github.com/TykTechnologies/tyk-operator/issues/186)
- track: issues with Policies created through Operator in Tyk Dashboard [\#114](https://github.com/TykTechnologies/tyk-operator/issues/114)
- bug: security policies are not idempotent [\#182](https://github.com/TykTechnologies/tyk-operator/issues/182)
- Test environment [\#116](https://github.com/TykTechnologies/tyk-operator/issues/116)

**Merged pull requests:**

- enabling strip\_auth\_data and detailed recording in the api definition… [\#206](https://github.com/TykTechnologies/tyk-operator/pull/206) ([asoorm](https://github.com/asoorm))
- Ingress Support [\#201](https://github.com/TykTechnologies/tyk-operator/pull/201) ([asoorm](https://github.com/asoorm))
- Idempotent security policy [\#200](https://github.com/TykTechnologies/tyk-operator/pull/200) ([gernest](https://github.com/gernest))
- Rel 1 [\#195](https://github.com/TykTechnologies/tyk-operator/pull/195) ([asoorm](https://github.com/asoorm))
- Chart.yaml not Charts.yaml [\#194](https://github.com/TykTechnologies/tyk-operator/pull/194) ([alephnull](https://github.com/alephnull))
- Ignore refs to accommodate manual runs [\#193](https://github.com/TykTechnologies/tyk-operator/pull/193) ([alephnull](https://github.com/alephnull))
- Update helm chart when a version is pushed [\#192](https://github.com/TykTechnologies/tyk-operator/pull/192) ([alephnull](https://github.com/alephnull))
- 105 tls certificates [\#189](https://github.com/TykTechnologies/tyk-operator/pull/189) ([asoorm](https://github.com/asoorm))
- ensuring everything is run on ci [\#188](https://github.com/TykTechnologies/tyk-operator/pull/188) ([gernest](https://github.com/gernest))
- scrap make rule surgery [\#187](https://github.com/TykTechnologies/tyk-operator/pull/187) ([gernest](https://github.com/gernest))
- Use Default webhooks  to set default values [\#184](https://github.com/TykTechnologies/tyk-operator/pull/184) ([gernest](https://github.com/gernest))
- Update pro.yaml [\#183](https://github.com/TykTechnologies/tyk-operator/pull/183) ([asoorm](https://github.com/asoorm))

## [v0.3.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.3.0) (2020-11-27)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.2.0...v0.3.0)

**Resolved issues:**

- make helm should be interoperable with Mac & Linux [\#170](https://github.com/TykTechnologies/tyk-operator/issues/170)

**Closed issues:**

- ci: update CI to build the operator & install it via helm [\#165](https://github.com/TykTechnologies/tyk-operator/issues/165)
- Update documentation for development environment [\#176](https://github.com/TykTechnologies/tyk-operator/issues/176)
- docs: how to configure JetBrains & VS code intellisense & validation plugins [\#171](https://github.com/TykTechnologies/tyk-operator/issues/171)
- bug: permissions issue with events [\#166](https://github.com/TykTechnologies/tyk-operator/issues/166)

**Merged pull requests:**

- Update dev workflow with kind [\#181](https://github.com/TykTechnologies/tyk-operator/pull/181) ([gernest](https://github.com/gernest))
- Added IDE Integration section with VS Code support [\#178](https://github.com/TykTechnologies/tyk-operator/pull/178) ([hellobudha](https://github.com/hellobudha))
- fixing flakey tests [\#175](https://github.com/TykTechnologies/tyk-operator/pull/175) ([asoorm](https://github.com/asoorm))
- comment kustomize patches for coversion webhook [\#174](https://github.com/TykTechnologies/tyk-operator/pull/174) ([gernest](https://github.com/gernest))
- replace sed with a go script [\#173](https://github.com/TykTechnologies/tyk-operator/pull/173) ([gernest](https://github.com/gernest))
- validating webhooks for apidefinition resource [\#172](https://github.com/TykTechnologies/tyk-operator/pull/172) ([gernest](https://github.com/gernest))
- docs: update cors feature  status with warning [\#169](https://github.com/TykTechnologies/tyk-operator/pull/169) ([gernest](https://github.com/gernest))
- Feat/105 cert storage [\#168](https://github.com/TykTechnologies/tyk-operator/pull/168) ([asoorm](https://github.com/asoorm))

## [v0.2.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.2.0) (2020-11-17)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/v0.1.0...v0.2.0)

**Resolved issues:**

- apidef: enable CORS configuration on an API Definition object [\#158](https://github.com/TykTechnologies/tyk-operator/issues/158)
- feat: Continuous delivery github actions [\#152](https://github.com/TykTechnologies/tyk-operator/issues/152)
- ci: deploy Tyk Pro for CI integration testing [\#68](https://github.com/TykTechnologies/tyk-operator/issues/68)

**Closed issues:**

- docs: active flag in ApiDefinition resource is Pro feature [\#151](https://github.com/TykTechnologies/tyk-operator/issues/151)
- bug: deleting all apis in the gateway should return 404 not found. [\#148](https://github.com/TykTechnologies/tyk-operator/issues/148)
- research: how to package & deploy the Tyk Operator [\#91](https://github.com/TykTechnologies/tyk-operator/issues/91)

**Merged pull requests:**

- faster build times using cross-build image rather than docker building [\#167](https://github.com/TykTechnologies/tyk-operator/pull/167) ([asoorm](https://github.com/asoorm))
- ability to restrict watched resources by namespace [\#164](https://github.com/TykTechnologies/tyk-operator/pull/164) ([asoorm](https://github.com/asoorm))
- Feat/helm crds [\#163](https://github.com/TykTechnologies/tyk-operator/pull/163) ([asoorm](https://github.com/asoorm))
- add CORS [\#160](https://github.com/TykTechnologies/tyk-operator/pull/160) ([gernest](https://github.com/gernest))
- Bug/flakey bdd [\#157](https://github.com/TykTechnologies/tyk-operator/pull/157) ([asoorm](https://github.com/asoorm))
- Feat/helm [\#156](https://github.com/TykTechnologies/tyk-operator/pull/156) ([asoorm](https://github.com/asoorm))
- docs: api\_definitions document active field [\#153](https://github.com/TykTechnologies/tyk-operator/pull/153) ([gernest](https://github.com/gernest))
- Resilent deploy\_tyk\_pro script [\#120](https://github.com/TykTechnologies/tyk-operator/pull/120) ([tbuchaillot](https://github.com/tbuchaillot))

## [v0.1.0](https://github.com/TykTechnologies/tyk-operator/tree/v0.1.0) (2020-11-05)

[Full Changelog](https://github.com/TykTechnologies/tyk-operator/compare/374344334c847a2cc3444ec11297674fd236dc8d...v0.1.0)

**Resolved issues:**

- feat: support gRPC plugins [\#149](https://github.com/TykTechnologies/tyk-operator/issues/149)
- apidef: introduce udg support [\#98](https://github.com/TykTechnologies/tyk-operator/issues/98)
- apidef: introduce GraphQL proxy support [\#95](https://github.com/TykTechnologies/tyk-operator/issues/95)

**Closed issues:**

- bug: unable to deploy operator inside cluster when webhooks enabled [\#90](https://github.com/TykTechnologies/tyk-operator/issues/90)
- Store Mongo IDs for Tyk Pro objects in CR status field [\#81](https://github.com/TykTechnologies/tyk-operator/issues/81)
- Referencing objects that already exist [\#79](https://github.com/TykTechnologies/tyk-operator/issues/79)
- Store dependencies in ConfigMap [\#78](https://github.com/TykTechnologies/tyk-operator/issues/78)
- fr: Webhooks [\#62](https://github.com/TykTechnologies/tyk-operator/issues/62)

**Merged pull requests:**

- Feat/tags [\#155](https://github.com/TykTechnologies/tyk-operator/pull/155) ([asoorm](https://github.com/asoorm))
- Feat/tags [\#154](https://github.com/TykTechnologies/tyk-operator/pull/154) ([asoorm](https://github.com/asoorm))
- \#149 scenarios for gRPC plugin support [\#150](https://github.com/TykTechnologies/tyk-operator/pull/150) ([asoorm](https://github.com/asoorm))
- completing the tests for XML-\>JSON transform [\#147](https://github.com/TykTechnologies/tyk-operator/pull/147) ([asoorm](https://github.com/asoorm))
- adding feature file for certificates [\#146](https://github.com/TykTechnologies/tyk-operator/pull/146) ([asoorm](https://github.com/asoorm))
- scenario to delete an Api [\#145](https://github.com/TykTechnologies/tyk-operator/pull/145) ([asoorm](https://github.com/asoorm))
- Feat/bdd [\#144](https://github.com/TykTechnologies/tyk-operator/pull/144) ([asoorm](https://github.com/asoorm))
- Feat/bdd [\#143](https://github.com/TykTechnologies/tyk-operator/pull/143) ([asoorm](https://github.com/asoorm))
- Feat/bdd [\#142](https://github.com/TykTechnologies/tyk-operator/pull/142) ([asoorm](https://github.com/asoorm))
- Feat/bdd [\#141](https://github.com/TykTechnologies/tyk-operator/pull/141) ([asoorm](https://github.com/asoorm))
- Feat/bdd [\#140](https://github.com/TykTechnologies/tyk-operator/pull/140) ([asoorm](https://github.com/asoorm))
- feat: add coprocess auth [\#139](https://github.com/TykTechnologies/tyk-operator/pull/139) ([Jesse0Michael](https://github.com/Jesse0Michael))
- fixing issue with requests & sessions [\#138](https://github.com/TykTechnologies/tyk-operator/pull/138) ([asoorm](https://github.com/asoorm))
- adding hiring and installation disclaimer [\#137](https://github.com/TykTechnologies/tyk-operator/pull/137) ([asoorm](https://github.com/asoorm))
- Feat/fix oss gw [\#136](https://github.com/TykTechnologies/tyk-operator/pull/136) ([asoorm](https://github.com/asoorm))
- supporting url rewrite and creating HttpMethod type [\#135](https://github.com/TykTechnologies/tyk-operator/pull/135) ([asoorm](https://github.com/asoorm))
- Update installation.md [\#131](https://github.com/TykTechnologies/tyk-operator/pull/131) ([rewsmith](https://github.com/rewsmith))
- routing & endpoint tracking [\#130](https://github.com/TykTechnologies/tyk-operator/pull/130) ([asoorm](https://github.com/asoorm))
- Comment out organization controller [\#129](https://github.com/TykTechnologies/tyk-operator/pull/129) ([sedkis](https://github.com/sedkis))
- context variables and gateway tags [\#127](https://github.com/TykTechnologies/tyk-operator/pull/127) ([asoorm](https://github.com/asoorm))
- rollback from v1 to v1beta1 [\#125](https://github.com/TykTechnologies/tyk-operator/pull/125) ([asoorm](https://github.com/asoorm))
- Feat/webhook fix [\#123](https://github.com/TykTechnologies/tyk-operator/pull/123) ([asoorm](https://github.com/asoorm))
- installation documentation [\#122](https://github.com/TykTechnologies/tyk-operator/pull/122) ([asoorm](https://github.com/asoorm))
- Feat/env vars operator [\#121](https://github.com/TykTechnologies/tyk-operator/pull/121) ([asoorm](https://github.com/asoorm))
- Feat/docs links examples [\#119](https://github.com/TykTechnologies/tyk-operator/pull/119) ([asoorm](https://github.com/asoorm))
- Feat/docs links examples [\#118](https://github.com/TykTechnologies/tyk-operator/pull/118) ([asoorm](https://github.com/asoorm))
- Fixing links between READMEs [\#115](https://github.com/TykTechnologies/tyk-operator/pull/115) ([sedkis](https://github.com/sedkis))
- Doc Updates [\#113](https://github.com/TykTechnologies/tyk-operator/pull/113) ([sedkis](https://github.com/sedkis))
- Update api\_access.md [\#111](https://github.com/TykTechnologies/tyk-operator/pull/111) ([sedkis](https://github.com/sedkis))
- Update api\_definitions.md [\#110](https://github.com/TykTechnologies/tyk-operator/pull/110) ([sedkis](https://github.com/sedkis))
- Update policies.md [\#109](https://github.com/TykTechnologies/tyk-operator/pull/109) ([sedkis](https://github.com/sedkis))
- Update api\_definitions.md [\#107](https://github.com/TykTechnologies/tyk-operator/pull/107) ([sedkis](https://github.com/sedkis))
- Small changes [\#106](https://github.com/TykTechnologies/tyk-operator/pull/106) ([sedkis](https://github.com/sedkis))
- Idempotent Policy IDs [\#103](https://github.com/TykTechnologies/tyk-operator/pull/103) ([sedkis](https://github.com/sedkis))
- Add ability to reference existing APIs as well as declare new APIs [\#102](https://github.com/TykTechnologies/tyk-operator/pull/102) ([sedkis](https://github.com/sedkis))
- Fix nonexisting apis [\#101](https://github.com/TykTechnologies/tyk-operator/pull/101) ([sedkis](https://github.com/sedkis))
- Update policies.md [\#100](https://github.com/TykTechnologies/tyk-operator/pull/100) ([sedkis](https://github.com/sedkis))
- improving documentation for the api definition custom resource. [\#99](https://github.com/TykTechnologies/tyk-operator/pull/99) ([asoorm](https://github.com/asoorm))
- Feat/api udg [\#97](https://github.com/TykTechnologies/tyk-operator/pull/97) ([asoorm](https://github.com/asoorm))
- Feat/api graphql proxy [\#96](https://github.com/TykTechnologies/tyk-operator/pull/96) ([asoorm](https://github.com/asoorm))
- Fixes webhook manifests generation [\#94](https://github.com/TykTechnologies/tyk-operator/pull/94) ([excieve](https://github.com/excieve))
- update path of tyk.json file to point to correct directory [\#88](https://github.com/TykTechnologies/tyk-operator/pull/88) ([sredxny](https://github.com/sredxny))
- Feat/release [\#87](https://github.com/TykTechnologies/tyk-operator/pull/87) ([asoorm](https://github.com/asoorm))
- Feat/release [\#86](https://github.com/TykTechnologies/tyk-operator/pull/86) ([asoorm](https://github.com/asoorm))
- Feat/release [\#85](https://github.com/TykTechnologies/tyk-operator/pull/85) ([asoorm](https://github.com/asoorm))
- Store ids [\#84](https://github.com/TykTechnologies/tyk-operator/pull/84) ([sedkis](https://github.com/sedkis))
- login to docker when push to master [\#83](https://github.com/TykTechnologies/tyk-operator/pull/83) ([asoorm](https://github.com/asoorm))
- Feat/envtest [\#82](https://github.com/TykTechnologies/tyk-operator/pull/82) ([asoorm](https://github.com/asoorm))
- Feat/admin license secret [\#80](https://github.com/TykTechnologies/tyk-operator/pull/80) ([asoorm](https://github.com/asoorm))
- Feat/bootstrap org test [\#77](https://github.com/TykTechnologies/tyk-operator/pull/77) ([asoorm](https://github.com/asoorm))
- DO NOT MERGE: speed up tests for CI [\#76](https://github.com/TykTechnologies/tyk-operator/pull/76) ([asoorm](https://github.com/asoorm))
- jq wrangling [\#75](https://github.com/TykTechnologies/tyk-operator/pull/75) ([alephnull](https://github.com/alephnull))
- boilerplate for orgs [\#73](https://github.com/TykTechnologies/tyk-operator/pull/73) ([asoorm](https://github.com/asoorm))
- Feat/68 ci cd [\#72](https://github.com/TykTechnologies/tyk-operator/pull/72) ([asoorm](https://github.com/asoorm))
- Feat/webhooks [\#70](https://github.com/TykTechnologies/tyk-operator/pull/70) ([sedkis](https://github.com/sedkis))
- Update policies.md [\#67](https://github.com/TykTechnologies/tyk-operator/pull/67) ([sedkis](https://github.com/sedkis))
- Update api\_definitions.md [\#65](https://github.com/TykTechnologies/tyk-operator/pull/65) ([sedkis](https://github.com/sedkis))
- Update api\_definitions.md [\#61](https://github.com/TykTechnologies/tyk-operator/pull/61) ([sedkis](https://github.com/sedkis))
- enhancing documentation for the api\_definition cr [\#60](https://github.com/TykTechnologies/tyk-operator/pull/60) ([asoorm](https://github.com/asoorm))
- documenting enforced timeout and circuitbreaker support [\#58](https://github.com/TykTechnologies/tyk-operator/pull/58) ([asoorm](https://github.com/asoorm))
- broken validation example [\#57](https://github.com/TykTechnologies/tyk-operator/pull/57) ([asoorm](https://github.com/asoorm))
- Default Rate limit [\#56](https://github.com/TykTechnologies/tyk-operator/pull/56) ([sedkis](https://github.com/sedkis))
- documenting request & response body transforms [\#54](https://github.com/TykTechnologies/tyk-operator/pull/54) ([asoorm](https://github.com/asoorm))
- Docs updates for SecPolicies and other general [\#53](https://github.com/TykTechnologies/tyk-operator/pull/53) ([sedkis](https://github.com/sedkis))
- Feat/global response headers [\#51](https://github.com/TykTechnologies/tyk-operator/pull/51) ([asoorm](https://github.com/asoorm))
- Update middleware.md [\#49](https://github.com/TykTechnologies/tyk-operator/pull/49) ([sedkis](https://github.com/sedkis))
- Add custom plugin example [\#48](https://github.com/TykTechnologies/tyk-operator/pull/48) ([sedkis](https://github.com/sedkis))
- Policy CRUD with updated ID mapping logic [\#47](https://github.com/TykTechnologies/tyk-operator/pull/47) ([sedkis](https://github.com/sedkis))
- Feat/v1alpha1 [\#46](https://github.com/TykTechnologies/tyk-operator/pull/46) ([asoorm](https://github.com/asoorm))
- \[wip\] moving defaulting to admission webhook [\#45](https://github.com/TykTechnologies/tyk-operator/pull/45) ([asoorm](https://github.com/asoorm))
- Feat/apidef generate ingress resource [\#43](https://github.com/TykTechnologies/tyk-operator/pull/43) ([asoorm](https://github.com/asoorm))
- \[wip\] feat/integration tests [\#42](https://github.com/TykTechnologies/tyk-operator/pull/42) ([asoorm](https://github.com/asoorm))
- \[wip\] ci / cd pull request GH Actions [\#41](https://github.com/TykTechnologies/tyk-operator/pull/41) ([asoorm](https://github.com/asoorm))
- securitypolicy controller [\#40](https://github.com/TykTechnologies/tyk-operator/pull/40) ([sedkis](https://github.com/sedkis))
- adding TCP api def example [\#39](https://github.com/TykTechnologies/tyk-operator/pull/39) ([asoorm](https://github.com/asoorm))
- Policies stuff [\#38](https://github.com/TykTechnologies/tyk-operator/pull/38) ([sedkis](https://github.com/sedkis))
- adding TCP api def example [\#37](https://github.com/TykTechnologies/tyk-operator/pull/37) ([asoorm](https://github.com/asoorm))
- streamlining the reconciler. also, we are not using state [\#36](https://github.com/TykTechnologies/tyk-operator/pull/36) ([asoorm](https://github.com/asoorm))
- Feat/transform [\#35](https://github.com/TykTechnologies/tyk-operator/pull/35) ([asoorm](https://github.com/asoorm))
- sample global header middlewares [\#34](https://github.com/TykTechnologies/tyk-operator/pull/34) ([asoorm](https://github.com/asoorm))
- fixing issue with APIDEF policies . char [\#33](https://github.com/TykTechnologies/tyk-operator/pull/33) ([asoorm](https://github.com/asoorm))
- feat/middleware-cache: enabling endpoint caching middleware [\#32](https://github.com/TykTechnologies/tyk-operator/pull/32) ([asoorm](https://github.com/asoorm))
- Friendly ids [\#31](https://github.com/TykTechnologies/tyk-operator/pull/31) ([asoorm](https://github.com/asoorm))
- changing apache2 to mpl2 [\#30](https://github.com/TykTechnologies/tyk-operator/pull/30) ([asoorm](https://github.com/asoorm))
- refactor / cleanup the apidef reconciler [\#29](https://github.com/TykTechnologies/tyk-operator/pull/29) ([asoorm](https://github.com/asoorm))
- removing hardcoded dash / gateway configs from main [\#28](https://github.com/TykTechnologies/tyk-operator/pull/28) ([asoorm](https://github.com/asoorm))
- Feat/policies api improvements [\#27](https://github.com/TykTechnologies/tyk-operator/pull/27) ([asoorm](https://github.com/asoorm))
- Feat/policies api improvements [\#26](https://github.com/TykTechnologies/tyk-operator/pull/26) ([asoorm](https://github.com/asoorm))
- Feat/security policy 2 [\#25](https://github.com/TykTechnologies/tyk-operator/pull/25) ([asoorm](https://github.com/asoorm))
- Feat/security policy 2 [\#24](https://github.com/TykTechnologies/tyk-operator/pull/24) ([asoorm](https://github.com/asoorm))
- some more TODOs [\#23](https://github.com/TykTechnologies/tyk-operator/pull/23) ([asoorm](https://github.com/asoorm))
- starting work on fleshing out a minmum viable security policy [\#22](https://github.com/TykTechnologies/tyk-operator/pull/22) ([asoorm](https://github.com/asoorm))
- \[WIP\] Feat/api auth [\#21](https://github.com/TykTechnologies/tyk-operator/pull/21) ([asoorm](https://github.com/asoorm))
- usage demo [\#20](https://github.com/TykTechnologies/tyk-operator/pull/20) ([asoorm](https://github.com/asoorm))
- Feat/universal client [\#17](https://github.com/TykTechnologies/tyk-operator/pull/17) ([asoorm](https://github.com/asoorm))
- Feat/api id override [\#16](https://github.com/TykTechnologies/tyk-operator/pull/16) ([asoorm](https://github.com/asoorm))
- Add Dashboard API Policies [\#15](https://github.com/TykTechnologies/tyk-operator/pull/15) ([sedkis](https://github.com/sedkis))
- WiP - Add Policy CRD  Operator [\#14](https://github.com/TykTechnologies/tyk-operator/pull/14) ([sedkis](https://github.com/sedkis))
- Feat/dashboard client [\#13](https://github.com/TykTechnologies/tyk-operator/pull/13) ([asoorm](https://github.com/asoorm))
- Webhooks [\#11](https://github.com/TykTechnologies/tyk-operator/pull/11) ([asoorm](https://github.com/asoorm))
- Fix readme [\#10](https://github.com/TykTechnologies/tyk-operator/pull/10) ([sedkis](https://github.com/sedkis))
- Oss gateway [\#9](https://github.com/TykTechnologies/tyk-operator/pull/9) ([asoorm](https://github.com/asoorm))
- CRUD api definitions [\#8](https://github.com/TykTechnologies/tyk-operator/pull/8) ([asoorm](https://github.com/asoorm))
- few tweaks APIDef CRD [\#7](https://github.com/TykTechnologies/tyk-operator/pull/7) ([asoorm](https://github.com/asoorm))
- fixing delete reconcile logic [\#6](https://github.com/TykTechnologies/tyk-operator/pull/6) ([asoorm](https://github.com/asoorm))
- enabling CRUD for APIDef operator [\#5](https://github.com/TykTechnologies/tyk-operator/pull/5) ([asoorm](https://github.com/asoorm))
- Apidef CRDs + Finalizer [\#4](https://github.com/TykTechnologies/tyk-operator/pull/4) ([asoorm](https://github.com/asoorm))
- Feat/gateway client [\#3](https://github.com/TykTechnologies/tyk-operator/pull/3) ([asoorm](https://github.com/asoorm))
- cleanup [\#2](https://github.com/TykTechnologies/tyk-operator/pull/2) ([asoorm](https://github.com/asoorm))
- WIP [\#1](https://github.com/TykTechnologies/tyk-operator/pull/1) ([asoorm](https://github.com/asoorm))



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
