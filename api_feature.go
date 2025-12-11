package componenttest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-component-test/validator"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

type ServiceInitialiser func() (http.Handler, error)

func StaticHandler(handler http.Handler) ServiceInitialiser {
	return func() (http.Handler, error) {
		return handler, nil
	}
}

// APIFeature contains the information needed to test REST API requests
type APIFeature struct {
	ErrorFeature
	Initialiser          ServiceInitialiser
	HTTPResponse         *http.Response
	BeforeRequestHook    func() error
	requestHeaders       map[string]string
	StartTime            time.Time
	HealthCheckInterval  time.Duration
	ExpectedResponseTime time.Duration
}

// HealthCheckTest represents a test healthcheck struct that mimics the real healthcheck struct
type HealthCheckTest struct {
	Status    string                  `json:"status"`
	Version   healthcheck.VersionInfo `json:"version"`
	Uptime    time.Duration           `json:"uptime"`
	StartTime time.Time               `json:"start_time"`
	Checks    []*Check                `json:"checks"`
}

// Check represents a health status of a registered app that mimics the real check struct
type Check struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	StatusCode  int        `json:"status_code"`
	Message     string     `json:"message"`
	LastChecked *time.Time `json:"last_checked"`
	LastSuccess *time.Time `json:"last_success"`
	LastFailure *time.Time `json:"last_failure"`
}

// NewAPIFeature returns a new APIFeature, takes a function to retrieve the bound handler just before a request is made
func NewAPIFeature(initialiser ServiceInitialiser) *APIFeature {
	return &APIFeature{
		Initialiser:    initialiser,
		requestHeaders: make(map[string]string),
		StartTime:      time.Now(),
	}
}

// NewAPIFeatureWithHandler create a new APIFeature with a handler already bound with your endpoints
func NewAPIFeatureWithHandler(handler http.Handler) *APIFeature {
	return NewAPIFeature(StaticHandler(handler))
}

// Reset the request headers
func (f *APIFeature) Reset() {
	f.ErrorFeature.Reset()
	f.requestHeaders = make(map[string]string)
}

// RegisterSteps binds the APIFeature steps to the godog context to enable usage in the component tests
func (f *APIFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I set the "([^"]*)" header to "([^"]*)"$`, f.ISetTheHeaderTo)
	ctx.Step(`^I am authorised$`, f.IAmAuthorised)
	ctx.Step(`^I am not authorised$`, f.IAmNotAuthorised)
	ctx.Step(`^I GET "([^"]*)"$`, f.IGet)
	ctx.Step(`^I POST "([^"]*)"$`, f.IPostToWithBody)
	ctx.Step(`^I PUT "([^"]*)"$`, f.IPut)
	ctx.Step(`^I PATCH "([^"]*)"$`, f.IPatch)
	ctx.Step(`^I DELETE "([^"]*)"`, f.IDelete)
	ctx.Step(`^I am an admin user$`, f.adminJWTToken)
	ctx.Step(`^I am a publisher user$`, f.publisherJWTToken)
	ctx.Step(`^I am not authenticated$`, f.iAmNotAuthenticated)
	ctx.Step(`^the HTTP status code should be "([^"]*)"$`, f.TheHTTPStatusCodeShouldBe)
	ctx.Step(`^the response header "([^"]*)" should be "([^"]*)"$`, f.TheResponseHeaderShouldBe)
	ctx.Step(`^I should receive the following response:$`, f.IShouldReceiveTheFollowingResponse)
	ctx.Step(`^I have a healthcheck interval of (\d+) seconds?$`, f.iHaveAHealthCheckIntervalOfSecond)
	ctx.Step(`^the health checks should have completed within (\d+) seconds?$`, f.theHealthChecksShouldHaveCompletedWithinSeconds)
	ctx.Step(`^I should receive the following health JSON response:$`, f.iShouldReceiveTheFollowingHealthJSONResponse)
	ctx.Step(`^I should receive the following JSON response:$`, f.IShouldReceiveTheFollowingJSONResponse)
	ctx.Step(`^I should receive the following JSON response with status "([^"]*)":$`, f.IShouldReceiveTheFollowingJSONResponseWithStatus)
	ctx.Step(`^I use a service auth token "([^"]*)"$`, f.IUseAServiceAuthToken)
	ctx.Step(`^I use an X Florence user token "([^"]*)"$`, f.IUseAnXFlorenceUserToken)
	ctx.Step(`^I wait (\d+) seconds`, f.delayTimeBySeconds)
}

func (f *APIFeature) adminJWTToken() error {
	err := f.ISetTheHeaderTo("Authorization", authorisationtest.AdminJWTToken)
	return err
}

func (f *APIFeature) publisherJWTToken() error {
	err := f.ISetTheHeaderTo("Authorization", authorisationtest.PublisherJWTToken)
	return err
}

func (f *APIFeature) iAmNotAuthenticated() error {
	err := f.ISetTheHeaderTo("Authorization", "")
	return err
}

func (f *APIFeature) IUseAServiceAuthToken(serviceAuthToken string) error {
	err := f.ISetTheHeaderTo("Authorization", "Bearer "+serviceAuthToken)
	return err
}

func (f *APIFeature) IUseAnXFlorenceUserToken(xFlorenceToken string) error {
	err := f.ISetTheHeaderTo("X-Florence-Token", xFlorenceToken)
	return err
}

// ISetTheHeaderTo is a default step used to set a header and associated value for the next request
func (f *APIFeature) ISetTheHeaderTo(header, value string) error {
	f.requestHeaders[header] = value
	return nil
}

// IAmAuthorised sets the Authorization header to a bogus token
func (f *APIFeature) IAmAuthorised() error {
	err := f.ISetTheHeaderTo("Authorization", "bearer SomeFakeToken")
	return err
}

// IAmNotAuthorised removes any Authorization header set in the request headers
func (f *APIFeature) IAmNotAuthorised() error {
	delete(f.requestHeaders, "Authorization")
	return nil
}

// IGet makes a get request to the provided path with the current headers
func (f *APIFeature) IGet(path string) error {
	return f.makeRequest("GET", path, nil)
}

// IPostToWithBody makes a POST request to the provided path with the current headers and the body provided
func (f *APIFeature) IPostToWithBody(path string, body *godog.DocString) error {
	return f.makeRequest("POST", path, []byte(body.Content))
}

// IPut makes a PUT request to the provided path with the current headers and the body provided
func (f *APIFeature) IPut(path string, body *godog.DocString) error {
	return f.makeRequest("PUT", path, []byte(body.Content))
}

// IPatch makes a PATCH request to the provided path with the current headers and the body provided
func (f *APIFeature) IPatch(path string, body *godog.DocString) error {
	return f.makeRequest("PATCH", path, []byte(body.Content))
}

// IDelete makes a DELETE request to the provided path with the current headers
func (f *APIFeature) IDelete(path string) error {
	return f.makeRequest("DELETE", path, nil)
}

func (f *APIFeature) makeRequest(method, path string, data []byte) error {
	handler, err := f.Initialiser()
	if err != nil {
		return err
	}
	req := httptest.NewRequest(method, "http://foo"+path, bytes.NewReader(data))
	for key, value := range f.requestHeaders {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	f.HTTPResponse = w.Result()
	return nil
}

// IShouldReceiveTheFollowingResponse asserts the response body and expected response body are equal
func (f *APIFeature) IShouldReceiveTheFollowingResponse(expectedAPIResponse *godog.DocString) error {
	responseBody := f.HTTPResponse.Body
	body, _ := io.ReadAll(responseBody)

	assert.Equal(f, strings.TrimSpace(expectedAPIResponse.Content), strings.TrimSpace(string(body)))

	return f.StepError()
}

// IShouldReceiveTheFollowingJSONResponse asserts that the response body and expected response body are equal.
// This also validates any "{{DYNAMIC_TIMESTAMP}}" fields.
func (f *APIFeature) IShouldReceiveTheFollowingJSONResponse(expectedAPIResponse *godog.DocString) error {
	responseBody := f.HTTPResponse.Body
	body, err := io.ReadAll(responseBody)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	actualValidated, expectedValidated, err := validateDynamicValues(string(body), expectedAPIResponse.Content)
	if err != nil {
		return err
	}

	assert.JSONEq(f, expectedValidated, actualValidated)

	return f.StepError()
}

// TheHTTPStatusCodeShouldBe asserts that the status code of the response matches the expected code
func (f *APIFeature) TheHTTPStatusCodeShouldBe(expectedCodeStr string) error {
	expectedCode, err := strconv.Atoi(expectedCodeStr)
	if err != nil {
		return err
	}
	assert.Equal(f, expectedCode, f.HTTPResponse.StatusCode)
	return f.StepError()
}

// TheResponseHeaderShouldBe asserts the response header matches the expectation
func (f *APIFeature) TheResponseHeaderShouldBe(headerName, expectedValue string) error {
	assert.Equal(f, expectedValue, f.HTTPResponse.Header.Get(headerName))
	return f.StepError()
}

// IShouldReceiveTheFollowingJSONResponseWithStatus asserts the response code and body match the expectation.
// This also validates any "{{DYNAMIC_TIMESTAMP}}" fields.
func (f *APIFeature) IShouldReceiveTheFollowingJSONResponseWithStatus(expectedCodeStr string, expectedBody *godog.DocString) error {
	if err := f.TheHTTPStatusCodeShouldBe(expectedCodeStr); err != nil {
		return err
	}
	if err := f.TheResponseHeaderShouldBe("Content-Type", "application/json"); err != nil {
		return err
	}
	return f.IShouldReceiveTheFollowingJSONResponse(expectedBody)
}

// DynamicValidator represents a validator for dynamic placeholder values in JSON.
// It contains a validation function to check if a value is valid and a placeholder
// string to replace valid values with for comparison.
type DynamicValidator struct {
	ValidationFunc func(value string) bool
	Placeholder    string
}

// dynamicValidators is a registry of all available dynamic validators, keyed by
// their placeholder type (e.g., "TIMESTAMP", "UUID").
var dynamicValidators = map[string]DynamicValidator{
	"TIMESTAMP": {
		ValidationFunc: validator.ValidateTimestamp,
		Placeholder:    "VALID_TIMESTAMP",
	},
	"RECENT_TIMESTAMP": {
		ValidationFunc: validator.ValidateRecentTimestamp,
		Placeholder:    "VALID_RECENT_TIMESTAMP",
	},
	"URI_PATH": {
		ValidationFunc: validator.ValidateURIPath,
		Placeholder:    "VALID_URI_PATH",
	},
	"URL": {
		ValidationFunc: validator.ValidateURL,
		Placeholder:    "VALID_URL",
	},
	"UUID": {
		ValidationFunc: validator.ValidateUUID,
		Placeholder:    "VALID_UUID",
	},
}

// validateDynamicValues checks for any fields in expected with dynamic value
// strings, e.g. "{{DYNAMIC_TIMESTAMP}}", validates them and replaces them
// with a placeholder.
func validateDynamicValues(actual, expected string) (actualValidated, expectedValidated string, err error) {
	var actualJSON, expectedJSON interface{}

	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal actual JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}

	// Single pass validation and replacement
	if err := validateAndReplace(actualJSON, expectedJSON, ""); err != nil {
		return "", "", err
	}

	actualBytes, err := json.Marshal(actualJSON)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal actual JSON: %w", err)
	}

	expectedBytes, err := json.Marshal(expectedJSON)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal expected JSON: %w", err)
	}

	return string(actualBytes), string(expectedBytes), nil
}

// validateAndReplace traverses actual and expected JSON structures, validating
// dynamic placeholders and replacing them with normalized values.
func validateAndReplace(actual, expected interface{}, path string) error {
	switch exp := expected.(type) {
	case map[string]interface{}:
		return validateObject(actual, exp, path)
	case []interface{}:
		return validateArray(actual, exp, path)
	}
	return nil
}

// validateObject validates that actual is an object matching the structure of
// expected, checking all fields and recursing into nested structures.
func validateObject(actual interface{}, expected map[string]interface{}, path string) error {
	act, ok := actual.(map[string]interface{})
	if !ok {
		return fmt.Errorf("type mismatch at %s: expected object, got %T", path, actual)
	}

	for key, expValue := range expected {
		currentPath := buildPath(path, key)

		actValue, exists := act[key]
		if !exists {
			return fmt.Errorf("missing field at %s", currentPath)
		}

		if err := validateField(act, expected, key, actValue, expValue, currentPath); err != nil {
			return err
		}
	}
	return nil
}

// validateArray validates that actual is an array with the same length and
// structure as expected, validating each element.
func validateArray(actual interface{}, expected []interface{}, path string) error {
	act, ok := actual.([]interface{})
	if !ok {
		return fmt.Errorf("type mismatch at %s: expected array, got %T", path, actual)
	}

	if len(act) != len(expected) {
		return fmt.Errorf("array length mismatch at %s: expected %d, got %d",
			path, len(expected), len(act))
	}

	for i := range expected {
		currentPath := buildPath(path, fmt.Sprintf("[%d]", i))
		if err := validateAndReplace(act[i], expected[i], currentPath); err != nil {
			return err
		}
	}
	return nil
}

// validateField validates a single field, either by checking dynamic placeholders
// or recursing into nested structures.
func validateField(act, exp map[string]interface{}, key string, actValue, expValue interface{}, path string) error {
	// Check if it's a dynamic placeholder
	expStr, ok := expValue.(string)
	if ok && strings.HasPrefix(expStr, "{{DYNAMIC_") {
		return validateDynamicField(act, exp, key, actValue, expStr, path)
	}

	// Recurse into nested structures
	if shouldRecurse(actValue, expValue) {
		return validateAndReplace(actValue, expValue, path)
	}
	return nil
}

// validateDynamicField validates a field with a dynamic placeholder against the
// corresponding validator, and replaces both actual and expected values with
// normalized placeholders if validation succeeds.
func validateDynamicField(act, exp map[string]interface{}, key string, actValue interface{}, expStr, path string) error {
	validationType := extractValidationType(expStr)

	v, exists := dynamicValidators[validationType]
	if !exists {
		return fmt.Errorf("unknown validation type: %s", validationType)
	}

	actStr, ok := actValue.(string)
	if !ok {
		return fmt.Errorf("field at %s is not a string", path)
	}

	if !v.ValidationFunc(actStr) {
		return fmt.Errorf("field at %s value %q is not valid", path, actStr)
	}

	// Replace with normalized placeholder
	act[key] = v.Placeholder
	exp[key] = v.Placeholder
	return nil
}

// buildPath constructs a JSON path string by appending a component to an
// existing path, handling array indices appropriately.
func buildPath(path, component string) string {
	if path == "" {
		return component
	}
	if strings.HasPrefix(component, "[") {
		return path + component
	}
	return path + "." + component
}

// extractValidationType extracts the validation type from a dynamic placeholder
// string (e.g., "{{DYNAMIC_TIMESTAMP}}" returns "TIMESTAMP").
func extractValidationType(placeholder string) string {
	return strings.TrimSuffix(strings.TrimPrefix(placeholder, "{{DYNAMIC_"), "}}")
}

// shouldRecurse determines whether validation should recurse into nested
// structures based on the expected value type.
func shouldRecurse(actValue, expValue interface{}) bool {
	if actValue == nil || expValue == nil {
		return false
	}

	switch expValue.(type) {
	case map[string]interface{}, []interface{}:
		return true
	}
	return false
}

// iHaveAHealthCheckIntervalOfSecond sets healthcheck interval and critical timeout
func (f *APIFeature) iHaveAHealthCheckIntervalOfSecond(healthCheckInterval int) error {
	f.HealthCheckInterval = time.Duration(healthCheckInterval)

	return f.StepError()
}

// theHealthChecksShouldHaveCompletedWithinSeconds sets the expected healthcheck response time
func (f *APIFeature) theHealthChecksShouldHaveCompletedWithinSeconds(expectedResponseTime int) error {
	f.ExpectedResponseTime = time.Duration(expectedResponseTime)

	return f.StepError()
}

// iShouldReceiveTheFollowingHealthJSONResponse asserts the health response and body match the expectation
func (f *APIFeature) iShouldReceiveTheFollowingHealthJSONResponse(expectedResponse *godog.DocString) error {
	var healthResponse, expectedHealth HealthCheckTest

	responseBody, err := io.ReadAll(f.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read health response - error: %v", err)
	}

	err = json.Unmarshal(responseBody, &healthResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal health response - error: %v", err)
	}

	err = json.Unmarshal([]byte(expectedResponse.Content), &expectedHealth)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected health response - error: %v", err)
	}

	f.validateHealthCheckResponse(healthResponse, expectedHealth)

	return f.StepError()
}

func (f *APIFeature) validateHealthCheckResponse(healthResponse, expectedResponse HealthCheckTest) {
	maxExpectedStartTime := f.StartTime.Add((f.HealthCheckInterval + 1) + time.Second)

	assert.Equal(&f.ErrorFeature, expectedResponse.Status, healthResponse.Status)
	assert.True(&f.ErrorFeature, healthResponse.StartTime.Before(maxExpectedStartTime.UTC()))
	assert.Greater(&f.ErrorFeature, healthResponse.Uptime.Seconds(), float64(0))

	f.validateHealthVersion(healthResponse.Version, expectedResponse.Version, maxExpectedStartTime.UTC())

	for i, checkResponse := range healthResponse.Checks {
		f.validateHealthCheck(checkResponse, expectedResponse.Checks[i])
	}
}

func (f *APIFeature) validateHealthVersion(versionResponse, expectedVersion healthcheck.VersionInfo, maxExpectedStartTime time.Time) {
	assert.True(&f.ErrorFeature, versionResponse.BuildTime.Before(maxExpectedStartTime))
	assert.Equal(&f.ErrorFeature, expectedVersion.GitCommit, versionResponse.GitCommit)
	assert.Equal(&f.ErrorFeature, expectedVersion.Language, versionResponse.Language)
	assert.NotEmpty(&f.ErrorFeature, versionResponse.LanguageVersion)
	assert.Equal(&f.ErrorFeature, expectedVersion.Version, versionResponse.Version)
}

func (f *APIFeature) validateHealthCheck(checkResponse, expectedCheck *Check) {
	maxExpectedHealthCheckTime := f.StartTime.Add(f.ExpectedResponseTime * time.Second)

	assert.Equal(&f.ErrorFeature, expectedCheck.Name, checkResponse.Name)
	assert.Equal(&f.ErrorFeature, expectedCheck.Status, checkResponse.Status)
	assert.Equal(&f.ErrorFeature, expectedCheck.StatusCode, checkResponse.StatusCode)
	assert.Equal(&f.ErrorFeature, expectedCheck.Message, checkResponse.Message)
	assert.True(&f.ErrorFeature, checkResponse.LastChecked.Before(maxExpectedHealthCheckTime.UTC()))
	assert.True(&f.ErrorFeature, checkResponse.LastChecked.After(f.StartTime))

	if expectedCheck.Status == healthcheck.StatusOK {
		lastSuccess := checkResponse.LastSuccess

		if lastSuccess != nil {
			assert.True(&f.ErrorFeature, lastSuccess.Before(maxExpectedHealthCheckTime.UTC()))
			assert.True(&f.ErrorFeature, lastSuccess.After(f.StartTime))
		} else {
			assert.Fail(&f.ErrorFeature, "last success should not be nil")
		}
	} else {
		lastFailure := checkResponse.LastFailure

		if lastFailure != nil {
			assert.True(&f.ErrorFeature, lastFailure.Before(maxExpectedHealthCheckTime.UTC()))
			assert.True(&f.ErrorFeature, lastFailure.After(f.StartTime))
		} else {
			assert.Fail(&f.ErrorFeature, "last failure should not be nil")
		}
	}
}

// delayTimeBySeconds pauses the goroutine for the given seconds
// WARNING: This should not be used where other 'waits' would be possible. Use this only where absolutely necessary.
func (f *APIFeature) delayTimeBySeconds(sec int) error {
	time.Sleep(time.Duration(int64(sec)) * time.Second)
	return nil
}
