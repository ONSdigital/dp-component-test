# Usage

This file serves to provide directions to repositories which have currently implemented component testing, as well as useful resources so that anyone intending to start using this library will have several references to draw from. These references will be provided roughly in order of usefulness.

## dp-topic-api

[In this repository](https://github.com/ONSdigital/dp-topic-api) a [walkthrough document](https://github.com/ONSdigital/dp-topic-api/blob/develop/Adding%20Component%20Testing%20-%20HOWTO.md) has been created which details how component testing was added to this repository. This documentation details the process of adding component testing to dp-topic-api and should be very transferrable, although each project will have its own nuances.

This repository also details the mechanism by which component testing is added to the CI pipeline of an application.

## dp-dataset-api

[In this repository](https://github.com/ONSdigital/dp-dataset-api) component testing has been implemented for a service using auth, identity and mongo database. This should provide useful example implementation as, like the topic API, the project structure was generated from the same repository generation tool and therefore should have the same kind of structure as any new service.

## dp-observation-importer

[This project](https://github.com/ONSdigital/dp-observation-importer) shows how an older service had to be restructured to allow for dependency injection for component testing to be added. This is mostly detailed in [this commit](https://github.com/ONSdigital/dp-observation-importer/commit/66ade9ecf3dac07ed598c2e6846d0a2a209c8ced)

## Florence

An initial component test has been added to [Florence](https://github.com/ONSdigital/florence). It would be worthwhile looking into this repository if you are working on any similar UI services.

## Godog

[Godog](https://github.com/cucumber/godog) is the test runner used for the scenarios, and it is worthwhile understanding the basics of how the framework functions.
