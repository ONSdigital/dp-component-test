package featuretest

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

func NewAuthorizationFeature(t *testing.T) *AuthorizationFeature {
	f := &AuthorizationFeature{
		FakeAuthService: httpfake.New(httpfake.WithTesting(t)),
	}

	return f
}

type AuthorizationFeature struct {
	ErrorFeature
	FakeAuthService *httpfake.HTTPFake
}

func (f *AuthorizationFeature) iAmNotIdentified() error {
	f.FakeAuthService.NewHandler().Get("/identity").Reply(401)
	return nil
}

func (f *AuthorizationFeature) iAmIdentifiedAs(username string) error {
	f.FakeAuthService.NewHandler().Get("/identity").Reply(200).BodyString(`{ "identifier": "` + username + `"}`)
	return nil
}

func (f *AuthorizationFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I am not identified$`, f.iAmNotIdentified)
	ctx.Step(`^I am identified as "([^"]*)"$`, f.iAmIdentifiedAs)
}
