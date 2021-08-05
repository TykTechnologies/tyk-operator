package integration

import (
	"testing"
)

func TestPro(t *testing.T) {
	if string(e.Mode) != "pro" {
		t.Skip("Skipping pro tests")
	}
}
