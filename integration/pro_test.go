package integration

import (
	"testing"
)

func TestPro(t *testing.T) {
	if !isPro() {
		t.Skip("Skipping pro tests")
	}
}
