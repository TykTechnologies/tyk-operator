Feature: Support gRPC plugins in ApiDefinition custom resource
  In order to customize the middleware chain
  As a developer
  I need to be able to configure an api to use gRPC plugins

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

  @undone
  Scenario: Auth middleware hook unauthenticated
    Given there is a "TODO: gRPC AUTH" resource
    When i request /httpbin/get endpoint
    Then there should be a 401 http response code

  @undone
  Scenario: Auth middleware hook authentic
    Given there is a "TODO: gRPC AUTH" resource
    When i request /httpbin/get endpoint with authorization header "SOMEAUTHTOKEN"
    Then there should be a 200 http response code

  @undone
  Scenario: Auth middleware with ID Extractor
    Given there is a "TODO: gRPC AUTH" resource
    When i request /httpbin/get endpoint with authorization header "SOMESTATICAUTHTOKEN"
      And i request /httpbin/get endpoint with authorization header "SOMESTATICAUTHTOKEN"
    Then there should be a 200 http response code
      And the second response should be quicker than the first response

  @undone
  Scenario: Post auth middleware hook
    Given there is a "AuthToken" resource
    When i request "/httpbin/get" endpoint with authorization header "SOMESTATICAUTHTOKEN"
    Then there should be a 200 http response code
      And the response should match json:
      """
      {
        "todo": "something which shows gRPC plugin short-circuited at the post auth hook"
      }
      """
