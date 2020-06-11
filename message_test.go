package gobayeux

import (
	"testing"
	"time"
)

func TestMessage_TimestampAsTime(t *testing.T) {
	m := Message{Timestamp: "2020-05-01T06:28:51.00"}
	got, err := m.TimestampAsTime()
	if err != nil {
		t.Errorf("expected a valid timestamp, got err %q", err)
	}
	if want := time.Date(2020, time.May, 1, 6, 28, 51, 0, time.UTC); want != got {
		t.Errorf("unexpected time parse; want %v, got %v", want, got)
	}
}

func TestMessage_ParseError(t *testing.T) {
	testCases := []struct {
		name      string
		errorStr  string
		expected  MessageError
		shouldErr bool
	}{
		// Examples taken from specification
		{
			"no error args",
			"401::No client ID",
			MessageError{401, []string{""}, "No client ID"},
			false,
		},
		{
			"one nonsense error arg",
			"402:xj3sjdsjdsjad:Unknown Client ID",
			MessageError{402, []string{"xj3sjdsjdsjad"}, "Unknown Client ID"},
			false,
		},
		{
			"two args",
			"403:xj3sjdsjdsjad,/foo/bar:Subscription denied",
			MessageError{403, []string{"xj3sjdsjdsjad", "/foo/bar"}, "Subscription denied"},
			false,
		},
		{
			"one channel name arg",
			"404:/foo/bar:Unknown Channel",
			MessageError{404, []string{"/foo/bar"}, "Unknown Channel"},
			false,
		},
		// Following cases aren't from the specification directly
		{
			"invalid status code",
			"4o4:/foo/bar:Broken Error Code",
			MessageError{},
			true,
		},
		{
			"invalid error string",
			"404-/foo/bar-Unknown Channel",
			MessageError{},
			true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			m := Message{Error: tc.errorStr}
			got, err := m.ParseError()
			if err != nil && tc.shouldErr {
				return
			}
			if err != nil && !tc.shouldErr {
				t.Errorf("expected a parsed MessageError but got an err: %q", err)
			}
			if err == nil && tc.shouldErr {
				t.Error("expected an error but didn't get one")
			}

			want := tc.expected
			if want.ErrorCode != got.ErrorCode {
				t.Errorf("error parsing error code; want %v, got %v", want.ErrorCode, got.ErrorCode)
			}

			if want.ErrorMessage != got.ErrorMessage {
				t.Errorf("error parsing error message; want %v, got %v", want.ErrorMessage, got.ErrorMessage)
			}

			if len(want.ErrorArgs) != len(got.ErrorArgs) {
				t.Errorf("error parsing error args (found different lengths); want %v, got %v", want.ErrorArgs, got.ErrorArgs)
			}

			for index, arg := range want.ErrorArgs {
				if arg != got.ErrorArgs[index] {
					t.Errorf("error parsing error args (found different items at same position %d); want %v, got %v", index, want.ErrorArgs, got.ErrorArgs)
				}
			}
		})
	}
}

func TestMessage_GetExt(t *testing.T) {
	testCases := []struct {
		name         string
		message      *Message
		shouldCreate bool
		want         map[string]interface{}
	}{
		{
			name:         "nil extension is initialized as a map with create=true",
			message:      &Message{},
			shouldCreate: true,
			want:         make(map[string]interface{}),
		},
		{
			name:         "nil extension is not initialized with create=false",
			message:      &Message{},
			shouldCreate: false,
			want:         nil,
		},
		{
			name:         "non-nil extension is not overwritten with create=true",
			message:      &Message{Ext: map[string]interface{}{"foo": "bar"}},
			shouldCreate: true,
			want:         map[string]interface{}{"foo": "bar"},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.message.GetExt(tc.shouldCreate)
			if tc.want == nil && got != nil {
				t.Errorf("expected GetExt(%v) to return nil, got %v", tc.shouldCreate, got)
			}
			if tc.want != nil && got == nil {
				t.Errorf("expected GetExt(%v) to return %v, got nil", tc.shouldCreate, tc.want)
			}
			if len(tc.want) == len(got) {
				for k, vi := range tc.want {
					wantv, _ := vi.(string)
					gotv, _ := got[k].(string)
					if wantv != gotv {
						t.Errorf("expected Ext[%s] == %s, got %s", k, wantv, gotv)
					}
				}
			}
		})
	}
}

func TestAdvice_MustNotRetryOrHandshake(t *testing.T) {
	testCases := []struct {
		name      string
		reconnect string
		expected  bool
	}{
		{
			"reconnect advice is none",
			"none",
			true,
		},
		{
			"reconnect advice is retry",
			"retry",
			false,
		},
		{
			"reconnect advice is handshake",
			"handshake",
			false,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			a := Advice{Reconnect: tc.reconnect}
			if got, want := a.MustNotRetryOrHandshake(), tc.expected; want != got {
				t.Errorf("expected MustNotRetryOrHandshake() = %v, got %v", want, got)
			}
		})
	}
}

func TestAdvice_ShouldRetry(t *testing.T) {
	testCases := []struct {
		name      string
		reconnect string
		expected  bool
	}{
		{
			"reconnect advice is none",
			"none",
			false,
		},
		{
			"reconnect advice is retry",
			"retry",
			true,
		},
		{
			"reconnect advice is handshake",
			"handshake",
			false,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			a := Advice{Reconnect: tc.reconnect}
			if got, want := a.ShouldRetry(), tc.expected; want != got {
				t.Errorf("expected ShouldRetry() = %v, got %v", want, got)
			}
		})
	}
}

func TestAdvice_ShouldHandshake(t *testing.T) {
	testCases := []struct {
		name      string
		reconnect string
		expected  bool
	}{
		{
			"reconnect advice is none",
			"none",
			false,
		},
		{
			"reconnect advice is retry",
			"retry",
			false,
		},
		{
			"reconnect advice is handshake",
			"handshake",
			true,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			a := Advice{Reconnect: tc.reconnect}
			if got, want := a.ShouldHandshake(), tc.expected; want != got {
				t.Errorf("expected ShouldHandshake() = %v, got %v", want, got)
			}
		})
	}
}

func TestAdvice_TimeoutAsDuration(t *testing.T) {
	testCases := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			"two seconds",
			2000,
			time.Duration(2) * time.Second,
		},
		{
			"two hundred milliseconds",
			200,
			time.Duration(200) * time.Millisecond,
		},
		{
			"three minutes",
			180000,
			time.Duration(3) * time.Minute,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			a := Advice{Timeout: tc.timeout}
			if got, want := a.TimeoutAsDuration(), tc.expected; want != got {
				t.Errorf("expected TimeoutAsDuration() = %v, got %v", want, got)
			}
		})
	}
}

func TestAdvice_IntervalAsDuration(t *testing.T) {
	testCases := []struct {
		name     string
		interval int
		expected time.Duration
	}{
		{
			"two seconds",
			2000,
			time.Duration(2) * time.Second,
		},
		{
			"two hundred milliseconds",
			200,
			time.Duration(200) * time.Millisecond,
		},
		{
			"three minutes",
			180000,
			time.Duration(3) * time.Minute,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			a := Advice{Interval: tc.interval}
			if got, want := a.IntervalAsDuration(), tc.expected; want != got {
				t.Errorf("expected IntervalAsDuration() = %v, got %v", want, got)
			}
		})
	}
}
