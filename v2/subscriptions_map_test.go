package gobayeux

import "testing"

func TestSubscriptionsMap_Add(t *testing.T) {
	sm := newSubscriptionsMap()
	want := make(chan []Message)
	defer close(want)
	if err := sm.Add("/foo/bar", want); err != nil {
		t.Errorf("expected successful addition but got err %q", err)
	}

	got, ok := sm.subs["/foo/bar"]
	if !ok {
		t.Error("channel was not registered properly")
	}

	if want != got {
		t.Error("chan received was not the chan registered")
	}
}

func TestSubscriptionsMap_Remove(t *testing.T) {
	sm := newSubscriptionsMap()
	want := make(chan []Message)
	defer close(want)
	if err := sm.Add("/foo/bar", want); err != nil {
		t.Errorf("unable to add subscription for test: %q", err)
	}

	if ls := len(sm.subs); ls != 1 {
		t.Errorf("expected ls to be 1, got %d", ls)
	}

	sm.Remove("/foo/bar")

	if ls := len(sm.subs); ls != 0 {
		t.Errorf("expected ls to be 0, got %d", ls)
	}
}

func TestSubscriptionsMap_Get(t *testing.T) {
	sm := newSubscriptionsMap()
	if _, err := sm.Get("/foo/bar"); err == nil {
		t.Error("expected '/foo/bar' to not have a subscription, but had one")
	}

	want := make(chan []Message)
	sm.subs["/foo/bar"] = want
	if got, err := sm.Get("/foo/bar"); want != got {
		if err != nil {
			t.Errorf("expected Get(\"/foo/bar\") to return without error but got %q", err)
		} else {
			t.Error("chan retrieved was not the chan registered")
		}
	}
}

func BenchmarkSubscriptionsMapAddToEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sm := newSubscriptionsMap()
		_ = sm.Add("/foo/bar", nil)
	}
}

func BenchmarkSubscriptionsMapAddNewToNonEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sm := &subscriptionsMap{
			subs: map[Channel](chan []Message){
				"/":           nil,
				"/foo":        nil,
				"/bar":        nil,
				"/baz":        nil,
				"/frob":       nil,
				"/foo/baz":    nil,
				"/foo/frob":   nil,
				"/frob/foo":   nil,
				"/frob/baz":   nil,
				"/bar/baz":    nil,
				"/baz/bar":    nil,
				"/*":          nil,
				"/foo/*":      nil,
				"/bar/*":      nil,
				"/baz/*":      nil,
				"/frob/*":     nil,
				"/foo/baz/*":  nil,
				"/foo/frob/*": nil,
				"/frob/foo/*": nil,
				"/frob/baz/*": nil,
				"/bar/baz/*":  nil,
				"/baz/bar/*":  nil,
			},
		}
		_ = sm.Add("/foo/bar", nil)
	}
}

func BenchmarkSubscriptionsMapAddDuplicate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sm := &subscriptionsMap{subs: map[Channel](chan []Message){"/foo/bar": nil}}
		_ = sm.Add("/foo/bar", nil)
	}
}
