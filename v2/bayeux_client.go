package gobayeux

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

// BayeuxClient is a way of acting as a client with a given Bayeux server
type BayeuxClient struct {
	stateMachine  *ConnectionStateMachine
	client        *http.Client
	serverAddress *url.URL
	state         *clientState
	exts          []MessageExtender
	logger        logrus.FieldLogger
}

// NewBayeuxClient initializes a BayeuxClient for the user
func NewBayeuxClient(client *http.Client, transport http.RoundTripper, serverAddress string, logger logrus.FieldLogger) (*BayeuxClient, error) {
	if client == nil {
		client = http.DefaultClient

		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			return nil, err
		}
		client.Jar = jar
	}
	if transport == nil {
		transport = http.DefaultTransport
	}
	client.Transport = transport

	parsedAddress, err := url.Parse(serverAddress)
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &BayeuxClient{
		stateMachine:  NewConnectionStateMachine(),
		client:        client,
		serverAddress: parsedAddress,
		state:         &clientState{},
		logger:        logger,
	}, nil
}

// Handshake sends the handshake request to the Bayeux Server
func (b *BayeuxClient) Handshake(ctx context.Context) ([]Message, error) {
	logger := b.logger.WithField("at", "handshake")
	start := time.Now()
	logger.Debug("starting")
	if err := b.stateMachine.ProcessEvent(handshakeSent); err != nil {
		logger.WithError(err).Debug("invalid action for current state")
		return nil, HandshakeFailedError{err}
	}
	builder := NewHandshakeRequestBuilder()
	if err := builder.AddVersion("1.0"); err != nil {
		return nil, HandshakeFailedError{err}
	}
	if err := builder.AddSupportedConnectionType("long-polling"); err != nil {
		return nil, HandshakeFailedError{err}
	}
	ms, err := builder.Build()
	if err != nil {
		return nil, HandshakeFailedError{err}
	}
	resp, err := b.request(ctx, ms)
	if err != nil {
		logger.WithError(err).Debug("error during request")
		return nil, HandshakeFailedError{err}
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		logger.WithError(err).Debug("error parsing response")
		return response, HandshakeFailedError{err}
	}
	if len(response) > 1 {
		return response, HandshakeFailedError{ErrTooManyMessages}
	}

	var message Message
	for _, m := range response {
		if m.Channel == MetaHandshake {
			message = m
		}
	}
	if message.Channel == emptyChannel {
		return response, HandshakeFailedError{ErrBadChannel}
	}
	if !message.Successful {
		return response, newHandshakeError(message.Error)
	}
	b.state.SetClientID(message.ClientID)
	_ = b.stateMachine.ProcessEvent(successfullyConnected)
	logger.WithField("duration", time.Since(start)).Debug("finishing")
	return response, nil
}

// Connect sends the connect request to the Bayeux Server. The specification
// says that clients MUST maintain only one outstanding connect request. See
// https://docs.cometd.org/current/reference/#_bayeux_meta_connect
func (b *BayeuxClient) Connect(ctx context.Context) ([]Message, error) {
	logger := b.logger.WithField("at", "connect")
	start := time.Now()
	logger.Debug("starting")
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		return nil, ErrClientNotConnected
	}
	builder := NewConnectRequestBuilder()
	builder.AddClientID(clientID)
	_ = builder.AddConnectionType(ConnectionTypeLongPolling)
	ms, err := builder.Build()
	if err != nil {
		return nil, ConnectionFailedError{err}
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		logger.WithError(err).Debug("error during request")
		return nil, ConnectionFailedError{err}
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		logger.WithError(err).Debug("error parsing response")
		return response, ConnectionFailedError{err}
	}

	for _, m := range response {
		if m.Channel == MetaConnect && !m.Successful {
			return response, ConnectionFailedError{ErrFailedToConnect}
		}
	}
	logger.WithField("duration", time.Since(start)).Debug("finishing")
	return response, nil
}

// Subscribe issues a MetaSubscribe request to the server to subscribe to the
// channels in the subscriptions slice
func (b *BayeuxClient) Subscribe(ctx context.Context, subscriptions []Channel) ([]Message, error) {
	logger := b.logger.WithField("at", "subscribe")
	start := time.Now()
	logger.Debug("starting")
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		logger.Debug("cannot subscribe because client is not connected")
		return nil, SubscriptionFailedError{subscriptions, ErrClientNotConnected}
	}

	builder := NewSubscribeRequestBuilder()
	builder.AddClientID(clientID)
	for _, s := range subscriptions {
		if err := builder.AddSubscription(s); err != nil {
			return nil, SubscriptionFailedError{subscriptions, err}
		}
	}

	ms, err := builder.Build()
	if err != nil {
		return nil, SubscriptionFailedError{subscriptions, err}
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, SubscriptionFailedError{subscriptions, err}
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return nil, SubscriptionFailedError{subscriptions, err}
	}

	for _, m := range response {
		if m.Channel == MetaSubscribe && !m.Successful {
			return nil, SubscriptionFailedError{
				Channels: subscriptions,
				Err:      newSubscribeError(m.Error),
			}
		}
	}
	logger.WithField("duration", time.Since(start)).Debug("finishing")
	return response, nil
}

// Unsubscribe issues a MetaUnsubscribe request to the server to subscribe to the
// channels in the subscriptions slice
func (b *BayeuxClient) Unsubscribe(ctx context.Context, subscriptions []Channel) ([]Message, error) {
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		return nil, UnsubscribeFailedError{subscriptions, ErrClientNotConnected}
	}

	builder := NewUnsubscribeRequestBuilder()
	builder.AddClientID(clientID)
	for _, s := range subscriptions {
		if err := builder.AddSubscription(s); err != nil {
			return nil, UnsubscribeFailedError{subscriptions, err}
		}
	}

	ms, err := builder.Build()
	if err != nil {
		return nil, UnsubscribeFailedError{subscriptions, err}
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, UnsubscribeFailedError{subscriptions, err}
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return response, UnsubscribeFailedError{subscriptions, err}
	}

	for _, m := range response {
		if m.Channel == MetaUnsubscribe && !m.Successful {
			return response, UnsubscribeFailedError{
				Channels: subscriptions,
				Err:      newUnsubscribeError(m.Error),
			}
		}
	}
	return response, nil
}

// Disconnect sends a /meta/disconnect request to the Bayeux server to
// terminate the session
func (b *BayeuxClient) Disconnect(ctx context.Context) ([]Message, error) {
	clientID := b.state.GetClientID()
	if !b.stateMachine.IsConnected() || clientID == "" {
		return nil, DisconnectFailedError{ErrClientNotConnected}
	}

	builder := NewDisconnectRequestBuilder()
	builder.AddClientID(clientID)
	ms, err := builder.Build()
	if err != nil {
		return nil, DisconnectFailedError{err}
	}

	resp, err := b.request(ctx, ms)
	if err != nil {
		return nil, DisconnectFailedError{err}
	}

	response, err := b.parseResponse(resp)
	if err != nil {
		return response, DisconnectFailedError{err}
	}

	for _, m := range response {
		if m.Channel == MetaDisconnect && !m.Successful {
			return response, DisconnectFailedError{nil}
		}
	}
	return response, nil
}

// UseExtension adds the provided MessageExtender to the list of known
// extensions
func (b *BayeuxClient) UseExtension(ext MessageExtender) error {
	for _, registered := range b.exts {
		if ext == registered {
			return AlreadyRegisteredError{ext}
		}
	}
	b.exts = append(b.exts, ext)
	return nil
}

func (b *BayeuxClient) request(ctx context.Context, ms []Message) (*http.Response, error) {
	for _, ext := range b.exts {
		for _, m := range ms {
			ext.Outgoing(&m)
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(ms); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.serverAddress.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return b.client.Do(req)
}

func (b *BayeuxClient) parseResponse(resp *http.Response) ([]Message, error) {
	messages := make([]Message, 0)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, BadResponseError{resp.StatusCode, resp.Status}
	}

	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
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
