package componenttest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
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
// This also validates any dynamic fields.
func (f *APIFeature) IShouldReceiveTheFollowingJSONResponse(expectedAPIResponse *godog.DocString) error {
	responseBody := f.HTTPResponse.Body
	body, err := io.ReadAll(responseBody)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	var actual, expected map[string]interface{}
	if err := json.Unmarshal(body, &actual); err != nil {
		return fmt.Errorf("error unmarshalling actual response body: %w", err)
	}
	if err := json.Unmarshal([]byte(expectedAPIResponse.Content), &expected); err != nil {
		return fmt.Errorf("error unmarshalling expected response body: %w", err)
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

type DynamicValidator struct {
	ValidationFunc func(value string) bool
	Placeholder    string
}

func (v DynamicValidator) Validate(coord string, actual, expected map[string]interface{}) (actualValidated, expectedValidated map[string]interface{}, err error) {
	actualFieldValue, err := getNestedValueByCoords(actual, coord)
	if err != nil {
		return nil, nil, err
	}

	validValue := v.ValidationFunc(actualFieldValue.(string))
	if !validValue {
		return nil, nil, fmt.Errorf("field %q value %q is not a valid value", coord, actualFieldValue)
	}

	actualValidated, err = setNestedValueByCoords(actual, coord, v.Placeholder)
	if err != nil {
		return nil, nil, err
	}
	expectedValidated, err = setNestedValueByCoords(expected, coord, v.Placeholder)
	if err != nil {
		return nil, nil, err
	}

	return actualValidated, expectedValidated, nil
}

type DynamicValueType string

const (
	ValidPrefix                             = "VALID"
	DynamicTimestamp       DynamicValueType = "TIMESTAMP"
	DynamicRecentTimestamp DynamicValueType = "RECENT_TIMESTAMP"
	DynamicURIPath         DynamicValueType = "URI_PATH"
	DynamicURL             DynamicValueType = "URL"
	DynamicUUID            DynamicValueType = "UUID"
)

var dynamicValidators = map[DynamicValueType]DynamicValidator{
	DynamicTimestamp: {
		ValidationFunc: validator.ValidateTimestamp,
		Placeholder:    fmt.Sprintf("%s_%s", ValidPrefix, DynamicTimestamp),
	},
	DynamicRecentTimestamp: {
		ValidationFunc: validator.ValidateRecentTimestamp,
		Placeholder:    fmt.Sprintf("%s_%s", ValidPrefix, DynamicRecentTimestamp),
	},
	DynamicURIPath: {
		ValidationFunc: validator.ValidateURIPath,
		Placeholder:    fmt.Sprintf("%s_%s", ValidPrefix, DynamicURIPath),
	},
	DynamicURL: {
		ValidationFunc: validator.ValidateURL,
		Placeholder:    fmt.Sprintf("%s_%s", ValidPrefix, DynamicURL),
	},
	DynamicUUID: {
		ValidationFunc: validator.ValidateUUID,
		Placeholder:    fmt.Sprintf("%s_%s", ValidPrefix, DynamicUUID),
	},
}

type DynamicValue struct {
	Coord string
	Type  DynamicValueType
}

// validateDynamicValues checks for any fields in expected with dynamic value strings, e.g. "{{DYNAMIC_TIMESTAMP}}", validates them and replaces
// them with a placeholder.
func validateDynamicValues(actual, expected string) (actualValidated, expectedValidated string, err error) {
	var parsedActual map[string]interface{}
	var parsedExpected map[string]interface{}
	if err := json.Unmarshal([]byte(actual), &parsedActual); err != nil {
		return "", "", fmt.Errorf("error unmarshalling actual response: %w", err)
	}
	if err := json.Unmarshal([]byte(expected), &parsedExpected); err != nil {
		return "", "", fmt.Errorf("error unmarshalling expected response: %w", err)
	}

	var dynamicValues []DynamicValue
	findDynamicValueCoordinates(parsedExpected, "", &dynamicValues)

	// For each field with a dynamic value, validate and replace in both actual and expected
	for _, dynamicValue := range dynamicValues {
		validatorFunc, exists := dynamicValidators[dynamicValue.Type]
		if !exists {
			return "", "", fmt.Errorf("unknown validation type: %s", dynamicValue.Type)
		}

		parsedActual, parsedExpected, err = validatorFunc.Validate(dynamicValue.Coord, parsedActual, parsedExpected)
		if err != nil {
			return "", "", err
		}
	}

	actualBytes, err := json.Marshal(parsedActual)
	if err != nil {
		return "", "", fmt.Errorf("error marshalling validated actual: %w", err)
	}
	expectedBytes, err := json.Marshal(parsedExpected)
	if err != nil {
		return "", "", fmt.Errorf("error marshalling validated expected: %w", err)
	}
	return string(actualBytes), string(expectedBytes), nil
}

// findDynamicValueCoordinates recursively searches a nested map/array structure for dynamic value placeholders
// and appends their coordinates and types to the results slice.
func findDynamicValueCoordinates(m map[string]interface{}, prefix string, results *[]DynamicValue) {
	var dynamicPattern = regexp.MustCompile(`^\{\{DYNAMIC_([A-Z_0-9]+)\}\}$`)

	for key, value := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// Check the type of the value
		switch val := value.(type) {
		// If object, recurse
		case map[string]interface{}:
			findDynamicValueCoordinates(val, fullKey, results)
		// If array, check each item
		case []interface{}:
			for i, item := range val {
				if subMap, ok := item.(map[string]interface{}); ok {
					findDynamicValueCoordinates(subMap, fmt.Sprintf("%s[%d]", fullKey, i), results)
				} else if str, ok := item.(string); ok && dynamicPattern.MatchString(str) {
					*results = append(*results, DynamicValue{Coord: fmt.Sprintf("%s[%d]", fullKey, i), Type: DynamicValueType(dynamicPattern.FindStringSubmatch(str)[1])})
				}
			}
		// If string, check for dynamic pattern
		default:
			if reflect.TypeOf(val).Kind() == reflect.String && dynamicPattern.MatchString(val.(string)) {
				*results = append(*results, DynamicValue{Coord: fullKey, Type: DynamicValueType(dynamicPattern.FindStringSubmatch(val.(string))[1])})
			}
		}
	}
}

// traversePath traverses a nested map/array structure and returns the parent, key/index, and value at the path
// in JSON dot notation (e.g. "data.items[0].id")
func traversePath(root interface{}, coords string) (parent, key, value interface{}) {
	parts := strings.Split(coords, ".")
	var val = root
	var prev interface{}
	var k interface{}
	for i, part := range parts {
		// Check for array index
		if idxStart := strings.Index(part, "["); idxStart != -1 && strings.HasSuffix(part, "]") {
			keyStr := part[:idxStart]
			idxStr := part[idxStart+1 : len(part)-1]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil, nil, nil
			}
			if m2, ok := val.(map[string]interface{}); ok {
				arr, ok := m2[keyStr].([]interface{})
				if !ok || idx >= len(arr) {
					return nil, nil, nil
				}
				prev = arr
				k = idx
				val = arr[idx]
			} else {
				return nil, nil, nil
			}
		} else {
			// Check if object
			if m2, ok := val.(map[string]interface{}); ok {
				prev = m2
				k = part
				val = m2[part]
			} else {
				// No longer traversable
				return nil, nil, nil
			}
		}
		// If last part, return parent, key, value
		if i == len(parts)-1 {
			return prev, k, val
		}
	}
	// Didn't find coords
	return nil, nil, nil
}

// getNestedValueByCoords retrieves a value from a nested map/array structure at the specified path
// in JSON dot notation (e.g. "data.items[0].id")
func getNestedValueByCoords(m map[string]interface{}, coords string) (interface{}, error) {
	_, _, value := traversePath(m, coords)
	if value == nil {
		return nil, fmt.Errorf("value not found for coords: %s", coords)
	}
	return value, nil
}

// setNestedValueByCoords sets a value in a nested map/array structure at the specified path
// in JSON dot notation (e.g. "data.items[0].id")
func setNestedValueByCoords(m map[string]interface{}, coords string, value interface{}) (map[string]interface{}, error) {
	parent, key, _ := traversePath(m, coords)
	// Check the parent type
	switch p := parent.(type) {
	// if object
	case map[string]interface{}:
		if k, ok := key.(string); ok {
			p[k] = value
			return m, nil
		}
	// if array
	case []interface{}:
		if idx, ok := key.(int); ok && idx < len(p) {
			p[idx] = value
			return m, nil
		}
	}
	return m, fmt.Errorf("could not set value for coords: %s", coords)
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
