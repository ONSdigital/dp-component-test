# Step definitions

This file aims to provide a description of what each implemented step in this repository does and should serve to be a reference guide to anyone writing scenarios for a service.

Below is a table which contains all of the currently implemented steps within this repository. Each step will have a description of its use as well as which part of the scenario it should be used in (i.e. Given, When or Then)

This library provides these generic steps that should be useable across a variety of projects, however they will probably not cover all of the desired features or scenarios you might want to test your application against. In this case you will want to create and define your own steps inside the application which is being tested - examples of this can be found in the links in the [USAGE](USAGE.md) markdown file

**KEY**

"QUOTED_VALUE" : a value we pass to the step so that we can customise the scenario to the situation we want

\_DOC_STRING\_ : a value that is provided on a new line underneath the step in a set of three quotation marks - an example is provided at the bottom of this file.

[LIST] : represents a list of strings e.g `"item1,item2..."`.

| Step                                                                                 | What it does                                                                         | Scenario Position |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ----------------- |
| I am authorised                                                                      | set the request Authorization header to a random token                               | Given             |
| I am not authorised                                                                  | clear any existing Authorization token from the request                              | Given             |
| I am not identified                                                                  | remove the /identity endpoint from the stubbed identity server                       | Given             |
| I am identified as "USER"                                                            | set /identity endpoint to return response with USER identity                         | Given             |
| the following document exists in the "COLLECTION" collection: \_BODY\_               | put document BODY in the COLLECTION collection                                       | Given             |
| I set the "KEY" header to "VALUE"                                                    | set a HTTP header of the request to the value                                        | Given             |
| I GET "URL"                                                                          | make a GET request to the provided URL                                               | When              |
| I DELETE "URL"                                                                       | make a DELETE request to the provided URL                                            | When              |
| I PUT "URL" "BODY"                                                                   | make a PUT request to the provided URL with the given body                           | When              |
| I PATCH "URL" "BODY"                                                                 | make a PATCH request to the provided URL with the given body                         | When              |
| I POST "URL" "BODY"                                                                  | make a POST request to the provided URL with the given body                          | When              |
| the HTTP status code should be "CODE"                                                | Assert that the response code from the request is CODE                               | Then              |
| the response header "KEY" should be "VALUE"                                          | Assert that the response header KEY has value VALUE                                  | Then              |
| I should recieve the following response \_BODY\_                                     | Assert that the response body matches BODY                                           | Then              |
| I should receive the following JSON response: \_BODY\_                               | Assert that the response body is JSON and that ir matches BODY                       | Then              |
| I should receive the following JSON response with status "CODE": \_BODY\_            | Assert that the response code is CODE and the body is json which matches BODY        | Then              |
| the document with "KEY" set to "VALUE" does not exist in the "COLLECTION" collection | Assert that a document with KEY set to VALUE does not exist in COLLECTION collection | Then              |
| I navigate to "URL"                                                                  | Navigate to URL in Chrome | When              |
| element "SELECTOR" should be visible                                                 | Assert that a HTML element on the web page matches SELECTOR | Then              |
| the page should have the following content \_CONTENT\_                               | Assert that a HTML element on the web page matches CONTENT selector and value (CONTENT selector must include a HTML id or class - an element type alone will not work!) | Then              |
| the beta phase banner should be visible                                              | Assert that the beta phase banner exists on the web page | Then              |
| the improve this page banner should be visible                                       | Assert that the improve this page banner exists on the web page | Then              |
| the page should contain "KEY" with list element text [LIST] at INT depth             | Assert that the expected breadcrumbs siblings are present | Then
---

## Example

```gherkin
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

```

You can see how the steps are used in the example above. Single values are provided through the "VALUE" quote step options, and larger structures are provided through the multi-line doc string options - these are shown as \_BODY\_ variables in the table above.
