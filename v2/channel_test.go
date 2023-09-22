package gobayeux

import (
	"testing"
)

func TestType(t *testing.T) {
	tests := []struct {
		name  string
		input Channel
		want  ChannelType
	}{
		{
			name:  "valid meta channel",
			input: "/meta/connect",
			want:  MetaChannel,
		},
		{
			name:  "invalid meta channel",
			input: "meta/connect",
			want:  BroadcastChannel,
		},
		{
			name:  "valid service channel",
			input: "/service/chat",
			want:  ServiceChannel,
		},
		{
			name:  "broadcast channel",
			input: "/foo/bar",
			want:  BroadcastChannel,
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.Type()
			if tc.want != got {
				t.Errorf("unexpected channel type got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestHasWildcard(t *testing.T) {
	tests := []struct {
		name  string
		input Channel
		want  bool
	}{
		{
			name:  "no wildcard",
			input: "/meta/connect",
			want:  false,
		},
		{
			name:  "single wildcard",
			input: "/foo/*",
			want:  true,
		},
		{
			name:  "double wildcard",
			input: "/foo/**",
			want:  true,
		},
		{
			name:  "invalid wildcard",
			input: "/foo/**/biz",
			want:  false,
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.HasWildcard()
			if tc.want != got {
				t.Errorf("unexpected result checking for wildcard got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		input Channel
		want  bool
	}{
		{
			name:  "valid channel without wildcards",
			input: "/foo",
			want:  true,
		},
		{
			name:  "valid channel with single wildcard",
			input: "/foo/*",
			want:  true,
		},
		{
			name:  "valid channel with double wildcard",
			input: "/foo/**",
			want:  true,
		},
		{
			name:  "invalid channel with wildcard",
			input: "/foo/*/bar",
			want:  false,
		},
		{
			name:  "invalid channel",
			input: "foo/bar",
			want:  false,
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.input.IsValid(); tc.want != got {
				t.Errorf("expected Channel(\"%s\").IsValid() == %v, got %v", string(tc.input), tc.want, got)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern Channel
		input   Channel
		want    bool
	}{
		{
			name:    "matching channels without wildcards",
			pattern: "/meta/connect",
			input:   "/meta/connect",
			want:    true,
		},
		{
			name:    "non-matching channels without wildcards",
			pattern: "/meta/connect",
			input:   "/foo/bar",
			want:    false,
		},
		{
			name:    "matching channels with single wildcard",
			pattern: "/foo/*",
			input:   "/foo/bar",
			want:    true,
		},
		{
			name:    "channel with too few wildcards",
			pattern: "/foo/*",
			input:   "/foo/bar/baz",
			want:    false,
		},
		{
			name:    "matching channel with wildcards",
			pattern: "/foo/**",
			input:   "/foo/bar",
			want:    true,
		},
		{
			name:    "matching a longer channel with wildcards",
			pattern: "/foo/**",
			input:   "/foo/bar/baz",
			want:    true,
		},
		{
			name:    "matching an invalid channel with wildcards",
			pattern: "*",
			input:   "/foo",
			want:    false,
		},
		{
			name:    "matching against a wildcard with different prefix",
			pattern: "/foo/*",
			input:   "/bar/baz",
			want:    false,
		},
		{
			name:    "invalid wildcard pattern",
			pattern: "/foo/***",
			input:   "/foo/bar",
			want:    false,
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pattern.Match(tc.input)
			if tc.want != got {
				t.Errorf("expected pattern match got %v, want %v", got, tc.want)
			}
		})
	}
}
