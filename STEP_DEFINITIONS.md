# Step definitions

This file aims to provide a description of what each implemented step in this repository does and should serve to be a
reference guide to anyone writing scenarios for a service.

Below is a table which contains all of the currently implemented steps within this repository. Each step will have a
description of its use as well as which part of the scenario it should be used in (i.e. Given, When or Then)

This library provides these generic steps that should be useable across a variety of projects, however they will
probably not cover all of the desired features or scenarios you might want to test your application against. In this
case you will want to create and define your own steps inside the application which is being tested - examples of this
can be found in the links in the [USAGE](USAGE.md) markdown file

## Key

"QUOTED_VALUE" : a value we pass to the step so that we can customise the scenario to the situation we want

\_DOC_STRING\_ : a value that is provided on a new line underneath the step in a set of three quotation marks - an
example is provided at the bottom of this file.

[LIST] : represents a list of strings e.g `"item1,item2..."`.

### Example

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

You can see how the steps are used in the example above. Single values are provided through the "VALUE" quote step
options, and larger structures are provided through the multi-line doc string options - these are shown as \_BODY\_
variables in the table above.

## Steps

### API Feature steps

| Step                                                                                 | What it does                                                                         | Scenario Position |
|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|-------------------|
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
| I have a healthcheck interval of "SECONDS" seconds                                   | Set the healthcheck interval                                                         | Given             |
| the health checks should have completed within "SECONDS" seconds                     | Set the expected time for health check completion                                    | When              |
| I should receive the following health JSON response \_BODY\_                         | Assert that the health check response body matches BODY                              | Then              |
| I should receive the following JSON response: \_BODY\_                               | Assert that the response body is JSON and that it matches BODY                       | Then              |
| I should receive the following JSON response with status "CODE": \_BODY\_            | Assert that the response code is CODE and the body is json which matches BODY        | Then              |
| I wait "SECONDS" seconds                                                             | Waits a given amount of seconds                                                      | Then              |
| the document with "KEY" set to "VALUE" does not exist in the "COLLECTION" collection | Assert that a document with KEY set to VALUE does not exist in COLLECTION collection | Then              |

### Redis Feature steps

| Step                                                    | What it does                                         | Scenario Position |
|---------------------------------------------------------|------------------------------------------------------|-------------------|
| the key "KEY" has a value of "VALUE" in the Redis store | Set the KEY to VALUE in the fake redis               | Given             |
| redis is healthy                                        | This pings the in-memory redis to check it's running | Given             |
| redis stops running                                     | This shuts down the in-memory redis                  | Given             |

### Authorization Feature steps

| Step                                                 | What it does                                                                                          | Scenario Position |
|------------------------------------------------------|-------------------------------------------------------------------------------------------------------|-------------------|
| I am authorised                                      | set the request Authorization header to a random token                                                | Given             |
| I am not authorised                                  | clear any existing Authorization token from the request                                               | Given             |
| I am not identified                                  | remove the /identity endpoint from the stubbed identity server                                        | Given             |
| I am an admin user                                   | set the request Authorization header to an admin JWT token                                            | Given             |
| I am not authenticated                               | removes any Authorization header set in the request headers                                           | Given             |
| I am identified as "USER"                            | set /identity endpoint to return response with USER identity                                          | Given             |
| service "SERVICE" has the "PERMISSION" permission    | Configure the fake permissions API to grant a permission to a specific service account                | Given             |
| an admin user has the "PERMISSION" permission        | Configure the fake permissions API to grant a single permission to the admin user                     | Given             |
| an admin user has the following permissions as JSON: | Configure the fake permissions API to grant multiple permissions to the admin user using a JSON input | Given             |

### UI Feature steps

| Step                                                                     | What it does                                                                                        | Scenario Position |
|--------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------|-------------------|
| I navigate to "URL"                                                      | Navigate to URL in Chrome                                                                           | When              |
| input element "SELECTOR" has value "VALUE"                               | Assert that a HTML input element on the web page has the value VALUE                                | Then              |
| element "SELECTOR" should be visible                                     | Assert that a HTML element on the web page matches SELECTOR                                         | Then              |
| element "SELECTOR" should not be visible                                 | Assert that no HTML element on the web page matches SELECTOR                                        | Then              |
| the page should have the following content \_CONTENT\_                   | Assert that a HTML element on the web page matches CONTENT selector and value [^1]                  | Then              |
| the page should contain "KEY" with list element text [LIST] at INT depth | Assert that the expected breadcrumbs siblings are present                                           | Then              |
| I fill in input element "SELECTOR" with value "VALUE"                    | Find a HTML input element on the web page that matches SELECTOR and fill it with VALUE              | When              |
| I click the "SELECTOR" element                                           | Find a HTML element on the web page that matches SELECTOR and simulate a click event                | When              |
| the page should be accessible                                            | Assert that the page meets WCAG A and AA accessibility criteria                                     | Then              |
| the page should be accessible with the following exceptions \_LIST\_     | Assert that the page meets WCAG A and AA accessibility criteria whilst ignoring the exceptions.[^2] | Then              |

[^1]: CONTENT selector must include a HTML id or class - an element type alone will not work!
[^2]: List should be a table of ids taken from
the [list of axe-core rules](https://github.com/dequelabs/axe-core/blob/develop/doc/rule-descriptions.md)
