package v1alpha1

import (
	"regexp"
	"testing"
)

const (
	// PercentPattern is a regular expression to match a percentage value.
	PercentPattern = "^(?:[+]?(?:0))?(?:\\.[0-9]*)?(?:[eE][\\+\\-]?(?:[0-9]+))?$"
)

func TestPercent(t *testing.T) {
	r := regexp.MustCompile(PercentPattern)
	s := []struct {
		v  string
		ok bool
	}{
		// must be always positive
		{"-0.5", false},
		{"0.5", true},

		// must be <1
		{"1.5", false},
		{"2.5", false},
	}

	for i, v := range s {
		m := r.MatchString(v.v)
		if m != v.ok {
			t.Errorf("%d %q is not valid expected %v got %v", i, v.v, v.ok, m)
		}
	}
}
