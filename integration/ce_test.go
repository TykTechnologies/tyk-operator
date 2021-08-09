package integration

import "testing"

func TestCommunity(t *testing.T) {
	if !isCE() {
		t.Skip("skipping community edition tests")
	}
}
