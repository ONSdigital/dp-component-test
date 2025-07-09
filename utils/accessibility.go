// revive:disable:var-naming
package utils // ignoring revive var-naming temporarily till a more permanent fix is implemented

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
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

func (v *Violation) toErrorMessage() string {
	return v.ID + ": " + v.Description + " (" + strconv.Itoa(len(v.Nodes)) + " violations)"
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

type Rule struct {
	Enabled bool `json:"enabled"`
}

// Exports the Accessibility config as a JSON string to supply to axe-core
func (cfg *AccessibilityConfig) JSON() (string, error) {
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(cfgJSON), nil
}

func provisionAccessibilityTooling(ctx context.Context) error {
	if !accessibilityToolsExist(ctx) {
		return addAccessibilityTools(ctx)
	}
	return nil
}

func addAccessibilityTools(ctx context.Context) error {
	var buf []byte

	//nolint:gosec // test code only and no way to exploit.
	provisionScript := template.JS(fmt.Sprintf(`
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
	`, axeVersion))

	provisionScriptString := string(provisionScript)

	if err := chromedp.Run(ctx,
		chromedp.Evaluate(provisionScriptString, &buf),
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

	//nolint:gosec // test code only and no way to exploit.
	testScript := template.JS(fmt.Sprintf(`
	 	window.returnValue = axe
			.run(%s)
			.then(results => {
				return results.violations
			})
			.catch(err => {
				return err.message
			});
	`, cfgJSON))

	testScriptString := string(testScript)

	if err := chromedp.Run(ctx,
		chromedp.Evaluate(testScriptString, &buf, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
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
			errMessage += "\n" + violations[i].toErrorMessage()
		}
	}

	return errMessage
}
