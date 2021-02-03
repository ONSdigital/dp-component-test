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
        And I should receive the following response:
        """
        403 - Forbidden
        """