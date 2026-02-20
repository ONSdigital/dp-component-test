Feature: Example feature

    Scenario: Avro event consumed causes single event produced
        Given the service is started with Avro configured
        When this "input" Avro event is queued, to be consumed:
            """
            {
              "input":         "Hello",
              "qty"  : 1
            }
            """
        Then this "output" Avro event is produced:
            """
            {
              "id" : 0,
              "input":         "Hello",
              "output":        "World!"
            }
            """

    Scenario: Avro event consumed causes no event produced
        Given the service is started with Avro configured
        When this "input" Avro event is queued, to be consumed:
            """
            {
              "input":         "Nothing",
              "qty" : 0
            }
            """
        Then no "output" event is produced within 5 seconds

    Scenario: Avro event consumed causes multiple events produced
        Given the service is started with Avro configured
        When this "input" Avro event is queued, to be consumed:
            """
            {
              "input":         "Hello",
              "qty"  : 3
            }
            """
        Then this "output" Avro event is produced:
            """
            {
              "id" : 0,
              "input":         "Hello",
              "output":        "World!"
            }
            """
        And this "output" Avro event is produced:
            """
            {
              "id" : 1,
              "input":         "Hello",
              "output":        "World!"
            }
            """
        And this "output" Avro event is produced:
            """
            {
              "id" : 2,
              "input":         "Hello",
              "output":        "World!"
            }
            """
