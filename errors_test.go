package pushover

import (
	"testing"
)

// TestErrorsString tests the custom error string
func TestErrorsString(t *testing.T) {
	e := &Errors{"error1", "error2"}
	got := e.Error()
	expected := "Errors:\nerror1\nerror2"

	if got != expected {
		t.Errorf("invalid error string\ngot:\n%s\nexpected:\n%s\n", got, expected)
	}
}
