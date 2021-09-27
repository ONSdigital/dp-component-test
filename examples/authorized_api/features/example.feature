Feature: Example feature

    Scenario: Example 1 endpoint scenario
        When I GET "/example1"
        Then I should receive the following JSON response with status "200":
            """
            {
                "example_type": 1
            }
            """

    Scenario: accessing restricted endpoint without authorization
        Given I am not authorised
        When I POST "/example3"
            """
            foo bar
            """
        Then the HTTP status code should be "401"
        And I should receive the following response:
            """
            401 - Unauthorized
            """

    Scenario: accessing restricted endpoint with authorization
        Given I am authorised
        When I POST "/example3"
            """
            foo bar
            """
        Then the HTTP status code should be "201"
        And I should receive the following response:
            """
            accepted
            """

    Scenario: accessing restricted endpoint that requires identity
        Given I am authorised
        And I am identified as "admin"
        When I POST "/example4"
            """
            foo bar
            """
        Then the HTTP status code should be "201"
        And I should receive the following response:
            """
            accepted
            """

    Scenario: accessing restricted endpoint that requires identity without identity
        Given I am authorised
        And I am not identified
        When I POST "/example4"
            """
            foo bar
            """
        Then the HTTP status code should be "401"
        And I should receive the following response:
            """
            User has not been identified as an admin
            """

    Scenario: accessing zebedee endpoint without service authorization
        Given I use an invalid service auth token
        When I POST "/example5"
            """
            foo bar
            """
        Then the HTTP status code should be "401"
        And I should receive the following response:
            """
            CMD permissions request denied: service account not found
            """

#    Scenario: accessing zebedee endpoint with service authorization
#        Given I use valid service auth token "abcdefg"
#        When I POST "/example5"
#            """
#            foo bar
#            """
#        Then the HTTP status code should be "200"
#        And I should receive the following response:
#            """
#            ["DELETE", "READ", "CREATE", "UPDATE"]
#            """
