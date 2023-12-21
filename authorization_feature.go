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
	f.ErrorFeature.Reset()
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

func (f *AuthorizationFeature) zebedeeDoesNotRecogniseTheServiceAuthToken() error {
	f.FakeAuthService.NewHandler().Get("/serviceInstancePermissions").Reply(401).BodyString(`{ "message": "CMD permissions request denied: service account not found"}`)
	return nil
}

func (f *AuthorizationFeature) zebedeeRecognisesTheServiceAuthTokenAsValid() error {
	f.FakeAuthService.NewHandler().Get("/serviceInstancePermissions").Reply(200).BodyString(`{ "permissions": ["DELETE", "READ", "CREATE", "UPDATE"]}`)
	return nil
}

func (f *AuthorizationFeature) zebedeeDoesNotRecogniseTheUserToken() error {
	f.FakeAuthService.NewHandler().Get("/userInstancePermissions").Reply(401).BodyString(`{ "message": "CMD permissions request denied: session not found"}`)
	return nil
}

func (f *AuthorizationFeature) zebedeeRecognisesTheUserTokenAsValid() error {
	// NB. These permissions are what would be returned for an Admin or Publisher user.
	// A viewer would get empty or just "READ" if they were assigned to a team with preview access to a collection/dataset.
	f.FakeAuthService.NewHandler().Get("/userInstancePermissions").Reply(200).BodyString(`{ "permissions": ["DELETE", "READ", "CREATE", "UPDATE"]}`)
	return nil
}

func (f *AuthorizationFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I am not identified$`, f.iAmNotIdentified)
	ctx.Step(`^I am identified as "([^"]*)"$`, f.iAmIdentifiedAs)
	ctx.Step(`^zebedee recognises the service auth token as valid$`, f.zebedeeRecognisesTheServiceAuthTokenAsValid)
	ctx.Step(`^zebedee recognises the user token as valid$`, f.zebedeeRecognisesTheUserTokenAsValid)
	ctx.Step(`^zebedee does not recognise the service auth token$`, f.zebedeeDoesNotRecogniseTheServiceAuthToken)
	ctx.Step(`^zebedee does not recognise the user token$`, f.zebedeeDoesNotRecogniseTheUserToken)
}
