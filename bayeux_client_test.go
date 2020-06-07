package gobayeux

import "testing"

func TestClientState_GetClientID(t *testing.T) {
	want := "fakeClientID"
	state := clientState{clientID: want}
	got := state.GetClientID()
	if want != got {
		t.Errorf("error retrieving client ID; want %s got %s", want, got)
	}
}

func TestClientState_SetClientID(t *testing.T) {
	want := "fakeClientID"
	state := clientState{}
	state.SetClientID(want)
	if got := state.clientID; want != got {
		t.Errorf("error retrieving client ID; want %s got %s", want, got)
	}
}
