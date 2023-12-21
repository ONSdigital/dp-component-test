Feature: Example feature

    Scenario: View web page scenario
        When I navigate to "/example"
        Then the page should have the following content
        """
            {"p.example-paragraph" : "Example web page"}
        """
        And input element ".example-input" has value "test value"
