Feature: Example feature

    Scenario: Return the dataset when it exists in collection
        Given the following document exists in the "datasets" collection:
	        """
            {
                "_id": "6021403f3a21177b2837d12f",
                "id": "a1b2c3",
                "exampleData": "some data"
            }
            """
        When I GET "/datasets/a1b2c3"
        Then I should receive the following JSON response with status "200":
            """
            {
                "_id": "6021403f3a21177b2837d12f",
                "id": "a1b2c3",
                "exampleData": "some data"
            }
            """

    Scenario: get 404 if dataset does not exist
        Given the following document exists in the "datasets" collection:
            """
            {
                "_id": "6021403f3a21177b2837d12f",
                "id": "a1b2c3",
                "exampleData": "some data"
            }
            """
        When I GET "/datasets/a1b2c12345678"
        Then the HTTP status code should be "404"
