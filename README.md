# dp-component-test

Library to help write feature-level tests against a REST api / microservice

The steps available to use from this library are described in [STEP_DEFINITIONS.md](STEP_DEFINITIONS.md).

For help and examples of using this library please see [USAGE.md](USAGE.md).

## Background

The intention of this library is to help when writing component tests for a new or existing component (microservice).
The library contains a set of useful helper steps to make writing new gherkin tests easy.

The steps in api_feature have been written as to be easily reusable when setting up tests against a REST API, and
there are other additional parts of this library which can be plugged in to help test outputs of the tests e.g. setting
up an in memory mongo to assert against changes to the database.

## Installation

To install this package in your project simply run:

```bash
go get github.com/ONSdigital/dp-component-test
```

## Running tests

`go test -component`

This package uses the Godog BDD framework.
For instructions on writing Godog tests [it is best to follow the instructions found here](https://github.com/cucumber/godog)

To integrate your component tests with this library all you need to do is update your root level test file to pass
the http handler of your application to our NewAPIFeature, register the steps and add the reset function to the BeforeScenario function.

```go
package main

import (
	componenttest "github.com/ONSdigital/dp-component-test"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(ctx *godog.ScenarioContext) {
	myAppComponent := NewMyAppComponent() // This is the part that YOU will implement
	apiFeature := componenttest.NewAPIFeature(myAppComponent.Handler)

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
	})
	ctx.AfterScenario(func(*godog.Scenario, error) {
	})

	apiFeature.RegisterSteps(ctx)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
		}
		
		status := godog.TestSuite{
			Name:                 "component_tests",
			ScenarioInitializer:  InitializeScenario,
			TestSuiteInitializer: InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
        }
	} else {
		t.Skip("component flag required to run component tests")
	}
}

```

## Repository structure

The features that can be used all exist on the root level of the project.

The examples folder contains three different examples of how to use this library, each using different
features and having a slightly different way of setting up.

## Adding new component test features

If you feel like there are handy common steps missing from this library which you would like to add, please do!

The mechanism by which the tests and steps are validated (testing the test library) is through feature tests in the examples which get exercised in the CI pipeline.
If you add any new steps, make sure you also add sufficient feature tests to exercise them in appropriate examples in the examples folder.

If you are adding a new testing feature entirely, it might be worth adding a new example service which exercises the steps of any new feature you add.
