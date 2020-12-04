Feature: Managing https APIs
  In order to terminate TLS, Tyk operator
  must be able to load, rotate and delete
  TLS certificates stored as secrets
  into Tyk's certificate manager.

  @undone
  Scenario: Ignore certificates which have not been referenced by an API Definition
    Given there is a "self-signed-issuer" resource
    And there is a "certificate" resource
    Then the admin "/certs" should not be created

  @undone
  Scenario: Create a certificate
    Given there is a "self-signed-issuer" resource
      And there is a "certificate" resource
      And there is a "https" resource
    Then the admin "/certs/ID" should be created
      And the "https" resource should reference the new certificate

  @undone
  Scenario: Rotate/Renew a certificate
    Given there is a "self-signed-issuer" resource
      And there is a "certificate" resource
      And there is a "https" resource
    When i kubectl cert-manager renew
    Then the admin "/certs/ID" should be created
      And the "https" resource should reference the new certificate

  @undone
  Scenario: Delete a certificate
    Given there is a "self-signed-issuer" resource
      And there is a "certificate" resource
      And there is a "https" resource
    When i delete the "certificate" resource
      And i delete the "secret" resource
    Then the admin "/certs/ID" should be deleted
      And the "https" resource should not reference any certificates
