package gobayeux

import (
	"fmt"
)

const (
	// ErrClientNotConnected is returned when the client is not connected
	ErrClientNotConnected = sentinel("client not connected to server")

	// ErrTooManyMessages is returned when there is more than one handshake message
	ErrTooManyMessages = sentinel("more messages than expected in handshake response")

	// ErrBadChannel is returned when the handshake response is on the wrong channel
	ErrBadChannel = sentinel("handshake responses must come back via the /meta/handshake channel")

	// ErrFailedToConnect is a general connection error
	ErrFailedToConnect = sentinel("connect request was not successful")

	// ErrNoSupportedConnectionTypes is returned when the client and server
	// aren't able to agree on a connection type
	ErrNoSupportedConnectionTypes = sentinel("no supported connection types provided")

	// ErrNoVersion is returned when a version is not provided
	ErrNoVersion = sentinel("no version specified")

	// ErrMissingClientID is returned when the client id has not been set
	ErrMissingClientID = sentinel("missing clientID value")

	// ErrMissingConnectionType is returned when the connection type is unset
	ErrMissingConnectionType = sentinel("missing connectionType value")
)

type sentinel string

func (s sentinel) Error() string {
	return string(s)
}

// ConnectionFailedError is returned whenever Connect is called and it fails
type ConnectionFailedError struct {
	Err error
}

func (e ConnectionFailedError) Error() string {
	return fmt.Sprintf("connection failed (%s)", e.Err)
}

func (e ConnectionFailedError) Unwrap() error {
	return e.Err
}

// HandshakeFailedError is returned whenever the handshake fails
type HandshakeFailedError struct {
	Err error
}

func (e HandshakeFailedError) Error() string {
	return e.Err.Error()
}

func (e HandshakeFailedError) Unwrap() error {
	return e.Err
}

func newHandshakeError(msg string) *HandshakeFailedError {
	return &HandshakeFailedError{
		fmt.Errorf("handshake was not successful: %s", msg),
	}
}

// SubscriptionFailedError is returned for any errors on Subscribe
type SubscriptionFailedError struct {
	Channels []Channel
	Err      error
}

func (e SubscriptionFailedError) Error() string {
	return fmt.Sprintf("subscription failed (%s)", e.Err)
}

func (e SubscriptionFailedError) Unwrap() error {
	return e.Err
}

// UnsubscribeFailedError is returned for any errors on Unsubscribe
type UnsubscribeFailedError struct {
	Channels []Channel
	Err      error
}

func (e UnsubscribeFailedError) Error() string {
	return fmt.Sprintf("subscription failed (%s)", e.Err)
}

func (e UnsubscribeFailedError) Unwrap() error {
	return e.Err
}

// ActionFailedError is a general purpose error returned by the BayeuxClient
type ActionFailedError struct {
	Action       string
	ErrorMessage string
}

func (e ActionFailedError) Error() string {
	return fmt.Sprintf("unable to %s channels: %s", e.Action, e.ErrorMessage)
}

func newSubscribeError(msg string) *ActionFailedError {
	return &ActionFailedError{"subscribe to", msg}
}

func newUnsubscribeError(msg string) *ActionFailedError {
	return &ActionFailedError{"unsubscribe from", msg}
}

// DisconnectFailedError is returned when the call to Disconnect fails
type DisconnectFailedError struct {
	Err error
}

func (e DisconnectFailedError) Error() string {
	msg := "unable to disconnect from Bayeux server"

	if e.Err == nil {
		return msg
	}

	return fmt.Sprintf("%s (%s)", msg, e.Err)
}

func (e DisconnectFailedError) Unwrap() error {
	return e.Err
}

// AlreadyRegisteredError signifies that the given MessageExtender is already
// registered with the client
type AlreadyRegisteredError struct {
	MessageExtender
}

func (e AlreadyRegisteredError) Error() string {
	return fmt.Sprintf("extension already registered: %s", e.MessageExtender)
}

// BadResponseError is returned when we get an unexpected HTTP response from the server
type BadResponseError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e BadResponseError) Error() string {
	return fmt.Sprintf(
		"expected 200 response from bayeux server, got %d with status '%s' and body '%s'",
		e.StatusCode,
		e.Status,
		e.Body,
	)
}

// BadConnectionTypeError is returned when we don't know how to handle the
// requested connection type
type BadConnectionTypeError struct {
	ConnectionType string
}

func (e BadConnectionTypeError) Error() string {
	return fmt.Sprintf("%q is not a valid connection type", e.ConnectionType)
}

// BadConnectionVersionError is returned when we can't support the requested
// version number
type BadConnectionVersionError struct {
	Version string
}

func (e BadConnectionVersionError) Error() string {
	return fmt.Sprintf("version %q is invalid for Bayeux protocol", e.Version)
}

// InvalidChannelError is the result of a failure to validate a channel name
type InvalidChannelError struct {
	Channel
}

func (e InvalidChannelError) Error() string {
	return fmt.Sprintf("channel %q appears to not be a valid channel", e.Channel)
}

// EmptySliceError is returned when an empty slice is unexpected
type EmptySliceError string

func (e EmptySliceError) Error() string {
	return fmt.Sprintf("no %s provided", string(e))
}

// ErrMessageUnparsable is returned when we fail to parse a message
type ErrMessageUnparsable string

func (e ErrMessageUnparsable) Error() string {
	return fmt.Sprintf("error message not parseable: %s", string(e))
}

// BadStateError is returned when the state machine transition is not valid
type BadStateError struct {
	CurrentState int32
	FromState    int32
	ToState      int32
	Message      string
}

func (e BadStateError) Error() string {
	return fmt.Sprintf("%s, (current: %s, from: %s, to: %s)", e.Message, stateName(e.CurrentState), stateName(e.FromState), stateName(e.ToState))
}

// BadHandshakeError is returned when trying to handshake but not unconnected
type BadHandshakeError struct {
	*BadStateError
}

func newBadHanshake(current, from, to int32) *BadHandshakeError {
	return &BadHandshakeError{
		&BadStateError{
			Message:      "attempting to handshake but not in unconnected state",
			CurrentState: current,
			FromState:    from,
			ToState:      to,
		},
	}
}

// BadConnectionError is returned when trying to connected but not connecting
type BadConnectionError struct {
	*BadStateError
}

func newBadConnection(current, from, to int32) *BadConnectionError {
	return &BadConnectionError{
		&BadStateError{
			Message:      "invalid state for successful connect response event",
			CurrentState: current,
			FromState:    from,
			ToState:      to,
		},
	}
}

// UnknownEventTypeError is returned when the next state is unknown
type UnknownEventTypeError struct {
	Event
}

func (e UnknownEventTypeError) Error() string {
	return fmt.Sprintf("unknown event type (%q)", e.Event)
}
