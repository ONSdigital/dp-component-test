package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

const axeVersion = "4.10.0"

type Violation struct {
	ID          string          `json:"id"`
	Impact      string          `json:"impact"`
	Tags        []string        `json:"tags"`
	Description string          `json:"description"`
	Help        string          `json:"help"`
	HelpURL     string          `json:"helpUrl"`
	Nodes       []ViolationNode `json:"nodes"`
}

type ViolationNode struct {
	Impact string `json:"impact"`
	HTML   string `json:"html"`
}

type AccessibilityConfig struct {
	Rules   map[string]Rule `json:"rules,omitempty"`
	RunOnly RunOnly         `json:"runOnly,omitempty"`
}

type RunOnly struct {
	Type   string   `json:"type,omitempty"`
	Values []string `json:"values,omitempty"`
}

func (cfg *AccessibilityConfig) JSON() (string, error) {
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(cfgJSON), nil
}

type Rule struct {
	Enabled bool `json:"enabled"`
}

func provisionAccessibilityTooling(ctx context.Context) error {
	if !accessibilityToolsExist(ctx) {
		return addAccessibilityTools(ctx)
	}
	return nil
}

func addAccessibilityTools(ctx context.Context) error {
	var buf []byte

	provisionScript := fmt.Sprintf(`
		(function(d, script) {
			script = d.createElement('script');
			script.type = 'text/javascript';
			script.async = true;
			script.onload = function(){
				console.log("axe loaded");
			};
			script.src = 'https://cdnjs.cloudflare.com/ajax/libs/axe-core/%s/axe.min.js';
			d.getElementsByTagName('head')[0].appendChild(script);
		}(document));
	`, axeVersion)

	if err := chromedp.Run(ctx,
		chromedp.Evaluate(provisionScript, &buf),
	); err != nil {
		return fmt.Errorf("failed to evaluate script to add accessibility tooling")
	}

	// TODO: await onload trigger
	time.Sleep(2 * time.Second)

	return nil
}

func accessibilityToolsExist(ctx context.Context) bool {
	var val int

	checkToolingScript := `
		axe
	`

	if err := chromedp.Run(ctx,
		chromedp.Evaluate(checkToolingScript, &val),
	); err != nil {
		return false
	}

	return true
}

func RunTestWithConfig(ctx context.Context, cfg AccessibilityConfig) ([]Violation, string, error) {
	err := provisionAccessibilityTooling(ctx)
	if err != nil {
		return nil, "", err
	}

	var buf []byte

	// default to WCAG2.X A + AA // TODO: allow this to be overrideable
	cfg.RunOnly = RunOnly{
		Type:   "tag",
		Values: []string{"wcag2a", "wcag2aa", "wcag21a", "wcag21aa"},
	}

	cfgJSON, err := cfg.JSON()
	if err != nil {
		return nil, "", err
	}

	testScript := fmt.Sprintf(`
	 	window.returnValue = axe
			.run(%s)
			.then(results => {
				return results.violations
			})
			.catch(err => {
				return err.message
			});
	`, cfgJSON)

	if err := chromedp.Run(ctx,
		chromedp.Evaluate(testScript, &buf, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	); err != nil {
		return nil, "", fmt.Errorf("error occured running accessibility script")
	}

	var violations []Violation

	err = json.Unmarshal(buf, &violations)
	if err != nil {
		return nil, "", err
	}

	message := generateViolationMessage(violations)

	return violations, message, nil
}

func RunTest(ctx context.Context) ([]Violation, string, error) {
	return RunTestWithConfig(ctx, AccessibilityConfig{})
}

func generateViolationMessage(violations []Violation) string {
	var errMessage string

	if len(violations) > 0 {
		errMessage = "The following accessibility rules have been violated:"

		// TODO: we should add the HTML snippets in here.
		for i := range violations {
			errMessage += "\n" + violations[i].ID + ": " + violations[i].Description + " (" + strconv.Itoa(len(violations[i].Nodes)) + " violations)"
		}
	}

	return errMessage
}
