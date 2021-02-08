Feature: Example feature

    Scenario: Example 1 endpoint scenario
        When I GET "/example1"
        Then I should receive the following JSON response with status "200":
            """
            {
                "example_type": 1
            }
            """

    Scenario: Example 3 accessing restricted endpoint without authorization
        Given I am not identified
        When I POST "/example3"
            """
            foo bar
            """
        Then the HTTP status code should be "401"
        And I should receive the following response:
            """
            401 - Unauthorized
            """
