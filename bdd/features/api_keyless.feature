Feature: Managing REST APIs
  In order to control the Tyk Gateway
  As a developer
  I need to be able to create, update and delete API resources
  And the gateway should reconcile accordingly
  Scenario: Create an api
    Given there is a httpbin api
    When i request /httpbin/get
    Then there should be a 200 response

  Scenario: Update an api
    Given there is a ./config/samples/httpbin.yaml resource
    When i update ./config/samples/httpbin_protected.yaml resource
    When i request /httpbin/get
    Then there should be 401 response
