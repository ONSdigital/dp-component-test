Feature: Example feature

    Scenario: event consumed causes event produced
        Given this "input" event is queued, to be consumed:
            """
            {
              "input":         "Hello"
            }
            """
        Then this "output" event is produced:
            """
            {
              "input":         "Hello",
              "output":        "World!"
            }
            """

    Scenario: event consumed causes no event produced
        Given this "input" event is queued, to be consumed:
            """
            {
              "input":         "no-output"
            }
            """
        Then no "output" event is produced within 5 seconds
