Feature: Support gRPC plugins in ApiDefinition custom resource
  In order to customize the middleware chain
  As a developer
  I need to be able to configure an api to use gRPC plugins

  @undone
  Scenario: Pre middleware hook
    Given there is a "TODO: KEYLESS WITH PRE HOOK" resource
    When i request /httpbin/headers endpoint
    Then there should be a 404 http response code
      And the response should match json:
      """
      {
        "todo": "something which shows gRPC plugin short-circuited at the pre hook"
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

  @undone
  Scenario: Post middleware or pre-proxy hook
    Given there is a "keyless" resource
    When i request "/httpbin/get" endpoint with authorization header "SOMESTATICAUTHTOKEN"
    Then there should be a 200 http response code
    And the response should match json:
      """
      {
        "todo": "something which shows gRPC plugin short-circuited at the post hook"
      }
      """
