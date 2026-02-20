Feature: Example feature

    Scenario: JSON event consumed causes single event produced
        Given the service is started with JSON configured
        When this "input" event is queued, to be consumed:
            """
            {
              "input":         "Hello",
              "qty"  : 1
            }
            """
        Then this "output" event is produced:
            """
            {
              "id" : 0,
              "input":         "Hello",
              "output":        "World!"
            }
            """

    Scenario: JSON event consumed causes no event produced
        Given the service is started with JSON configured
        When this "input" event is queued, to be consumed:
            """
            {
              "input":         "Nothing",
              "qty" : 0
            }
            """
        Then no "output" event is produced within 5 seconds

    Scenario: JSON event consumed causes multiple events produced
        Given the service is started with JSON configured
        When this "input" event is queued, to be consumed:
            """
            {
              "input":         "Hello",
              "qty"  : 3
            }
            """
        Then this "output" event is produced:
            """
            {
              "id" : 0,
              "input":         "Hello",
              "output":        "World!"
            }
            """
        And this "output" event is produced:
            """
            {
              "id" : 1,
              "input":         "Hello",
              "output":        "World!"
            }
            """
        And this "output" event is produced:
            """
            {
              "id" : 2,
              "input":         "Hello",
              "output":        "World!"
            }
            """
