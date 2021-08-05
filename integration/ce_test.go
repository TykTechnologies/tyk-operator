package integration

import "testing"

func TestCommunity(t *testing.T) {
	if string(e.Mode) != "ce" {
		t.Skip("skipping community edition tests")
	}
}
