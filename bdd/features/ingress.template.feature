Feature: Creating IngressClass templates
  In order to use ingress class templates, I should be able to create
  ApiDefinition custom resources with a label "isIngressTemplate": "true"
  The ApiDefinition reconciler should then ignore this resource so that it
  is only stored as a custom resource for later consumption by IngressResources.

  @wip
  Scenario: Create an ApiDefinition ingress template resource
    Given there is a "./custom_resources/ingress_template.yaml" resource
    When i request /httpbin/get endpoint
    Then there should be a 404 http response code
