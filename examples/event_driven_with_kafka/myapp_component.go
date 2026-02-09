package main

import componenttest "github.com/ONSdigital/dp-component-test"

type MyAppComponent struct {
	kafkaFeature *componenttest.KafkaFeature
}

func NewMyAppComponent(kafkaFeature *componenttest.KafkaFeature) (*MyAppComponent, error) {
	c := &MyAppComponent{kafkaFeature: kafkaFeature}

	return c, nil
}
