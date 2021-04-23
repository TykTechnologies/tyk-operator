package controllers

import (
	"testing"

	"gotest.tools/assert"
	"sigs.k8s.io/gateway-api/apis/v1alpha1"
)

func TestGatewayApiReconciler_HostsToCustomDomains(t *testing.T) {
	tests := map[string]struct {
		hostname  v1alpha1.Hostname
		want      string
		wantError error
		skip      bool
	}{
		"no hosts (match all)": {
			hostname:  v1alpha1.Hostname("*"),
			want:      "",
			wantError: nil,
		},
		"single host": {
			hostname:  v1alpha1.Hostname("foo.bar.com"),
			want:      "foo.bar.com",
			wantError: nil,
		},
		"wildcard host": {
			hostname:  v1alpha1.Hostname("*.bar.com"),
			want:      "{?:[^.]+}.bar.com",
			wantError: nil,
			skip:      false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(innerTest *testing.T) {
			if tc.skip {
				innerTest.Skip()
			}
			c := GatewayApiReconciler{}
			customDomain, err := c.hostToCustomDomain(tc.hostname)
			assert.Equal(innerTest, tc.want, customDomain)
			assert.Equal(innerTest, tc.wantError, err)
		})
	}
}
