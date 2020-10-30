Feature: Managing http APIs
  In order to control the Tyk Gateway
  As a developer
  I need to be able to create, update and delete API resources
  And the gateway should reconcile accordingly
  Scenario: Create a keyless api
    Given there is a ./../config/samples/httpbin.yaml resource
    When i request /httpbin/get endpoint
    Then there should be a 200 http response code

  Scenario: Update an api from keyless to auth token
    Given there is a ./../config/samples/httpbin.yaml resource
    When i update a ./../config/samples/httpbin_protected.yaml resource
      And i request /httpbin/get endpoint
    Then there should be a 401 http response code
