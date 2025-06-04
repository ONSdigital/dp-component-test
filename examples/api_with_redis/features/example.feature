Feature: Example feature

    Scenario: Return the value when the key exists in redis
        Given the key "cheese" has a value of "crackers" in the Redis store
        When I GET "/desserts/cheese"
        Then I should receive the following JSON response with status "200":
            """
            {
                "key": "cheese",
                "value": "crackers"
            }
            """

    Scenario: Return a 404 when the key doesn't in redis
        When I GET "/desserts/jelly"
        Then the HTTP status code should be "404"

