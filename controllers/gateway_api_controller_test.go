package controllers

import (
	"testing"

	"gotest.tools/assert"
	"sigs.k8s.io/gateway-api/apis/v1alpha1"
)

func TestGatewayApiReconciler_HostsToCustomDomains(t *testing.T) {
	tests := map[string]struct {
		route     *v1alpha1.HTTPRoute
		want      string
		wantError error
		skip      bool
	}{
		"no hosts (match all)": {
			route: &v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Hostnames: []v1alpha1.Hostname{},
				},
			},
			want:      "",
			wantError: nil,
		},
		"single host": {
			route: &v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Hostnames: []v1alpha1.Hostname{
						"foo.bar.com",
					},
				},
			},
			want:      "foo.bar.com",
			wantError: nil,
		},
		"wildcard host": {
			route: &v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Hostnames: []v1alpha1.Hostname{
						"*.bar.com",
					},
				},
			},
			want:      "{?:[^.]+}.foo.com",
			wantError: nil,
			skip:      true,
		},
		"multiple exact hosts": {
			route: &v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Hostnames: []v1alpha1.Hostname{
						"foo.bar.com",
						"bar.foo.com",
					},
				},
			},
			want:      "{?:(foo.bar.com|bar.foo.com)}",
			wantError: nil,
		},
		"multiple wildcard hosts": {
			route: &v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Hostnames: []v1alpha1.Hostname{
						"*.bar.com",
						"*.foo.com",
					},
				},
			},
			want:      "{?:([^.]+.bar.com|[^.]+.foo.com)}", // TODO: need to actually check this - maybe im making it up
			wantError: nil,
			skip:      true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(innerTest *testing.T) {
			if tc.skip {
				innerTest.Skip()
			}
			c := GatewayApiReconciler{}
			customDomain, err := c.HostsToCustomDomains(tc.route.Spec.Hostnames)
			assert.Equal(innerTest, tc.want, customDomain)
			assert.Equal(innerTest, tc.wantError, err)
		})
	}
}
