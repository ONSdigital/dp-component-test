Feature: Example feature

    Scenario: Return the dataset when it exists in collection
        Given the following document exists in the "datasets" collection:
            """
            {
                "_id": "6021403f3a21177b2837d12f",
                "id": "a1b2c3",
                "example_data": "some data"
            }
            """
        When I GET "/datasets/a1b2c3"
        Then I should receive the following JSON response with status "200":
            """
            {
                "_id": "6021403f3a21177b2837d12f",
                "id": "a1b2c3",
                "example_data": "some data"
            }
            """

    Scenario: Return the dataset in text when accept header set to text/html
        Given the following document exists in the "datasets" collection:
            """
            {
                "_id": "6021403f3a21177b2837d12f",
                "id": "a1b2c3",
                "example_data": "some data"
            }
            """
        When I set the "Accept" header to "text/html"
        And I GET "/datasets/a1b2c3"
        Then the response header "Content-Type" should be "text/html"
        And I should receive the following response:
            """
            <value id="_id">6021403f3a21177b2837d12f</value><value id="id">a1b2c3</value><value id="example_data">some data</value>
            """

    Scenario: get 404 if dataset does not exist
        Given the following document exists in the "datasets" collection:
            """
            {
                "_id": "6021403f3a21177b2837d12f",
                "id": "a1b2c3",
                "example_data": "some data"
            }
            """
        When I GET "/datasets/a1b2c12345678"
        Then the HTTP status code should be "404"

    Scenario: data removed from db if dataset has been deleted
        Given the following document exists in the "datasets" collection:
            """
            {
                "id": "a1b2c3",
                "example_data": "some data"
            }
            """
        When I DELETE "/datasets/a1b2c3"
        Then the HTTP status code should be "204"
        And the document with "id" set to "a1b2c3" does not exist in the "datasets" collection

    Scenario: document stored in database after a PUT
        When I PUT "/datasets/1"
            """
            {
                "_id": "somevalue",
                "id": "1",
                "example_data": "some data"
            }
            """
        Then the HTTP status code should be "200"