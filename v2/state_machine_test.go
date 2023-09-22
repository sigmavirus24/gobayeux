package gobayeux

import "testing"

func TestNewConnectionStateMachineDefaults(t *testing.T) {
	csm := NewConnectionStateMachine()
	if csm.IsConnected() == true {
		t.Error("expected IsConnected() to be false, got true")
	}
	*csm.currentState = connected
	if csm.IsConnected() != true {
		t.Error("expected IsConnected() to be true, got false")
	}
}

func TestProcessEvent(t *testing.T) {
	testCases := []struct {
		name          string
		startingState int32
		event         Event
		shouldErr     bool
		endingState   int32
	}{
		{
			"unconnected state machine gets handshake request sent event",
			unconnected,
			handshakeSent,
			false,
			connecting,
		},
		{
			"connected state machine gets handshake request sent event",
			connected,
			handshakeSent,
			true,
			connected,
		},
		{
			"unconnected state machine gets successful connect response",
			unconnected,
			successfullyConnected,
			true,
			unconnected,
		},
		{
			"unconnected state machine gets unknown event",
			unconnected,
			"random",
			true,
			unconnected,
		},
		{
			"unconnected state machine gets timeout",
			unconnected,
			timeout,
			false,
			unconnected,
		},
		{
			"connecting state machine gets successfully connected response",
			connecting,
			successfullyConnected,
			false,
			connected,
		},
		{
			"connecting state machine gets timeout",
			connecting,
			timeout,
			false,
			unconnected,
		},
		{
			"connecting state machine gets disconnect request sent",
			connecting,
			disconnectSent,
			false,
			unconnected,
		},
		{
			"connecting state machine gets unknown event",
			connecting,
			"random",
			true,
			unconnected,
		},
		{
			"connected state machine gets timeout",
			connected,
			timeout,
			false,
			unconnected,
		},
		{
			"connected state machine gets disconnect request sent",
			connected,
			disconnectSent,
			false,
			unconnected,
		},
		{
			"connected state machine gets unknown event",
			connected,
			"random",
			true,
			unconnected,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			startingState := tc.startingState
			csm := &ConnectionStateMachine{&startingState}
			err := csm.ProcessEvent(tc.event)
			if tc.shouldErr && err == nil {
				t.Error("expected ProcessEvent to error but it didn't")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("didn't expect ProcessEvent to error but it did: %q", err)
			}
			if tc.shouldErr && err != nil {
				return
			}
			if tc.endingState != *csm.currentState {
				t.Errorf("unexpected ending state: want %d, got %d", tc.endingState, *csm.currentState)
			}
		})
	}
}

func TestCurrentState(t *testing.T) {
	testCases := []struct {
		name  string
		state int32
		want  StateRepresentation
	}{
		{
			name:  "connecting",
			state: connecting,
			want:  connectingRepr,
		},
		{
			name:  "connected",
			state: connected,
			want:  connectedRepr,
		},
		{
			name:  "unconnected",
			state: unconnected,
			want:  unconnectedRepr,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			csm := &ConnectionStateMachine{&tc.state}
			if got := csm.CurrentState(); got != tc.want {
				t.Errorf("expected CurrentState() == %s, got %s", tc.want, got)
			}
		})
	}
}
