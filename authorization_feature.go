package componenttest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	permissionsSDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

func NewAuthorizationFeature() *AuthorizationFeature {
	f := &AuthorizationFeature{
		FakeAuthService:    httpfake.New(),
		FakePermissionsAPI: setupFakePermissionsAPI(),
	}

	return f
}

type AuthorizationFeature struct {
	ErrorFeature
	FakeAuthService    *httpfake.HTTPFake
	FakePermissionsAPI *authorisationtest.FakePermissionsAPI
}

func (f *AuthorizationFeature) Reset() {
	f.ErrorFeature.Reset()
	f.FakeAuthService.Reset()
	f.FakePermissionsAPI.Reset()
}

func (f *AuthorizationFeature) Close() {
	f.FakeAuthService.Close()
	f.FakePermissionsAPI.Close()
}

func (f *AuthorizationFeature) RegisterDefaultPermissionsBundle() error {
	emptyBundle := permissionsSDK.Bundle{}
	return f.FakePermissionsAPI.UpdatePermissionsBundleResponse(&emptyBundle)
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

func setupFakePermissionsAPI() *authorisationtest.FakePermissionsAPI {
	fake := authorisationtest.NewFakePermissionsAPI()

	// Optional: preload with a default bundle (empty)
	defaultBundle := permissionsSDK.Bundle{}
	if err := fake.UpdatePermissionsBundleResponse(&defaultBundle); err != nil {
		log.Error(context.Background(), "failed to set default permissions bundle", err)
	}

	return fake
}

func (f *AuthorizationFeature) adminUserHasPermission(permission string) error {
	bundle := &permissionsSDK.Bundle{
		permission: {
			"groups/role-admin": {
				{ID: "1"},
			},
		},
	}
	return f.FakePermissionsAPI.UpdatePermissionsBundleResponse(bundle)
}

func (f *AuthorizationFeature) serviceUserHasPermission(service, permission string) error {
	bundle := &permissionsSDK.Bundle{
		permission: {
			service: {
				{ID: "1"},
			},
		},
	}
	return f.FakePermissionsAPI.UpdatePermissionsBundleResponse(bundle)
}

func (f *AuthorizationFeature) adminUserHasPermissionsJSON(jsonInput string) error {
	var raw map[string]map[string][]permissionsSDK.Policy

	err := json.Unmarshal([]byte(jsonInput), &raw)
	if err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}

	bundle := permissionsSDK.Bundle{}
	for action, entityMap := range raw {
		bundle[action] = make(permissionsSDK.EntityIDToPolicies)
		for entityID, policies := range entityMap {
			bundle[action][entityID] = policies
		}
	}

	return f.FakePermissionsAPI.UpdatePermissionsBundleResponse(&bundle)
}

func (f *AuthorizationFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I am not identified$`, f.iAmNotIdentified)
	ctx.Step(`^I am identified as "([^"]*)"$`, f.iAmIdentifiedAs)
	ctx.Step(`^zebedee recognises the service auth token as valid$`, f.zebedeeRecognisesTheServiceAuthTokenAsValid)
	ctx.Step(`^zebedee recognises the user token as valid$`, f.zebedeeRecognisesTheUserTokenAsValid)
	ctx.Step(`^zebedee does not recognise the service auth token$`, f.zebedeeDoesNotRecogniseTheServiceAuthToken)
	ctx.Step(`^zebedee does not recognise the user token$`, f.zebedeeDoesNotRecogniseTheUserToken)
	ctx.Step(`^service "([^"]*)" has the "([^"]*)" permission$`, f.serviceUserHasPermission)
	ctx.Step(`^an admin user has the "([^"]*)" permission$`, f.adminUserHasPermission)
	ctx.Step(`^an admin user has the following permissions as JSON:$`, f.adminUserHasPermissionsJSON)
}
