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
            user has not been identified as an admin
            """

    Scenario: accessing zebedee endpoint without service authorization
        Given I use a service auth token "invalidServiceAuthToken"
        And zebedee does not recognise the service auth token
        When I POST "/example5"
            """
            foo bar
            """
        Then the HTTP status code should be "401"
        And I should receive the following response:
            """
            CMD permissions request denied: service account not found
            """

    Scenario: accessing zebedee endpoint with service authorization
        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        When I POST "/example5"
            """
            foo bar
            """
        Then the HTTP status code should be "200"
        And I should receive the following response:
            """
            ["DELETE", "READ", "CREATE", "UPDATE"]
            """

    Scenario: accessing zebedee endpoint without X Florence user permissions

        Given I use an X Florence user token "invalidXFlorenceToken"
        And I am not identified
        And zebedee does not recognise the user token
        When I POST "/example6"
            """
            foo bar
            """
        Then the HTTP status code should be "401"
        And I should receive the following response:
            """
            CMD permissions request denied: session not found
            """

    Scenario: accessing zebedee endpoint with X Florence user permissions

        Given I use an X Florence user token "validXFlorenceToken"
        And I am identified as "someone@somewhere.com"
        And zebedee recognises the user token as valid
        When I POST "/example6"
            """
            foo bar
            """
        Then the HTTP status code should be "200"
        And I should receive the following response:
            """
            ["DELETE", "READ", "CREATE", "UPDATE"]
            """


    Scenario: accessing a zebedee endpoint with a service that has a specific permission
        Given I use a service auth token "validServiceToken"
        And service "my-service" has the "read" permission
        When I POST "/example3"
        """
        some payload
        """
        Then the HTTP status code should be "201"
        And I should receive the following response:
        """
        accepted
        """


    Scenario: accessing a zebedee endpoint with an admin user that has a specific permission
        Given I am authorised
        And an admin user has the "update" permission
        And I am an admin user
        When I POST "/example3"
        """
        some update payload
        """
        Then the HTTP status code should be "201"
        And I should receive the following response:
        """
        accepted
        """

    Scenario: accessing a zebedee endpoint with an unauthenticated user
        Given I am authorised
        And an admin user has the "update" permission
        And I am not authenticated
        When I POST "/example3"
        """
        some update payload
        """
        Then the HTTP status code should be "401"
        And I should receive the following response:
        """
        401 - Unauthorized
        """


    Scenario: an admin user has multiple permissions provided as JSON
        Given I am authorised
        And an admin user has the following permissions as JSON:
            """
            {
                "read": {
                    "users/admin": [
                        { "id": "1" }
                    ]
                },
                "update": {
                    "users/admin": [
                        { "id": "1" }
                    ]
                },
                "delete": {
                    "users/admin": [
                        { "id": "1" }
                    ]
                }
            }
            """
        And I am an admin user
        When I POST "/example3"
            """
            delete this thing
            """
        Then the HTTP status code should be "201"
        And I should receive the following response:
            """
            accepted
            """

    Scenario: a publisher user has multiple permissions provided as JSON
        Given I am authorised
        And a publisher user has the following permissions as JSON:
            """
            {
                "read": {
                    "users/publisher": [
                        { "id": "1" }
                    ]
                },
                "update": {
                    "users/publisher": [
                        { "id": "1" }
                    ]
                }
            }
            """
        And I am a publisher user
        When I PUT "/example3"
            """
            update something
            """
        Then the HTTP status code should be "200"
        And I should receive the following response:
            """
            accepted
            """


