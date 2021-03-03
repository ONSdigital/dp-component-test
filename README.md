# dp-component-test

Library to help write feature-level tests against a REST api / microservice

## Background

The intention of this library is to help when writing feature tests for a new or existing component (microservice).
The library contains a set of useful helper feature steps to make writing new gherkin tests easy.

The steps in api_feature have been written as to be easily reusable when setting up tests against a REST API, and
there are other additional parts of this library which can be plugged in to help test outputs of the tests e.g. setting
up an in memory mongo to assert against changes to the database.

## Installation

To install this package in your project simply run:

```bash
go get github.com/ONSdigital/dp-component-test
```

## Running tests

`go test -component -v`

This package uses the Godog BDD framework. 
For instructions on writing Godog tests [it is best to follow the instructions found here](https://github.com/cucumber/godog)

To integrate your component tests with this library all you need to do is update your root level test file to pass
the http handler of your application to our NewAPIFeature, register the steps and add the reset function to the BeforeScenario function.

```go
package main

import (
	componenttest "github.com/ONSdigital/dp-component-test"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	myAppFeature := NewMyAppFeature() // This is the part that YOU will implement
	apiFeature := componenttest.NewAPIFeature(myAppFeature.Handler)

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

func TestMain(t *testing.T) {
	if *componentFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
		}

		f := &ComponentTest{}

		status := godog.TestSuite{
			Name:                 "component_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
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
