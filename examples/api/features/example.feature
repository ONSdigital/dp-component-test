Feature: Example feature

    Scenario: Example 1 endpoint scenario
        When I GET "/example1"
        Then I should receive the following JSON response with status "200":
        """
        {"example_type": 1}
        """

    Scenario: Example 2 endpoint scenario
        When I POST "/example2"
        """
        foo bar
        """
        Then the HTTP status code should be "403"
        And I wait 2 seconds
        And I should receive the following response:
        """
        403 - Forbidden
        """
    
    Scenario: Example health scenario
        Given I have a healthcheck interval of 1 second
        When I GET "/health"
        Then the health checks should have completed within 2 seconds
        And I should receive the following health JSON response:
          """
            {
              "status": "OK",
              "version": {
                "git_commit": "6584b786caac36b6214ffe04bf62f058d4021538",
                "language": "go",
                "language_version": "go1.24.2",
                "version": "v1.2.3"
              },
              "checks": [
                {
                  "name": "Redis",
                  "status": "OK",
                  "status_code": 200,
                  "message": "redis is healthy"
                }
              ]
            }
          """
    Scenario: Dynamic validation handler
        When I GET "/dynamic/validation"
        Then I should receive the following JSON response with status "200":
        """
        {
          "timestamp": "{{DYNAMIC_TIMESTAMP}}",
          "id": "{{DYNAMIC_UUID}}",
          "embedded": {
            "inner_timestamp": "{{DYNAMIC_RECENT_TIMESTAMP}}"
          },
          "url": "{{DYNAMIC_URL}}"
        }
        """
      