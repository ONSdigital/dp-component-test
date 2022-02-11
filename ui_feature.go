package componenttest

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// Chrome contains Chrome session-related resources
type Chrome struct {
	ExecAllocatorCanceller context.CancelFunc
	CtxCanceller           context.CancelFunc
	ctx                    context.Context
}

// UIFeature contains the information needed to test UI interactions
type UIFeature struct {
	ErrorFeature
	BaseURL     string
	Chrome      Chrome
	waitTimeOut time.Duration
}

// NewUIFeature returns a new UIFeature configured with baseUrl
func NewUIFeature(baseUrl string) *UIFeature {
	f := &UIFeature{
		BaseURL:     baseUrl,
		waitTimeOut: 10 * time.Second,
	}

	return f
}

func (f *UIFeature) setChromeContext() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		// set this to false to be able to watch the browser in action
		chromedp.Flag("headless", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	f.Chrome.ExecAllocatorCanceller = cancel
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	f.Chrome.CtxCanceller = cancel
	log.Print("re-starting chrome ...")
	f.Chrome.ctx = ctx
}

// Reset the chrome context
func (f *UIFeature) Reset() {
	f.setChromeContext()
}

// Close Chrome
func (f *UIFeature) Close() {
	f.Chrome.CtxCanceller()
	f.Chrome.ExecAllocatorCanceller()
}

// RegisterSteps binds the APIFeature steps to the godog context to enable usage in the component tests
func (f *UIFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I navigate to "([^"]*)"`, f.iNavigateTo)
	ctx.Step(`^element "([^"]*)" should be visible$`, f.elementShouldBeVisible)
	ctx.Step(`^the beta phase banner should be visible$`, f.theBetaBannerShouldBeVisible)
	ctx.Step(`^the improve this page banner should be visible$`, f.theImproveThisPageBannerShouldBeVisible)
	ctx.Step(`^the page should have the following content$`, f.thePageShouldHaveTheFollowingContent)
}

func (f *UIFeature) iNavigateTo(route string) error {
	err := chromedp.Run(f.Chrome.ctx,
		chromedp.Navigate(f.BaseURL+route),
	)
	if err != nil {
		return f.StepError()
	}

	return nil
}

func (f *UIFeature) elementShouldBeVisible(elementSelector string) error {
	err := chromedp.Run(f.Chrome.ctx,
		f.runWithTimeOut(&f.Chrome.ctx, f.waitTimeOut, chromedp.Tasks{
			chromedp.WaitVisible(elementSelector),
		}),
	)
	assert.Nil(f, err)

	return f.StepError()
}

func (f *UIFeature) theBetaBannerShouldBeVisible() error {
	return f.elementShouldBeVisible(".ons-phase-banner")
}

func (f *UIFeature) theImproveThisPageBannerShouldBeVisible() error {
	return f.elementShouldBeVisible(".improve-this-page")
}

func (f *UIFeature) thePageShouldHaveTheFollowingContent(expectedAPIResponse *godog.DocString) error {
	var contentElements map[string]string

	err := json.Unmarshal([]byte(expectedAPIResponse.Content), &contentElements)
	if err != nil {
		return err
	}

	for selector, expectedContent := range contentElements {
		var actualContent string
		err = chromedp.Run(f.Chrome.ctx,
			f.runWithTimeOut(&f.Chrome.ctx, f.waitTimeOut, chromedp.Tasks{
				chromedp.Text(selector, &actualContent, chromedp.NodeVisible),
			}),
		)

		if err != nil {
			return err
		}

		assert.Equal(f, expectedContent, actualContent)
	}

	return f.StepError()
}

func (f *UIFeature) runWithTimeOut(ctx *context.Context, timeout time.Duration, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		timeoutContext, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return tasks.Do(timeoutContext)
	}
}
