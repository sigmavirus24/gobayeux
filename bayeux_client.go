package gobayeux

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

// BayeuxClient is a way of acting as a client with a given Bayeux server
type BayeuxClient struct {
	stateMachine  *ConnectionStateMachine
	client        *http.Client
	serverAddress *url.URL
	state         *clientState
	exts          []MessageExtender
}

// NewBayeuxClient initializes a BayeuxClient for the user
func NewBayeuxClient(transport *http.Transport, serverAddress string) (*BayeuxClient, error) {
	if transport == nil {
		transport = &http.Transport{
			Dial:                  (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).Dial,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
		}
	}

	parsedAddress, err := url.Parse(serverAddress)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	return &BayeuxClient{
		stateMachine:  NewConnectionStateMachine(),
		client:        &http.Client{Transport: transport, Jar: jar},
		serverAddress: parsedAddress,
		state:         &clientState{},
		//subscriptionsChannel: make(chan subscriptionRequest, 10),
	}, nil
}

// Handshake sends the handshake request to the Bayeux Server
func (b *BayeuxClient) Handshake(ctx context.Context) ([]Message, error) {
	if err := b.stateMachine.ProcessEvent(handshakeSent); err != nil {
		return nil, err
	}
	builder := NewHandshakeRequestBuilder()
	if err := builder.AddVersion("1.0"); err != nil {
		return nil, err
	}
	if err := builder.AddSupportedConnectionType("long-polling"); err != nil {
		return nil, err
	}
	ms, err := builder.Build()
	if err != nil {
		return nil, err
	}
	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, err
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return response, err
	}
	if len(response) > 1 {
		return response, errors.New("more messages than expected in handshake response")
	}

	message := response[0]
	if !message.Successful {
		return response, fmt.Errorf("handshake was not successful: %s", message.Error)
	}
	if message.Channel != MetaHandshake {
		return response, errors.New("handshake responses must come back via the /meta/handshake channel")
	}
	b.state.SetClientID(message.ClientID)
	_ = b.stateMachine.ProcessEvent(successfullyConnected)
	return response, nil
}

// Connect sends the connect request to the Bayeux Server. The specification
// says that clients MUST maintain only one outstanding connect request. See
// https://docs.cometd.org/current/reference/#_bayeux_meta_connect
func (b *BayeuxClient) Connect(ctx context.Context) ([]Message, error) {
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		return nil, errors.New("client not connected to server")
	}
	builder := NewConnectRequestBuilder()
	builder.AddClientID(clientID)
	_ = builder.AddConnectionType(ConnectionTypeLongPolling)
	ms, err := builder.Build()
	if err != nil {
		return nil, err
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, err
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return response, err
	}

	if !response[0].Successful {
		return response, errors.New("connect request was not successful")
	}

	return response, nil
}

// Subscribe issues a MetaSubscribe request to the server to subscribe to the
// channels in the subscriptions slice
func (b *BayeuxClient) Subscribe(ctx context.Context, subscriptions []Channel) ([]Message, error) {
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		return nil, errors.New("client not connected to server")
	}

	builder := NewSubscribeRequestBuilder()
	builder.AddClientID(clientID)
	for _, s := range subscriptions {
		if err := builder.AddSubscription(s); err != nil {
			return nil, err
		}
	}

	ms, err := builder.Build()
	if err != nil {
		return nil, err
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, err
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return response, err
	}

	message := response[0]
	if !message.Successful {
		return response, fmt.Errorf("unable to subscribe to channels: %s", message.Error)
	}
	return response, nil
}

// Unsubscribe issues a MetaUnsubscribe request to the server to subscribe to the
// channels in the subscriptions slice
func (b *BayeuxClient) Unsubscribe(ctx context.Context, subscriptions []Channel) ([]Message, error) {
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		return nil, errors.New("client not connected to server")
	}

	builder := NewUnsubscribeRequestBuilder()
	builder.AddClientID(clientID)
	for _, s := range subscriptions {
		if err := builder.AddSubscription(s); err != nil {
			return nil, err
		}
	}

	ms, err := builder.Build()
	if err != nil {
		return nil, err
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, err
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return response, err
	}

	message := response[0]
	if !message.Successful {
		return response, fmt.Errorf("unable to unsubscribe from channels: %s", message.Error)
	}
	return response, nil
}

// Disconnect sends a /meta/disconnect request to the Bayeux server to
// terminate the session
func (b *BayeuxClient) Disconnect(ctx context.Context) ([]Message, error) {
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		return nil, errors.New("client isn't connected")
	}

	builder := NewDisconnectRequestBuilder()
	builder.AddClientID(clientID)
	ms, err := builder.Build()
	if err != nil {
		return nil, err
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, err
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return response, err
	}

	message := response[0]
	if !message.Successful {
		return response, errors.New("unable to disconnect from Bayeux server")
	}
	return response, nil
}

func (b *BayeuxClient) request(ctx context.Context, ms []Message) (*http.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(ms); err != nil {
		return nil, err
	}

	for _, ext := range b.exts {
		for _, m := range ms {
			ext.Outgoing(&m)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.serverAddress.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for _, cookie := range b.client.Jar.Cookies(b.serverAddress) {
		req.AddCookie(cookie)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}

	b.client.Jar.SetCookies(b.serverAddress, resp.Cookies())
	return resp, nil
}

func (b *BayeuxClient) parseResponse(response *http.Response) ([]Message, error) {
	messages := make([]Message, 0)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("expected 200 response from bayeux server, got %d with status '%s'", response.StatusCode, response.Status)
	}

	if err := json.NewDecoder(response.Body).Decode(&messages); err != nil {
		return nil, err
	}
	for _, ext := range b.exts {
		for _, m := range messages {
			ext.Incoming(&m)
		}
	}
	return messages, nil
}

type clientState struct {
	clientID string
	lock     sync.RWMutex
}

func (cs *clientState) GetClientID() string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()
	return cs.clientID
}

func (cs *clientState) SetClientID(clientID string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	cs.clientID = clientID
}
