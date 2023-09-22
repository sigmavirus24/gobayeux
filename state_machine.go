package gobayeux

import (
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

var stateNames = []string{"unconnected", "connecting", "connected"}

func stateName(state int32) string {
	s := int(state)
	if s < 0 || s >= len(stateNames) {
		return "unknown"
	}

	return stateNames[s]
}

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
			return newBadHanshake(atomic.LoadInt32(csm.currentState), unconnected, connecting)
		}
	case timeout:
		atomic.SwapInt32(csm.currentState, unconnected)
	case successfullyConnected:
		if !atomic.CompareAndSwapInt32(csm.currentState, connecting, connected) {
			return newBadConnection(atomic.LoadInt32(csm.currentState), connecting, connected)
		}
	case disconnectSent:
		currentState := atomic.LoadInt32(csm.currentState)
		if currentState == connected || currentState == connecting {
			atomic.StoreInt32(csm.currentState, unconnected)
		}
	default:
		return UnknownEventTypeError{e}
	}
	return nil
}
