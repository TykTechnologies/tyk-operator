Feature: Managing http APIs
    In order to control the Tyk Gateway
    As a developer
    I need to be able to create, update and delete API resources
    And the gateway should reconcile accordingly
  Scenario: Create a keyless api
    Given there is a ./custom_resources/httpbin.keyless.apidefinition.yaml resource
    When i request /httpbin/get endpoint
    Then there should be a 200 http response code

  Scenario: Update an api from keyless to auth token
    Given there is a ./custom_resources/httpbin.keyless.apidefinition.yaml resource
    When i update a ./custom_resources/httpbin.apikey.apidefinition.yaml resource
    And i request /httpbin/get endpoint
    Then there should be a 401 http response code

  Scenario: Delete an API
    Given there is a ./custom_resources/httpbin.keyless.apidefinition.yaml resource
    When i delete a ./custom_resources/httpbin.keyless.apidefinition.yaml resource
    And i request /httpbin/get endpoint
    Then there should be a 404 http response code

  Scenario: Transform xml to json
    Given there is a ./../config/samples/httpbin_transform.yaml resource
    When i request /httpbin-transform/xml endpoint
    Then there should be a "Content-Type: application/json" response header

  Scenario: Blacklist IP address with blacklisted ip
    Given there is a ./custom_resources/ip/blacklist.yaml resource
    When i request /httpbin/headers endpoint with header X-Real-IP: 127.0.0.2
    Then there should be a 403 http response code
    And the response should match JSON:
      """
      {
        "error": "access from this IP has been disallowed"
      }
      """

  Scenario: Blacklist IP address without blacklisted ip
    Given there is a ./custom_resources/ip/blacklist.yaml resource
    When i request /httpbin/get endpoint
    Then there should be a 200 http response code

  Scenario: Whitelist IP address with blocked ip
    Given there is a ./custom_resources/ip/whitelist.yaml resource
    When i request /httpbin/get endpoint
    Then there should be a 403 http response code
    And the response should match JSON:
      """
      {
        "error": "access from this IP has been disallowed"
      }
      """

  Scenario: Whitelist IP address with whitelisted ip
    Given there is a ./custom_resources/ip/whitelist.yaml resource
    When i request /httpbin/headers endpoint with header X-Real-IP: 127.0.0.2
    Then there should be a 200 http response code

  Scenario: Method transform
    Given there is a ./custom_resources/transform/method.yaml resource
    When i request /get endpoint
    Then there should be a 200 http response code
    And the response should contain json key: method value: POST
