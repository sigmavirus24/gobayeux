package gobayeux

import (
	"encoding/json"
	"fmt"
)

func ExampleHandshakeRequestBuilder() {
	b := NewHandshakeRequestBuilder()
	if err := b.AddSupportedConnectionType(ConnectionTypeLongPolling); err != nil {
		return
	}
	if err := b.AddVersion("1.0"); err != nil {
		return
	}
	m, err := b.Build()
	if err != nil {
		return
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return
	}
	fmt.Println(string(jsonBytes))
	// Output:
	// [{"channel":"/meta/handshake","version":"1.0","supportedConnectionTypes":["long-polling"]}]
}

func ExampleConnectRequestBuilder() {
	b := NewConnectRequestBuilder()
	if err := b.AddConnectionType(ConnectionTypeLongPolling); err != nil {
		return
	}
	b.AddClientID("Un1q31d3nt1f13r")
	m, err := b.Build()
	if err != nil {
		return
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return
	}
	fmt.Println(string(jsonBytes))
	// Output:
	// [{"channel":"/meta/connect","clientId":"Un1q31d3nt1f13r","connectionType":"long-polling"}]
}

func ExampleSubscribeRequestBuilder() {
	b := NewSubscribeRequestBuilder()
	if err := b.AddSubscription("/foo/**"); err != nil {
		return
	}
	if err := b.AddSubscription("/foo/**"); err != nil { // NOTE We de-duplicate channels
		return
	}
	if err := b.AddSubscription("/bar/foo"); err != nil {
		return
	}
	b.AddClientID("Un1q31d3nt1f13r")
	m, err := b.Build()
	if err != nil {
		return
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return
	}
	fmt.Println(string(jsonBytes))
	// Output:
	// [{"channel":"/meta/subscribe","clientId":"Un1q31d3nt1f13r","subscription":["/foo/**","/bar/foo"]}]
}

func ExampleUnsubscribeRequestBuilder() {
	b := NewUnsubscribeRequestBuilder()
	if err := b.AddSubscription("/foo/**"); err != nil {
		return
	}
	if err := b.AddSubscription("/foo/**"); err != nil { // NOTE We de-duplicate channels
		return
	}
	if err := b.AddSubscription("/bar/foo"); err != nil {
		return
	}
	b.AddClientID("Un1q31d3nt1f13r")
	m, err := b.Build()
	if err != nil {
		return
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return
	}
	fmt.Println(string(jsonBytes))
	// Output:
	// [{"channel":"/meta/unsubscribe","clientId":"Un1q31d3nt1f13r","subscription":["/foo/**","/bar/foo"]}]
}
