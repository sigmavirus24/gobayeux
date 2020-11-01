package gobayeux

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// HandshakeRequestBuilder provides a way to safely and confidently create
// handshake requests to /meta/handshake.
//
// See also: https://docs.cometd.org/current/reference/#_handshake_request
type HandshakeRequestBuilder struct {
	// Required fields
	version                  string
	supportedConnectionTypes []string
	// Optional fields
	minimumVersion string
}

// NewHandshakeRequestBuilder provides an easy way to build a Message that can
// be sent as a Handshake Request as documented in
// https://docs.cometd.org/current/reference/#_handshake_request
func NewHandshakeRequestBuilder() *HandshakeRequestBuilder {
	return &HandshakeRequestBuilder{
		supportedConnectionTypes: make([]string, 0),
	}
}

// AddSupportedConnectionType accepts a string and will add it to the list of
// supported connection types for the /meta/handshake request. It validates
// the connection type. You're encouraged to use one of the constants created
// for these different connection types.
// This will de-duplicate connection types and returns an error if an invalid
// connection type was provided.
func (b *HandshakeRequestBuilder) AddSupportedConnectionType(connectionType string) error {
	switch connectionType {
	case ConnectionTypeCallbackPolling, ConnectionTypeLongPolling, ConnectionTypeIFrame:
		for _, ct := range b.supportedConnectionTypes {
			if ct == connectionType {
				return nil
			}
		}
		b.supportedConnectionTypes = append(b.supportedConnectionTypes, connectionType)
	default:
		return fmt.Errorf("'%s' is not a valid connection type", connectionType)
	}
	return nil
}

// AddVersion accepts the version of the Bayeux protocol that the client
// supports.
func (b *HandshakeRequestBuilder) AddVersion(version string) error {
	if len(version) < 1 {
		return fmt.Errorf("version '%s' is invalid for Bayeux protocol", version)
	}
	pieces := strings.SplitN(version, ".", 2)
	if _, err := strconv.Atoi(pieces[0]); err != nil {
		return err
	}
	b.version = version
	return nil
}

// AddMinimumVersion adds the minimum supported version
func (b *HandshakeRequestBuilder) AddMinimumVersion(version string) error {
	if len(version) < 1 {
		return fmt.Errorf("version '%s' is invalid for Bayeux protocol", version)
	}
	pieces := strings.SplitN(version, ".", 2)
	if _, err := strconv.Atoi(pieces[0]); err != nil {
		return err
	}
	b.minimumVersion = version
	return nil
}

// Build generates the final Message to be sent as a Handshake Request
func (b *HandshakeRequestBuilder) Build() ([]Message, error) {
	if len(b.supportedConnectionTypes) < 1 {
		return nil, errors.New("no supported connection types provided")
	}
	if len(b.version) == 0 {
		return nil, errors.New("no version specified")
	}
	m := Message{
		Channel:                  MetaHandshake,
		Version:                  b.version,
		SupportedConnectionTypes: b.supportedConnectionTypes,
	}
	if len(b.minimumVersion) > 0 {
		m.MinimumVersion = b.minimumVersion
	}
	// TODO After we've added methods for id, ext, and minimumVersion, update
	// those values in the struct here as well
	return []Message{m}, nil
}

// ConnectRequestBuilder provides a way to safely build a Message that can be
// sent as a /meta/connect request as documented in
// https://docs.cometd.org/current/reference/#_connect_request
type ConnectRequestBuilder struct {
	clientID       string
	connectionType string
}

// NewConnectRequestBuilder initializes a ConnectRequestBuilder as an easy way
// to build a Message that can be sent as a /meta/connect request.
//
// See also: https://docs.cometd.org/current/reference/#_connect_request
func NewConnectRequestBuilder() *ConnectRequestBuilder {
	return &ConnectRequestBuilder{}
}

// AddClientID adds the previously provided clientId to the request
func (b *ConnectRequestBuilder) AddClientID(clientID string) {
	b.clientID = clientID
}

// AddConnectionType adds the connection type used by the client for the
// purposes of this connection to the request
func (b *ConnectRequestBuilder) AddConnectionType(connectionType string) error {
	switch connectionType {
	case ConnectionTypeCallbackPolling, ConnectionTypeLongPolling, ConnectionTypeIFrame:
		b.connectionType = connectionType
	default:
		return fmt.Errorf("'%s' is not a valid connection type", connectionType)
	}
	return nil
}

// TODO Add methods for id and ext

// Build generates the final Message to be sent as a Connect Request
func (b *ConnectRequestBuilder) Build() ([]Message, error) {
	if b.clientID == "" {
		return nil, errors.New("missing clientID value")
	}

	if b.connectionType == "" {
		return nil, errors.New("missing connectionType value")
	}

	m := Message{
		Channel:        MetaConnect,
		ClientID:       b.clientID,
		ConnectionType: b.connectionType,
	}
	// TODO After we've added methods for id and ext, update
	// those values in the struct here as well
	return []Message{m}, nil
}

// SubscribeRequestBuilder provides an easy way to build a /meta/subscribe
// request per the specification in
// https://docs.cometd.org/current/reference/#_subscribe_request
type SubscribeRequestBuilder struct {
	clientID     string
	subscription []Channel
}

// NewSubscribeRequestBuilder initializes a SubscribeRequestBuilder as an easy
// way to build a Message that can be sent as a /meta/subscribe request. See
// also https://docs.cometd.org/current/reference/#_subscribe_request
func NewSubscribeRequestBuilder() *SubscribeRequestBuilder {
	return &SubscribeRequestBuilder{subscription: make([]Channel, 0)}
}

// AddClientID adds the previously provided clientId to the request
func (b *SubscribeRequestBuilder) AddClientID(clientID string) {
	b.clientID = clientID
}

// AddSubscription adds a given channel to the list of subscriptions being
// sent in a /meta/subscribe request
func (b *SubscribeRequestBuilder) AddSubscription(c Channel) error {
	if !c.IsValid() {
		return fmt.Errorf("channel %s appears to not be a valid channel", c)
	}

	for _, s := range b.subscription {
		if s == c {
			return nil
		}
	}
	b.subscription = append(b.subscription, c)
	return nil
}

// Build generates the final Message to be sent as a Subscribe Request
func (b *SubscribeRequestBuilder) Build() ([]Message, error) {
	if b.clientID == "" {
		return nil, errors.New("missing clientID value")
	}

	if len(b.subscription) < 1 {
		return nil, errors.New("no subscriptions provided")
	}

	ms := make([]Message, len(b.subscription))

	for i := range b.subscription {
		ms[i] = Message{
			Channel:      MetaSubscribe,
			ClientID:     b.clientID,
			Subscription: b.subscription[i],
		}
	}

	// TODO Add the ext and id fields once we're able to handle them with the
	// builder
	return ms, nil
}

// UnsubscribeRequestBuilder provides an easy way to build a /meta/unsubscribe
// request per the specification in
// https://docs.cometd.org/current/reference/#_unsubscribe_request
type UnsubscribeRequestBuilder struct {
	clientID     string
	subscription []Channel
}

// NewUnsubscribeRequestBuilder initializes a SubscribeRequestBuilder as an easy
// way to build a Message that can be sent as a /meta/subscribe request. See
// also https://docs.cometd.org/current/reference/#_unsubscribe_request
func NewUnsubscribeRequestBuilder() *UnsubscribeRequestBuilder {
	return &UnsubscribeRequestBuilder{subscription: make([]Channel, 0)}
}

// AddClientID adds the previously provided clientId to the request
func (b *UnsubscribeRequestBuilder) AddClientID(clientID string) {
	b.clientID = clientID
}

// AddSubscription adds a given channel to the list of subscriptions being
// sent in a /meta/unsubscribe request
func (b *UnsubscribeRequestBuilder) AddSubscription(c Channel) error {
	if !c.IsValid() {
		return fmt.Errorf("channel %s appears to not be a valid channel", c)
	}

	for _, s := range b.subscription {
		if s == c {
			return nil
		}
	}
	b.subscription = append(b.subscription, c)
	return nil
}

// Build generates the final Message to be sent as a Unsubscribe Request
func (b *UnsubscribeRequestBuilder) Build() ([]Message, error) {
	if b.clientID == "" {
		return nil, errors.New("missing clientID value")
	}

	if len(b.subscription) < 1 {
		return nil, errors.New("no subscriptions provided")
	}

	ms := make([]Message, len(b.subscription))

	for i := range b.subscription {
		ms[i] = Message{
			Channel:      MetaUnsubscribe,
			ClientID:     b.clientID,
			Subscription: b.subscription[i],
		}
	}
	// TODO Add the ext and id fields once we're able to handle them with the
	// builder
	return ms, nil
}

// DisconnectRequestBuilder provides an easy way to build a /meta/disconnect
// request per the specification in
// https://docs.cometd.org/current/reference/#_bayeux_meta_disconnect
type DisconnectRequestBuilder struct {
	clientID string
}

// NewDisconnectRequestBuilder initializes a DisconnectRequestBuilder as an
// easy way to build a Message that can be sent as a /meta/disconnect request.
func NewDisconnectRequestBuilder() *DisconnectRequestBuilder {
	return &DisconnectRequestBuilder{}
}

// AddClientID adds the previously provided clientId to the request
func (b *DisconnectRequestBuilder) AddClientID(clientID string) {
	b.clientID = clientID
}

// Build generates the final Message to be sent as a Disconnect Request
func (b *DisconnectRequestBuilder) Build() ([]Message, error) {
	if b.clientID == "" {
		return nil, errors.New("missing clientID value")
	}

	return []Message{{Channel: MetaDisconnect, ClientID: b.clientID}}, nil
}
