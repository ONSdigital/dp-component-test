package componenttest

import (
	"fmt"
	"testing"
)

type ErrorFeature struct {
	testing.TB
	// composing testing.TB allows ErrorFeature to be passed off as a testing.T thingy
	err error
}

func (t *ErrorFeature) Log(_ ...interface{}) {
	// stub to mimic testing.T
}

func (t *ErrorFeature) Logf(format string, args ...interface{}) {
	t.err = fmt.Errorf(format, args...)
}

func (t *ErrorFeature) Errorf(format string, args ...interface{}) {
	t.err = fmt.Errorf(format, args...)
}

func (t *ErrorFeature) StepError() error {
	return t.err
}

func (t *ErrorFeature) Reset() {
	t.err = nil
}

func (t *ErrorFeature) Helper() {
	// stub to mimic testing.T
}

func (t *ErrorFeature) Name() string {
	return "Component Test"
}
