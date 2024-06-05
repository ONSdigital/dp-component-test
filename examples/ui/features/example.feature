Feature: Example feature

    Scenario: View web page scenario
        Given I navigate to "/example"
        Then the page should have the following content
            """
            {
                "p.example-paragraph": "Example web page"
            }
            """
        And element ".example-paragraph" should be visible
        And element ".no-paragraph" should not be visible
        When I fill in ".example-input" with "new value"
        Then input element ".example-input" has value "new value"
        When I click the ".example-button" button
        Then input element ".example-input" has value "clicked"
