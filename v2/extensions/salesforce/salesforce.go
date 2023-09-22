// Package salesforce provides a simple way of authenticating with
// Salesforce.com Bayeux-powered services.
//
// An example usage looks like:
//
//	client := gobayeux.NewClient(serverAddress, gobayeux.WithHTTPTransport(salesforce.StaticTokenAuthenticator{myToken, http.DefaultTransport}))
package salesforce

import (
	"errors"
	"net/http"
	"strings"
)

// StaticTokenAuthenticator adds your Salesforce Access Token to your
// requests
type StaticTokenAuthenticator struct {
	// Token is the string obtained either from the Salesforce CX CLI (for
	// example). You can also retrieve this by using the curl command on
	// https://developer.salesforce.com/docs/atlas.en-us.api_iot.meta/api_iot/qs_auth_access_token.htm
	Token string
	// Transport is any http transport that satisfies the http.RoundTripper
	// interface
	Transport http.RoundTripper
	cookies   []*http.Cookie
}

// RoundTrip implements the RoundTripper interface
func (t *StaticTokenAuthenticator) RoundTrip(request *http.Request) (*http.Response, error) {
	if !strings.HasSuffix(request.URL.Hostname(), "salesforce.com") {
		return t.Transport.RoundTrip(request)
	}
	if t.Token == "" {
		return nil, errors.New("no Token provided to authenticator transport")
	}

	newRequest := deepCopyRequestWitHeaders(request)
	newRequest.Header.Set("Authorization", "Bearer "+t.Token)
	for _, cookie := range t.cookies {
		newRequest.AddCookie(cookie)
	}

	resp, err := t.Transport.RoundTrip(newRequest)
	if err != nil {
		return resp, err
	}
	t.cookies = resp.Cookies()
	return resp, nil
}

func deepCopyRequestWitHeaders(request *http.Request) *http.Request {
	newRequest := new(http.Request)
	*newRequest = *request

	newRequest.Header = make(http.Header, len(request.Header))
	for header, values := range request.Header {
		newRequest.Header[header] = append([]string(nil), values...)
	}
	return newRequest
}
