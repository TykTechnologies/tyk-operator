Feature: Support gRPC plugins in ApiDefinition custom resource
  In order to customize the middleware chain
  As a developer
  I need to be able to configure an api to use gRPC plugins

  @grpc
  Scenario: Pre middleware hook
    Given there is a ./custom_resources/httpbin.keyless.grpc-pre.apidefinition.yaml resource
    When i request /httpbin/headers endpoint
    Then there should be a 200 http response code
      And the response should match JSON:
        """
        {
          "headers": {
            "Accept-Encoding": "gzip",
            "Host": "httpbin.default.svc:8000",
            "Pre": "HelloFromPre",
            "User-Agent": "Go-http-client/1.1"
          }
        }
        """

  @grpc
  Scenario: Post middleware hook
    Given there is a ./custom_resources/httpbin.keyless.grpc-post.apidefinition.yaml resource
    When i request /httpbin/headers endpoint
    Then there should be a 200 http response code
    And the response should match JSON:
      """
      {
        "headers": {
          "Accept-Encoding": "gzip",
          "Host": "httpbin.default.svc:8000",
          "Post": "HelloFromPost",
          "User-Agent": "Go-http-client/1.1"
        }
      }
      """

  @grpc
  Scenario: Auth middleware hook auth missing
    Given there is a ./custom_resources/httpbin.keyless.grpc-auth.apidefinition.yaml resource
    When i request /httpbin/headers endpoint
    Then there should be a 400 http response code

  @grpc
  Scenario: Auth middleware hook authentic
    Given there is a ./custom_resources/httpbin.keyless.grpc-auth.apidefinition.yaml resource
    When i request /httpbin/headers endpoint with header Authorization: foobarbaz
    Then there should be a 200 http response code

  @grpc
  Scenario: Auth middleware with ID Extractor
    Given there is a ./custom_resources/httpbin.keyless.grpc-auth.apidefinition.yaml resource
    When i request /httpbin/headers endpoint with header Authorization: foobarbaz 2 times
    Then there should be a 200 http response code
      And the first response should be slowest

  @grpc @wip
  Scenario: Post auth middleware hook
    Given there is a ./custom_resources/httpbin.keyless.grpc-auth.apidefinition.yaml resource
    When i request /httpbin/headers endpoint with header Authorization: foobarbaz
    Then there should be a 200 http response code
    And the response should match JSON:
      """
      {
        "headers": {
          "Accept-Encoding": "gzip",
          "Host": "httpbin.default.svc:8000",
          "Authorization": "foobarbaz",
          "Postkeyauth": "HelloFromPostKeyAuth",
          "User-Agent": "Go-http-client/1.1"
        }
      }
      """
