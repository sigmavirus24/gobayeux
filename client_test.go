package gobayeux

import "testing"

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name          string
		serverAddress string
		shouldErr     bool
	}{
		{"valid url for server address", "https://example.com", false},
		{"invalid url for server address", "http://192.168.0.%31/", true},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(tc.serverAddress)
			if err != nil && !tc.shouldErr {
				t.Errorf("expected NewClient() to not return an err but it did, %q", err)
			} else if tc.shouldErr && err == nil {
				t.Error("expected NewClient() to err but it didn't")
			}
		})
	}
}
