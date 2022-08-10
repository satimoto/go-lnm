package mocks

import (
	"testing"

	"github.com/nsf/jsondiff"
)

var (
	diffOpts = jsondiff.DefaultJSONOptions()
)

func CompareJson(t *testing.T, actualBytes []byte, expectedBytes []byte) {
	CompareJsonWithDifference(t, actualBytes, expectedBytes, jsondiff.FullMatch)
}

func CompareJsonWithDifference(t *testing.T, actualBytes []byte, expectedBytes []byte, difference jsondiff.Difference) {
	diffRes, diff := jsondiff.Compare(actualBytes, expectedBytes, &diffOpts)

	if diffRes != difference {
		t.Errorf("JSON mismatch: %v", diff)
		t.Logf("JSON: %v", string(actualBytes))
	}
}
