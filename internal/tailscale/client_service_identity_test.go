package tailscale

import (
	"strings"
	"testing"

	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/views"
)

func TestHasTagBasedIdentity(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want bool
	}{
		{
			name: "empty",
			tags: nil,
			want: false,
		},
		{
			name: "user-authenticated-hostname-only",
			tags: []string{"autogroup:member"},
			want: false,
		},
		{
			name: "tag-based-identity",
			tags: []string{"tag:service-host"},
			want: true,
		},
		{
			name: "mixed-tags",
			tags: []string{"autogroup:member", "tag:prod"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasTagBasedIdentity(tt.tags); got != tt.want {
				t.Fatalf("unexpected tag identity result: got %t want %t", got, tt.want)
			}
		})
	}
}

func TestValidateServiceHostIdentity(t *testing.T) {
	mkStatus := func(tags []string) *ipnstate.Status {
		var viewTags *views.Slice[string]
		if tags != nil {
			v := views.SliceOf(tags)
			viewTags = &v
		}
		return &ipnstate.Status{
			Self: &ipnstate.PeerStatus{
				Tags: viewTags,
			},
		}
	}

	t.Run("fails without-tag-based-identity", func(t *testing.T) {
		err := validateServiceHostIdentity(mkStatus([]string{"autogroup:member"}), "svc:portal")
		if err == nil {
			t.Fatalf("expected validation error")
		}
		if got, want := err.Error(), "tag-based identity"; !strings.Contains(got, want) {
			t.Fatalf("expected error containing %q, got %q", want, got)
		}
	})

	t.Run("passes-with-tag-based-identity", func(t *testing.T) {
		if err := validateServiceHostIdentity(mkStatus([]string{"tag:service-host"}), "svc:portal"); err != nil {
			t.Fatalf("expected no validation error, got %v", err)
		}
	})
}
