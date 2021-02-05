package featuretest

import (
	"strconv"
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

func (f *AuthorizationFeature) PostingToTheEndpointReturnsStatusWithBody(postBody, postEndpoint, responseStatus string, responseBody *godog.DocString) error {
	responseStatusCode, err := strconv.Atoi(responseStatus)
	if err != nil {
		return err
	}
	f.FakeAuthService.NewHandler().Post(postEndpoint).AssertBody([]byte(postBody)).Reply(responseStatusCode).BodyString(responseBody.Content)
	return nil
}

func (f *AuthorizationFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I am not identified$`, f.iAmNotIdentified)
	ctx.Step(`^I am identified as "([^"]*)"$`, f.iAmIdentifiedAs)
	ctx.Step(`^POSTING '(.*)' to the endpoint "([^"]*)" returns status "([^"]*)" with body:$`, f.PostingToTheEndpointReturnsStatusWithBody)
}
