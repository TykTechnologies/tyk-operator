package model

import (
	"regexp"
	"testing"
)

func TestPercent(t *testing.T) {
	r := regexp.MustCompile(`^0\.\d+|1\.0$`)
	s := []struct {
		v  string
		ok bool
	}{
		// must be always positive
		{"-0.5", false},
		{"0.5", true},

		// must be <=1
		{"1.5", false},
		{"2.5", false},
		{"1.0", true},
	}

	for i, v := range s {
		m := r.MatchString(v.v)
		if m != v.ok {
			t.Errorf("%d %q is not valid expected %v got %v", i, v.v, v.ok, m)
		}
	}
}
