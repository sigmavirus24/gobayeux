package gobayeux

import (
	"errors"
	"fmt"
)

var (
	// ErrClientNotConnected is returned when the client is not connected
	ErrClientNotConnected = errors.New("client not connected to server")

	// ErrTooManyMessages is returned when there is more than one handshake message
	ErrTooManyMessages = errors.New("more messages than expected in handshake response")

	// ErrBadChannel is returned when the handshake response is on the wrong channel
	ErrBadChannel = errors.New("handshake responses must come back via the /meta/handshake channel")

	// ErrFailedToConnect is a general connection error
	ErrFailedToConnect = errors.New("connect request was not successful")

	// ErrNoSupportedConnectionTypes is returned when the client and server
	// aren't able to agree on a connection type
	ErrNoSupportedConnectionTypes = errors.New("no supported connection types provided")

	// ErrNoVersion is returned when a version is not provided
	ErrNoVersion = errors.New("no version specified")

	// ErrMissingClientID is returned when the client id has not been set
	ErrMissingClientID = errors.New("missing clientID value")

	// ErrMissingConnectionType is returned when the connection type is unset
	ErrMissingConnectionType = errors.New("missing connectionType value")
)

// ErrConnectionFailed is returned whenever Connect is called and it fails
type ErrConnectionFailed struct {
	err error
}

func (e ErrConnectionFailed) Error() string {
	return fmt.Sprintf("connection failed (%s)", e.err)
}

func (e ErrConnectionFailed) Unwrap() error {
	return e.err
}

// ErrHandshakeFailed is returned whenever the handshake fails
type ErrHandshakeFailed struct {
	err error
}

func (e ErrHandshakeFailed) Error() string {
	return e.err.Error()
}

func (e ErrHandshakeFailed) Unwrap() error {
	return e.err
}

func newHandshakeError(msg string) *ErrHandshakeFailed {
	return &ErrHandshakeFailed{
		fmt.Errorf("handshake was not successful: %s", msg),
	}
}

// ErrSubscriptionFailed is returned for any errors on Subscribe
type ErrSubscriptionFailed struct {
	Channels []Channel
	err      error
}

func (e ErrSubscriptionFailed) Error() string {
	return fmt.Sprintf("subscription failed (%s)", e.err)
}

func (e ErrSubscriptionFailed) Unwrap() error {
	return e.err
}

// ErrUnsubscribeFailed is returned for any errors on Unsubscribe
type ErrUnsubscribeFailed struct {
	Channels []Channel
	err      error
}

func (e ErrUnsubscribeFailed) Error() string {
	return fmt.Sprintf("subscription failed (%s)", e.err)
}

func (e ErrUnsubscribeFailed) Unwrap() error {
	return e.err
}

// ErrActionFailed is a general purpose error returned by the BayeuxClient
type ErrActionFailed struct {
	action string
	err    string
}

func (e ErrActionFailed) Error() string {
	return fmt.Sprintf("unable to %s channels: %s", e.action, e.err)
}

func newSubscribeError(msg string) *ErrActionFailed {
	return &ErrActionFailed{"subscribe to", msg}
}

func newUnsubscribeError(msg string) *ErrActionFailed {
	return &ErrActionFailed{"unsubscribe from", msg}
}

// ErrDisconnectFailed is returned when the call to Disconnect fails
type ErrDisconnectFailed struct {
	err error
}

func (e ErrDisconnectFailed) Error() string {
	msg := "unable to disconnect from Bayeux server"

	if e.err == nil {
		return msg
	}

	return fmt.Sprintf("%s (%s)", msg, e.err)
}

func (e ErrDisconnectFailed) Unwrap() error {
	return e.err
}

// ErrAlreadyRegistered signifies that the given MessageExtender is already
// registered with the client
type ErrAlreadyRegistered struct {
	MessageExtender
}

func (e ErrAlreadyRegistered) Error() string {
	return fmt.Sprintf("extension already registered: %s", e.MessageExtender)
}

// ErrBadResponse is returned when we get an unexpected HTTP response from the server
type ErrBadResponse struct {
	StatusCode int
	Status     string
}

func (e ErrBadResponse) Error() string {
	return fmt.Sprintf(
		"expected 200 response from bayeux server, got %d with status '%s'",
		e.StatusCode,
		e.Status,
	)
}

// ErrBadConnectionType is returned when we don't know how to handle the
// requested connection type
type ErrBadConnectionType struct {
	ConnectionType string
}

func (e ErrBadConnectionType) Error() string {
	return fmt.Sprintf("%q is not a valid connection type", e.ConnectionType)
}

// ErrBadConnectionVersion is returned when we can't support the requested
// version number
type ErrBadConnectionVersion struct {
	Version string
}

func (e ErrBadConnectionVersion) Error() string {
	return fmt.Sprintf("version %q is invalid for Bayeux protocol", e.Version)
}

// ErrInvalidChannel is the result of a failure to validate a channel name
type ErrInvalidChannel struct {
	Channel
}

func (e ErrInvalidChannel) Error() string {
	return fmt.Sprintf("channel %q appears to not be a valid channel", e.Channel)
}

// ErrEmptySlice is returned when an empty slice is unexpected
type ErrEmptySlice string

func (e ErrEmptySlice) Error() string {
	return fmt.Sprintf("no %s provided", string(e))
}

// ErrMessageUnparsable is returned when we fail to parse a message
type ErrMessageUnparsable string

func (e ErrMessageUnparsable) Error() string {
	return fmt.Sprintf("error message not parseable: %s", string(e))
}

// ErrBadState is returned when the state machine transition is not valid
type ErrBadState struct {
	CurrenctState int32
	FromState     int32
	ToState       int32
	msg           string
}

func (e ErrBadState) Error() string {
	return fmt.Sprintf("%s, (current: %s, from: %s, to: %s)", e.msg, stateName(e.CurrenctState), stateName(e.FromState), stateName(e.ToState))
}

// ErrBadHandshake is returned when trying to handshake but not unconnected
type ErrBadHandshake struct {
	*ErrBadState
}

func newBadHanshake(current, from, to int32) *ErrBadHandshake {
	return &ErrBadHandshake{
		&ErrBadState{
			msg:           "attempting to handshake but not in unconnected state",
			CurrenctState: current,
			FromState:     from,
			ToState:       to,
		},
	}
}

// ErrBadConnect is returned when trying to connected but not connecting
type ErrBadConnect struct {
	*ErrBadState
}

func newBadConnect(current, from, to int32) *ErrBadConnect {
	return &ErrBadConnect{
		&ErrBadState{
			msg:           "invalid state for successful connect response event",
			CurrenctState: current,
			FromState:     from,
			ToState:       to,
		},
	}
}

// ErrUnknownEventType is returned when the next state is unknown
type ErrUnknownEventType string

func (e ErrUnknownEventType) Error() string {
	return fmt.Sprintf("unknown event type (%q)", string(e))
}
