package salesforce

import (
	"net/http"
	"testing"
)

func TestStaticTokenAuthenticator(t *testing.T) {
	testCases := []struct {
		name              string
		url               string
		token             string
		expectedCallCount int
		shouldErr         bool
	}{
		{"Empty Token", "https://login.salesforce.com", "", 0, true},
		{"Non-empty Token", "https://login.salesforce.com", "token", 1, false},
		{"Request to something other than Salesforce", "https://github.com", "token", 0, false},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(testCase.name, func(t *testing.T) {
			trt := &TestRoundTripper{ExpectedToken: tc.token}
			sta := &StaticTokenAuthenticator{
				Token:     tc.token,
				Transport: trt,
			}
			req, _ := http.NewRequest("GET", tc.url, nil)
			_, err := sta.RoundTrip(req)
			if tc.shouldErr {
				if err == nil {
					t.Fatal("expected an error but received none")
				}
			}
			if err != nil && !tc.shouldErr {
				t.Fatalf("didn't expect an error but received one: %q", err)
			}
			if want, got := tc.expectedCallCount, trt.CallCount; want != got {
				t.Fatalf("expected to have called underlying transport with auth %d times but called it %d times", want, got)
			}
		})
	}
}

type TestRoundTripper struct {
	CallCount     int
	ExpectedToken string
}

// RoundTrip immplements the RoundTripper interface
func (t *TestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Header.Get("Authorization") == "Bearer "+t.ExpectedToken {
		t.CallCount++
	}
	return &http.Response{}, nil
}
