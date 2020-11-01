Feature: Managing tls Certificates
  In order manage tls certificates
  As an operator
  I need to be able to issue, rotate and rotate certificates
  And Tyks certificate manager should reconcile accordingly

  @undone
  Scenario: Create a certificate
    Given there is a ./custom_resources/self_signed_issuer.yaml resource
      And there is a ./custom_resources/certificate.yaml resource
    When i "GET" /certs endpoint on the management API
    Then there should be a 200 http response code

  @undone
  Scenario: Rotate a certificate
    Given there is a ./custom_resources/self_signed_issuer.yaml resource
      And there is a ./custom_resources/certificate.yaml resource
    When i rotate the certificate
      And i query the certificate on the management api
    Then there should be a 200 http response code

  @undone
  Scenario: Revoke a certificate
    Given there is a ./custom_resources/self_signed_issuer.yaml resource
      And there is a ./custom_resources/certificate.yaml resource
    When i delete the certificate secret
      And i query the certificate on the management api
    Then there should be a 200 http response code
