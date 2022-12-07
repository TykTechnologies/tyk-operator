Feature: Managing API ownership
    As a developer belonging to a team
    I want to add API Ownership of my team to an API definition
    So that other teams cannot edit my API

  Scenario: Create an api with a operator context and update it with a different operator context with different credentials, the update should not happen
    Given there is a ./custom_resources/httpbin.owned-by-a.apidefinition.yaml resource
    And there is a ./custom_resources/httpbin.owned-by-b.operatorcontext.yaml resource
    When i update a ./custom_resources/httpbin.owned-by-b.apidefinition.yaml resource
    And i request /httpbin-ownership-test-a endpoint
    Then there should be a 200 http response code
