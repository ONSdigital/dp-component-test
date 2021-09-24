package componenttest

import (
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

func NewAuthorizationFeature() *AuthorizationFeature {
	f := &AuthorizationFeature{
		FakeAuthService: httpfake.New(),
	}

	return f
}

type AuthorizationFeature struct {
	ErrorFeature
	FakeAuthService *httpfake.HTTPFake
}

func (f *AuthorizationFeature) Reset() {
	f.FakeAuthService.Reset()
}

func (f *AuthorizationFeature) Close() {
	f.FakeAuthService.Close()
}

func (f *AuthorizationFeature) iAmNotIdentified() error {
	f.FakeAuthService.NewHandler().Get("/identity").Reply(401)
	return nil
}

func (f *AuthorizationFeature) iAmIdentifiedAs(username string) error {
	f.FakeAuthService.NewHandler().Get("/identity").Reply(200).BodyString(`{ "identifier": "` + username + `"}`)
	return nil
}

func (f *AuthorizationFeature) iUseAnInvalidServiceAuthToken() error {
	f.FakeAuthService.NewHandler().Get("/serviceInstancePermissions").Reply(401).BodyString(`{ "message": "CMD permissions request denied: service account not found"}`)
	return nil
}

func (f *AuthorizationFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I am not identified$`, f.iAmNotIdentified)
	ctx.Step(`^I am identified as "([^"]*)"$`, f.iAmIdentifiedAs)
	ctx.Step(`^I use an invalid service auth token$`, f.iUseAnInvalidServiceAuthToken)
}
