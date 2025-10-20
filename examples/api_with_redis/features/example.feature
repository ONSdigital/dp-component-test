Feature: Example feature

    Scenario: Return the value when the key exists in redis
        Given the key "cheese" is already set to a value of "crackers" in the Redis store
        When I GET "/desserts/cheese"
        Then I should receive the following JSON response with status "200":
            """
            {
                "key": "cheese",
                "value": "crackers"
            }
            """

    Scenario: Return a 404 when the key doesn't exist in redis
        When I GET "/desserts/jelly"
        Then the HTTP status code should be "404"

    Scenario: Return a 200 when redis is healthy
        Given redis is healthy
        When I GET "/health"
        Then the HTTP status code should be "200"

    Scenario: Return a 500 when redis is not healthy
        Given redis stops running
        When I GET "/health"
        Then the HTTP status code should be "500"
