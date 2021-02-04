package featuretest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

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
	Initialiser       ServiceInitialiser
	httpResponse      *http.Response
	BeforeRequestHook func() error
	requestHeaders    map[string]string
}

// NewAPIFeature returns a new APIFeature, takes a function to retrieve the bound handler just before a request is made
func NewAPIFeature(initialiser ServiceInitialiser) *APIFeature {
	return &APIFeature{
		Initialiser:    initialiser,
		requestHeaders: make(map[string]string),
	}
}

// NewAPIFeatureWithHandler create a new APIFeature with a handler already bound with your endpoints
func NewAPIFeatureWithHandler(handler http.Handler) *APIFeature {
	return NewAPIFeature(StaticHandler(handler))
}

// Reset the request headers
func (f *APIFeature) Reset() {
	f.requestHeaders = make(map[string]string)
}

// RegisterSteps binds the APIFeature steps to the godog context to enable usage in the feature tests
func (f *APIFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I set the "([^"]*)" header to "([^"]*)"$`, f.ISetTheHeaderTo)
	ctx.Step(`^I am authorised$`, f.IAmAuthorised)
	ctx.Step(`^I am not authorised$`, f.IAmNotAuthorised)
	ctx.Step(`^I GET "([^"]*)"$`, f.IGet)
	ctx.Step(`^I POST "([^"]*)"$`, f.IPostToWithBody)
	ctx.Step(`^the HTTP status code should be "([^"]*)"$`, f.TheHTTPStatusCodeShouldBe)
	ctx.Step(`^the response header "([^"]*)" should be "([^"]*)"$`, f.TheResponseHeaderShouldBe)
	ctx.Step(`^I should receive the following response:$`, f.IShouldReceiveTheFollowingResponse)
	ctx.Step(`^I should receive the following JSON response:$`, f.IShouldReceiveTheFollowingJSONResponse)
	ctx.Step(`^I should receive the following JSON response with status "([^"]*)":$`, f.IShouldReceiveTheFollowingJSONResponseWithStatus)
}

// ISetTheHeaderTo is a default step used to set a header and associated value for the next request
func (f *APIFeature) ISetTheHeaderTo(header, value string) error {
	f.requestHeaders[header] = value
	return nil
}

// IAmAuthorised sets the Authorization header to a bogus token
func (f *APIFeature) IAmAuthorised() error {
	f.ISetTheHeaderTo("Authorization", "bearer SomeFakeToken")
	return nil
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

	f.httpResponse = w.Result()
	return nil
}

// IShouldReceiveTheFollowingResponse asserts the response body and expected response body are equal
func (f *APIFeature) IShouldReceiveTheFollowingResponse(expectedAPIResponse *godog.DocString) error {
	responseBody := f.httpResponse.Body
	body, _ := ioutil.ReadAll(responseBody)

	assert.Equal(f, strings.TrimSpace(expectedAPIResponse.Content), strings.TrimSpace(string(body)))

	return f.StepError()
}

// IShouldReceiveTheFollowingJSONResponse asserts that the response body and expected response body are equal
func (f *APIFeature) IShouldReceiveTheFollowingJSONResponse(expectedAPIResponse *godog.DocString) error {
	responseBody := f.httpResponse.Body
	body, _ := ioutil.ReadAll(responseBody)

	assert.JSONEq(f, expectedAPIResponse.Content, string(body))

	return f.StepError()
}

// TheHTTPStatusCodeShouldBe asserts that the status code of the response matches the expected code
func (f *APIFeature) TheHTTPStatusCodeShouldBe(expectedCodeStr string) error {
	expectedCode, err := strconv.Atoi(expectedCodeStr)
	if err != nil {
		return err
	}
	assert.Equal(f, expectedCode, f.httpResponse.StatusCode)
	return f.StepError()
}

// TheResponseHeaderShouldBe asserts the response header matches the expectation
func (f *APIFeature) TheResponseHeaderShouldBe(headerName, expectedValue string) error {
	assert.Equal(f, expectedValue, f.httpResponse.Header.Get(headerName))
	return f.StepError()
}

// IShouldReceiveTheFollowingJSONResponseWithStatus asserts the response code and body match the expectation
func (f *APIFeature) IShouldReceiveTheFollowingJSONResponseWithStatus(expectedCodeStr string, expectedBody *godog.DocString) error {
	if err := f.TheHTTPStatusCodeShouldBe(expectedCodeStr); err != nil {
		return err
	}
	if err := f.TheResponseHeaderShouldBe("Content-Type", "application/json"); err != nil {
		return err
	}
	return f.IShouldReceiveTheFollowingJSONResponse(expectedBody)
}
