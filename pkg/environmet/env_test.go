package environmet

import "testing"

func TestIsZero(t *testing.T) {
	var e Env
	if !e.IsZero() {
		t.Fatal("expected to be true")
	}
	e.Auth = "foo"
	if e.IsZero() {
		t.Fatal("expected to be false")
	}
}
