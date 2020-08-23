package gobayeux

import (
	"errors"
	"sync/atomic"
)

// StateRepresentation represents the current state of a connection as a
// string
type StateRepresentation string

const (
	unconnected int32 = iota
	connecting
	connected
)

const (
	unconnectedRepr StateRepresentation = "UNCONNECTED"
	connectingRepr  StateRepresentation = "CONNECTING"
	connectedRepr   StateRepresentation = "CONNECTED"
)

// Event represents and event that can change the state of a state machine
type Event string

const (
	handshakeSent         Event = "handshake request sent"
	timeout               Event = "Timeout"
	successfullyConnected Event = "Successful connect response"
	disconnectSent        Event = "Disconnect request sent"
)

// ConnectionStateMachine handles managing the connection's state
//
// See also: https://docs.cometd.org/current/reference/#_client_state_table
type ConnectionStateMachine struct {
	currentState *int32
}

// NewConnectionStateMachine creates a new ConnectionStateMachine to manage a
// connection's state
func NewConnectionStateMachine() *ConnectionStateMachine {
	defaultState := unconnected
	return &ConnectionStateMachine{&defaultState}
}

// IsConnected reflects whether the connection is connected to the Bayeux
// server
func (csm *ConnectionStateMachine) IsConnected() bool {
	return atomic.CompareAndSwapInt32(csm.currentState, connected, connected)
}

// CurrentState provides a string representation of the current state of the
// state machine
func (csm *ConnectionStateMachine) CurrentState() StateRepresentation {
	currentState := atomic.LoadInt32(csm.currentState)
	switch currentState {
	case connecting:
		return connectingRepr
	case connected:
		return connectedRepr
	default:
		return unconnectedRepr
	}
}

// ProcessEvent handles an event
func (csm *ConnectionStateMachine) ProcessEvent(e Event) error {
	switch e {
	case handshakeSent:
		if !atomic.CompareAndSwapInt32(csm.currentState, unconnected, connecting) {
			return errors.New("attempting to handshake but not in unconnected state")
		}
	case timeout:
		atomic.SwapInt32(csm.currentState, unconnected)
	case successfullyConnected:
		if !atomic.CompareAndSwapInt32(csm.currentState, connecting, connected) {
			return errors.New("invalid state for successful connect response event")
		}
	case disconnectSent:
		currentState := atomic.LoadInt32(csm.currentState)
		if currentState == connected || currentState == connecting {
			atomic.StoreInt32(csm.currentState, unconnected)
		}
	default:
		return errors.New("unknown event type")
	}
	return nil
}
