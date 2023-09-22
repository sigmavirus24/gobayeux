package gobayeux

import "testing"

func TestHandshakeRequestBuilder_AddSupportedConnectionType(t *testing.T) {
	testCases := []struct {
		name      string
		ct        string
		shouldErr bool
	}{
		{
			"valid long-polling",
			"long-polling",
			false,
		},
		{
			"valid callback-polling",
			"callback-polling",
			false,
		},
		{
			"valid iframe",
			"iframe",
			false,
		},
		{
			"invalid connection type",
			"invalid-polling",
			true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			b := NewHandshakeRequestBuilder()
			err := b.AddSupportedConnectionType(tc.ct)
			if err != nil && !tc.shouldErr {
				t.Errorf("expected connection type %s to be valid but got err %q", tc.ct, err)
			}
			if err == nil && tc.shouldErr {
				t.Error("expected an error but didn't get one")
			}
		})
	}
}

func TestHandshakeRequestBuilder_AddVersion(t *testing.T) {
	testCases := []struct {
		name      string
		version   string
		shouldErr bool
	}{
		{
			"valid version 1.0",
			"1.0",
			false,
		},
		{
			"valid version 1.0beta",
			"1.0beta",
			false,
		},
		{
			"valid version 10.0",
			"10.0",
			false,
		},
		{
			"invalid version .0",
			".0",
			true,
		},
		{
			"invalid version a.0",
			"a.0",
			true,
		},
		{
			"invalid version (empty)",
			"",
			true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			b := NewHandshakeRequestBuilder()
			err := b.AddVersion(tc.version)
			if err != nil && !tc.shouldErr {
				t.Errorf("expected version %s to be valid but got err %q", tc.version, err)
			}
			if err == nil && tc.shouldErr {
				t.Error("expected an error but didn't get one")
			}
		})
	}
}
