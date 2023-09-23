package replay

import (
	"encoding/json"
	"testing"

	bayeux "github.com/sigmavirus24/gobayeux/v2"
)

func TestNewInitializesOurState(t *testing.T) {
	e := New(NewMapStorage())
	if *e.supportedByServer != unsupported {
		t.Error("extension is initialized incorrectly")
	}
}

func TestOutgoingMetaHandshake(t *testing.T) {
	e := New(NewMapStorage())
	e.Registered(ExtensionName, nil)
	m := bayeux.Message{Channel: bayeux.MetaHandshake}
	if m.Ext != nil {
		t.Fatal("ext should be nil but isn't")
	}
	e.Outgoing(&m)
	v, ok := m.Ext[ExtensionName]
	if !ok {
		t.Fatal("replay extension was not included in the handshake")
	}

	value, ok := v.(bool)
	if !ok {
		t.Fatal("couldn't coerce extension value to a bool")
	}
	if !value {
		t.Fatal("replay extension not set to true")
	}
}

func TestSupportedOutgoingMetaSubscribe(t *testing.T) {
	want := 1234
	e := New(NewMapStorage())
	*e.supportedByServer = supported
	e.Registered(ExtensionName, nil)
	e.replayStore = &MapStorage{store: map[string]int{"/foo/bar": want}}
	m := bayeux.Message{Channel: bayeux.MetaSubscribe}
	e.Outgoing(&m)

	v, ok := m.Ext[ExtensionName]
	if !ok {
		t.Fatal("replay extension was not included in the subscribe")
	}

	value, ok := v.(map[string]int)
	if !ok {
		t.Fatal("replay extension value couldn't coerce to a map")
	}
	if len(value) > 1 {
		t.Fatalf("too many values in replay extension map: %d", len(value))
	}
	if got := value["/foo/bar"]; want != got {
		t.Fatalf("replay map mismatch expected %d, got %d", want, got)
	}
}

func TestUnsupportedOutgoingMetaSubscribe(t *testing.T) {
	e := New(NewMapStorage())
	e.Registered(ExtensionName, nil)
	e.replayStore = &MapStorage{store: map[string]int{"/foo/bar": 1}}
	m := bayeux.Message{Channel: bayeux.MetaSubscribe}
	e.Outgoing(&m)

	_, ok := m.Ext[ExtensionName]
	if ok {
		t.Fatal("replay extension added data when it was unsupported")
	}
}

func TestDetectsItIsSupported(t *testing.T) {
	e := New(NewMapStorage())
	e.Registered(ExtensionName, nil)
	m := bayeux.Message{
		Channel: bayeux.MetaHandshake,
		Ext: map[string]interface{}{
			ExtensionName: true,
		},
	}
	e.Incoming(&m)
	if e.isSupported() != true {
		t.Error("replay extension didn't recognize that the server supported it")
	}
}

func TestIncomingMetaUnsubscribeRemovesChannel(t *testing.T) {
	e := New(NewMapStorage())
	e.replayStore = &MapStorage{store: map[string]int{
		"/foo/bar": 1,
		"/bar/*":   2,
		"/":        3,
	}}
	m := bayeux.Message{
		Channel:      bayeux.MetaUnsubscribe,
		Subscription: bayeux.Channel("/"),
	}
	e.Incoming(&m)

	if _, ok := e.replayStore.Get("/"); ok {
		t.Fatal("expected '/' to be removed from replay map but wasn't")
	}
}

func TestIncomingEdges(t *testing.T) {
	testCases := []struct {
		name    string
		channel bayeux.Channel
	}{
		{"connect", "/meta/connect"},
		{"subscribe", "/meta/subscribe"},
		{"service channel", "/service/foo"},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := New(NewMapStorage())
			e.Incoming(&bayeux.Message{Channel: tc.channel})
		})
	}
}

func TestIncomingUpdatesReplayIDStore(t *testing.T) {
	testCases := []struct {
		name string
		data string
		want int
	}{
		{
			name: "valid data updates the id in the store",
			data: `{"event": {"replayId": 2, "body": "data"}}`,
			want: 2,
		},
		{
			name: "replay id is not a 'Number'",
			data: `{"event": {"replayId": "abc", "body": "data"}}`,
			want: 1,
		},
		{
			name: "missing event in data",
			data: `{"not_an_event": {"replay": 2, "body": "data"}}`,
			want: 1,
		},
		{
			name: "non-object event",
			data: `{"event": [{"replay": 2, "body": "data"}]}`,
			want: 1,
		},
		{
			name: "no replay key in event object",
			data: `{"event": {"body": "data"]}`,
			want: 1,
		},
		{
			name: "message data isn't json",
			data: "just some plain text",
			want: 1,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			md := &MessageData{Data: tc.data}
			b, _ := json.Marshal(md)

			e := New(NewMapStorage())
			e.replayStore = &MapStorage{store: map[string]int{"/foo/bar": 1}}
			m := bayeux.Message{
				Channel: "/foo/bar",
				Data:    b,
			}
			e.Incoming(&m)
			got, ok := e.replayStore.Get("/foo/bar")
			if !ok {
				t.Fatal("expected /foo/bar to be in the replay store but it wasn't")
			}
			if got != tc.want {
				t.Fatalf("expected the replay id for /foo/bar to be 2 but got %d", got)
			}
		})
	}
}

func TestRegistered(t *testing.T) {
	e := New(NewMapStorage())
	e.Registered(ExtensionName, nil)
	// NOTE: These functions do nothing but test stubs are nice for test
	// coverage
}

func TestUnregistered(t *testing.T) {
	e := New(NewMapStorage())
	e.Unregistered()
	// NOTE: These functions do nothing but test stubs are nice for test
	// coverage
}

func TestMapStorageSet(t *testing.T) {
	s := NewMapStorage()
	want := 1
	s.Set("/foo/bar", want)
	if got, ok := s.Get("/foo/bar"); !ok || want != got {
		if !ok {
			t.Fatal("expected s.Set to store value but it didn't")
		}
		t.Fatalf("expected offset to be %d but got %d", want, got)
	}
}

func TestEmptyMapStorageGet(t *testing.T) {
	s := NewMapStorage()
	if _, ok := s.Get("/foo/bar"); ok {
		t.Fatal("expected s.Get(\"/foo/bar\") to not return ok")
	}
}

func TestMapStorageGet(t *testing.T) {
	want := 1
	s := &MapStorage{store: map[string]int{"/foo/bar": want}}
	if got, ok := s.Get("/foo/bar"); !ok || want != got {
		t.Fatalf("expected s.Get(\"/foo/bar\") = %d; got %d", want, got)
	}
}

func TestMapStorageDelete(t *testing.T) {
	s := &MapStorage{store: map[string]int{"/foo/bar": 1}}
	s.Delete("/foo/bar")
	if _, ok := s.Get("/foo/bar"); ok {
		t.Fatal("expected s.Get(\"/foo/bar\") to not return ok")
	}
}

func TestMapStorageAsMap(t *testing.T) {
	s := &MapStorage{store: map[string]int{"/foo/bar": 1234}}
	m := s.AsMap()
	if len(m) != 1 {
		t.Fatalf("expected len(m) = %d, got %d", 1, len(m))
	}
	if m["/foo/bar"] != 1234 {
		t.Fatalf("expected m[\"/foo/bar\"] = %d, got %d", 1234, m["/foo/bar"])
	}
}
