package gobayeux

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	timestampFmt = "2006-01-02T15:04:05.00"
)

// Message represents a message received by a Bayeux client
//
// See also: https://docs.cometd.org/current/reference/#_bayeux_message_fields
type Message struct {
	// Advice provides a way for servers to inform clients of their preferred
	// mode of client operation.
	//
	// See also: https://docs.cometd.org/current/reference/#_bayeux_advice
	Advice *Advice `json:"advice,omitempty"`
	// ID represents the identifier of the specific message
	//
	// See also: https://docs.cometd.org/current/reference/#_bayeux_id
	ID string `json:"id,omitempty"`
	// Channel is the Channel on which the message was sent
	//
	// See also: https://docs.cometd.org/current/reference/#_channel
	Channel Channel `json:"channel"`
	// ClientID identifies a particular session via a session id token
	//
	// See also: https://docs.cometd.org/current/reference/#_bayeux_clientid
	ClientID string `json:"clientId,omitempty"`
	// Data contains an event information and optionally could contain
	// a binary representation of the data.
	// There is the MessageData type can be used for binary data
	//
	// See also:
	// https://docs.cometd.org/current/reference/#_data
	// https://docs.cometd.org/current/reference/#_concepts_binary_data
	Data json.RawMessage `json:"data,omitempty"`
	// Version indicates the protocol version expected by the client/server.
	// This MUST be included in messages to/from the `/meta/handshake`
	// channel.
	//
	// See also: https://docs.cometd.org/current/reference/#_version_2
	Version string `json:"version,omitempty"`
	// MinimumVersion indicates the oldest protocol version that can be handled
	// by the client/server. This MAY be included.
	//
	// See also: https://docs.cometd.org/current/reference/#_minimumversion
	MinimumVersion string `json:"minimumVersion,omitempty"`
	// SupportedConnectionTypes is included in messages to/from the
	// `/meta/handshake` channel to allow clients and servers to reveal the
	// transports that are supported. This is an array of strings.
	//
	// See also: https://docs.cometd.org/current/reference/#_bayeux_supported_connections
	SupportedConnectionTypes []string `json:"supportedConnectionTypes,omitempty"`
	// ConnectionType specifies the type of transport the client requires for
	// communication. This MUST be included in `/meta/connect` request
	// messages.
	//
	// See also:
	// https://docs.cometd.org/current/reference/#_connectiontype
	// https://docs.cometd.org/current/reference/#_bayeux_supported_connections
	ConnectionType string `json:"connectionType,omitempty"`
	// Timestamp is an optional field in all Bayeux messages. If present, it
	// SHOULD be specified in the following ISO 8601 profile:
	// `YYYY-MM-DDThh:mm:ss.ss`
	//
	// See also: https://docs.cometd.org/current/reference/#_timestamp
	Timestamp string `json:"timestamp,omitempty"`
	// Successful is a boolean field used to indicate success or failure and
	// MUST be included in responses to `/meta/handshake`, `/meta/connect`,
	// `/meta/subscribe`, `/meta/unsubscribe`, `/meta/disconnect`, and publish
	// channels.
	//
	// See also: https://docs.cometd.org/current/reference/#_successful
	Successful bool `json:"successful,omitempty"`
	// AuthSuccessful is not a common field but MAY be included on a handshake
	// response.
	AuthSuccessful bool `json:"authSuccessful,omitempty"`
	// Subscription specifies the channels the client wishes to subscribe to
	// or unsubscribe from and MUST be included in requests and responses
	// to/from the `/meta/subscribe` or `/meta/unsubscribe` channels.
	//
	// See also: https://docs.cometd.org/current/reference/#_subscription
	Subscription Channel `json:"subscription,omitempty"`
	// Error field is optional in any response and MAY indicate the type of
	// error that occurred when a request returns with a false successful
	// message.
	//
	// See also: https://docs.cometd.org/current/reference/#_error
	Error string `json:"error,omitempty"`
	// Ext is an optional field that SHOULD be JSON encoded. The contents can
	// be arbitrary values that allow extension sto be negotiated and
	// implemented between server and client implementations.
	//
	// See also: https://docs.cometd.org/current/reference/#_bayeux_ext
	Ext map[string]interface{} `json:"ext,omitempty"`
}

// TimestampAsTime returns the Timestamp in a message as a time.Time struct
func (m *Message) TimestampAsTime() (time.Time, error) {
	return time.Parse(timestampFmt, m.Timestamp)
}

// ParseError returns a struct representing the error message and parsed as
// defined in the specification.
//
// See also: https://docs.cometd.org/current/reference/#_error
func (m *Message) ParseError() (MessageError, error) {
	// TODO(sigmavirus24) actually parse the error
	pieces := strings.SplitN(m.Error, ":", 3)
	if len(pieces) != 3 {
		return MessageError{}, fmt.Errorf("error message not parseable: %s", m.Error)
	}
	errorCode, err := strconv.Atoi(pieces[0])
	if err != nil {
		return MessageError{}, err
	}
	return MessageError{
		errorCode,
		strings.Split(pieces[1], ","),
		pieces[2],
	}, nil
}

// GetExt retrieves the Ext field map. If passed `true` it will instantiate it
// if the map is not instantiated, otherwise it will just return the value of
// Ext.
func (m *Message) GetExt(create bool) map[string]interface{} {
	if m.Ext == nil && create {
		m.Ext = make(map[string]interface{})
	}
	return m.Ext
}

// Advice represents the field from the server which is used to inform clients
// of their preferred mode of client operation.
//
// See also: https://docs.cometd.org/current/reference/#_bayeux_advice
type Advice struct {
	// Reconnect indicates how the client should act in the case of a failure
	// to connect.
	//
	// See also: https://docs.cometd.org/current/reference/#_reconnect_advice_field
	Reconnect string `json:"reconnect,omitempty"`
	// Timeout represents the period of time, in milliseconds, for the server
	// to delay requests to the `/meta/connect` channel.
	//
	// See also: https://docs.cometd.org/current/reference/#_timeout_advice_field
	Timeout int `json:"timeout,omitempty"`
	// Interval represents the minimum period of time, in milliseconds, for the
	// client to delay subsequent requests to the /meta/connect channel.
	//
	// See also: https://docs.cometd.org/current/reference/#_interval_advice_field
	Interval int `json:"interval,omitempty"`
	// MultipleClients indicates that the server has detected multiple Bayeux
	// client instances running within the same web client
	//
	// See also: https://docs.cometd.org/current/reference/#_bayeux_multiple_clients_advice
	MultipleClients bool `json:"multiple-clients,omitempty"`
	// Hosts is an array of strings which if present indicates a list of host
	// names or IP addresses that MAY be used as alternate servers. If a
	// re-handshake advice is received by a client and the current server is
	// not in the supplied hosts list, then the client SHOULD try the hosts in
	// order.
	//
	// See also: https://docs.cometd.org/current/reference/#_hosts_advice_field
	Hosts []string `json:"hosts,omitempty"`
}

// MustNotRetryOrHandshake indicates whether neither a handshake or retry is
// allowed
func (a Advice) MustNotRetryOrHandshake() bool {
	return a.Reconnect == "none"
}

// ShouldRetry indicates whether a retry should occur
func (a Advice) ShouldRetry() bool {
	return a.Reconnect == "retry"
}

// ShouldHandshake indicates whether the advice is that a handshake should
// occur
func (a Advice) ShouldHandshake() bool {
	return a.Reconnect == "handshake"
}

// TimeoutAsDuration returns the Timeout field as a time.Duration for
// scheduling
func (a Advice) TimeoutAsDuration() time.Duration {
	return time.Duration(a.Timeout) * time.Millisecond
}

// IntervalAsDuration returns the Timeout field as a time.Duration for
// scheduling
func (a Advice) IntervalAsDuration() time.Duration {
	return time.Duration(a.Interval) * time.Millisecond
}

// MessageError represents a parsed Error field of a Message
//
// See also: https://docs.cometd.org/current/reference/#_error
type MessageError struct {
	ErrorCode    int
	ErrorArgs    []string
	ErrorMessage string
}

const (
	// ConnectionTypeLongPolling is a constant for the long-polling string
	ConnectionTypeLongPolling string = "long-polling"
	// ConnectionTypeCallbackPolling is a constant for the callback-polling string
	ConnectionTypeCallbackPolling = "callback-polling"
	// ConnectionTypeIFrame is a constant for the iframe string
	ConnectionTypeIFrame = "iframe"
)
