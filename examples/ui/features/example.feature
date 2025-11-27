Feature: Example feature

    Scenario: View web page scenario
        Given I set the viewport to mobile
        When I navigate to "/example"
        Then the page should have the following content
            """
            {
                "p.example-paragraph": "Example web page"
            }
            """
        And element ".example-paragraph" should be visible
        And element ".no-paragraph" should not be visible
        And the page should be accessible
        When I fill in input element ".example-input" with "new value"
        Then input element ".example-input" has value "new value"
        When I click the ".example-button" button
        Then input element ".example-input" has value "clicked"
    
    Scenario: View web page scenario excluding certain accessibility rules
        When I navigate to "/example-accessibility-exclusion"
        Then the page should have the following content
        """
            {"p.example-paragraph" : "Example web page"}
        """
        And the page should be accessible with the following exceptions
        | id        |
        | image-alt |

