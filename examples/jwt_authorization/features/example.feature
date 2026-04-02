Feature: Example feature

  Scenario: Accessing endpoint with JWT authorization
    Given I am a JWT user with email "viewer1@ons.gov.uk" and group "role-viewer-allowed"
    When I GET "/checkjwt"
    Then The HTTP status code should be "200"