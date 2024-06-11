package componenttest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// Chrome contains Chrome session-related resources
type Chrome struct {
	ExecAllocatorCanceller context.CancelFunc
	CtxCanceller           context.CancelFunc
	Ctx                    context.Context
}

// UIFeature contains the information needed to test UI interactions
type UIFeature struct {
	ErrorFeature
	BaseURL     string
	Chrome      Chrome
	WaitTimeOut time.Duration
}

// NewUIFeature returns a new UIFeature configured with baseURL
func NewUIFeature(baseURL string) *UIFeature {
	f := &UIFeature{
		BaseURL:     baseURL,
		WaitTimeOut: 10 * time.Second,
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
	f.Chrome.Ctx = ctx
}

// Reset the chrome context
func (f *UIFeature) Reset() {
	f.ErrorFeature.Reset()
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
	ctx.Step(`^element "([^"]*)" should not be visible$`, f.elementShouldNotBeVisible)
	ctx.Step(`^input element "([^"]*)" has value "([^"]*)"`, f.inputElementHasValue)
	ctx.Step(`^the page should have the following content$`, f.thePageShouldHaveTheFollowingContent)
	ctx.Step(`^the page should contain "([^"]*)" with list element text "([^"]*)" at (\d+) depth$`, f.innerListElementsShouldHaveText)
	ctx.Step(`^I fill in input element "([^"]*)" with value "([^"]*)"$`, f.iFillInInputElementWithValue)
	ctx.Step(`^I click the "([^"]*)" element$`, f.iClickElement)
}

func (f *UIFeature) iNavigateTo(route string) error {
	err := chromedp.Run(f.Chrome.Ctx,
		chromedp.Navigate(f.BaseURL+route),
	)
	if err != nil {
		return f.StepError()
	}

	return nil
}

func (f *UIFeature) elementShouldBeVisible(elementSelector string) error {
	err := chromedp.Run(f.Chrome.Ctx,
		f.RunWithTimeOut(f.WaitTimeOut, chromedp.Tasks{
			chromedp.WaitVisible(elementSelector),
		}),
	)
	assert.Nil(f, err)

	return f.StepError()
}

func (f *UIFeature) elementShouldNotBeVisible(elementSelector string) error {
	err := chromedp.Run(f.Chrome.Ctx,
		f.RunWithTimeOut(f.WaitTimeOut, chromedp.Tasks{
			chromedp.WaitNotPresent(elementSelector),
		}),
	)
	if err != nil {
		return err
	}

	return f.StepError()
}

func (f *UIFeature) inputElementHasValue(elementSelector, expectedValue string) error {
	var actualValue string

	err := chromedp.Run(f.Chrome.Ctx,
		f.RunWithTimeOut(f.WaitTimeOut, chromedp.Tasks{
			chromedp.WaitVisible(elementSelector),
			chromedp.Value(elementSelector, &actualValue),
		}),
	)
	if err != nil {
		return err
	}

	assert.Equal(f, expectedValue, actualValue)

	return f.StepError()
}

func (f *UIFeature) innerListElementsShouldHaveText(dataAttr, textList string, depth int) (err error) {
	var (
		elementSelector = fmt.Sprintf("[data-test='%s']", dataAttr)
		nodes           []*cdp.Node
		entireSubtree   = -1
		didMatch        = false
	)
	textSlices := strings.Split(textList, ",")
	err = chromedp.Run(f.Chrome.Ctx,
		f.RunWithTimeOut(f.WaitTimeOut, chromedp.Tasks{
			chromedp.Nodes(elementSelector, &nodes, chromedp.ByQuery),
			chromedp.ActionFunc(func(c context.Context) error {
				return dom.RequestChildNodes(nodes[0].NodeID).WithDepth(int64(entireSubtree)).Do(c)
			}),
			chromedp.Sleep(time.Second),
			chromedp.ActionFunc(func(_ context.Context) error {
				// All examples use depth = 3
				// Home - nodes[0].Children[0].Children[0].Children[0].Children[0].Children[0].NodeValue
				// Areas - nodes[0].Children[0].Children[0].Children[1].Children[0].Children[0].NodeValue
				// Get the point where the node branch splits based on the depth
				var currNode *cdp.Node
				currNode = nodes[0]
				for i := 0; i < depth-1; i++ {
					currNode = currNode.Children[0]
				}
				// eg. Areas - currNode.Children[1].Children[0].Children[0].NodeValue
				// At this point we have the split so transcend the node branch
				// if the end expected value exists update didMatch pointer to true
				for ii, node := range currNode.Children {
					didMatch = false
					// only loop over the slices we care about & not the len(currNode.Children) value
					if ii <= len(textSlices)-1 {
						getName(node, textSlices[ii], &didMatch)
						if !didMatch {
							return errors.New("no match for " + textSlices[ii])
						}
					}
				}
				return nil
			}),
		}),
	)
	return err
}

func (f *UIFeature) thePageShouldHaveTheFollowingContent(expectedAPIResponse *godog.DocString) error {
	var contentElements map[string]string

	err := json.Unmarshal([]byte(expectedAPIResponse.Content), &contentElements)
	if err != nil {
		return err
	}

	for selector, expectedContent := range contentElements {
		var actualContent string
		err = chromedp.Run(f.Chrome.Ctx,
			f.RunWithTimeOut(f.WaitTimeOut, chromedp.Tasks{
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

func (f *UIFeature) RunWithTimeOut(timeout time.Duration, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		timeoutContext, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return tasks.Do(timeoutContext)
	}
}

// ---------------------------------------
// Utils
func getName(node *cdp.Node, expected string, didMatch *bool) {
	// `didMatch` must be defaulted to false
	// e.g breadcrumb text node: node.Children[0].Children[0].Children[0].NodeValue
	if len(node.Children) == 0 {
		// End of node branch so check expected
		if node.NodeValue == expected {
			*didMatch = true
		}
		return
	}
	getName(node.Children[0], expected, didMatch)
}

func (f *UIFeature) iFillInInputElementWithValue(fieldSelector, value string) error {
	jsScript := fmt.Sprintf(`document.querySelector('%s').value = '%s';`, fieldSelector, value)

	err := chromedp.Run(f.Chrome.Ctx,
		chromedp.Evaluate(jsScript, nil),
	)
	if err != nil {
		return err
	}

	return f.StepError()
}

func (f *UIFeature) iClickElement(buttonSelector string) error {
	// if this doesn't work as expected, you might need a sleep after the click
	err := chromedp.Run(f.Chrome.Ctx,
		chromedp.Click(buttonSelector),
	)
	if err != nil {
		return err
	}

	return f.StepError()
}
