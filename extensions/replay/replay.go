package replay

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	bayeux "github.com/sigmavirus24/gobayeux"
)

const (
	// ExtensionName is the name used by Salesforce for its Bayeux extensions
	ExtensionName string = "replay"
	eventKey      string = "event"
	replayIDKey   string = "replayId"

	unsupported int32 = iota
	supported
)

// Extension represents the structure of the Salesforce Bayeux
// Message Extension and manages the state
type Extension struct {
	supportedByServer *int32
	replayStore       IDStore
}

// IDStore stores and manages the channels and replay IDs for a bayeux
// server that supports the replay extension
type IDStore interface {
	Set(channel string, replayID int)
	Get(channel string) (int, bool)
	Delete(channel string)
	AsMap() map[string]int
}

// New creates a new extension instance
func New() *Extension {
	defaultVal := unsupported
	return &Extension{supportedByServer: &defaultVal}
}

// Outgoing attaches any additional metadata to a message
func (e *Extension) Outgoing(ms *bayeux.Message) {
	switch ms.Channel {
	case bayeux.MetaHandshake:
		ext := ms.GetExt(true)
		ext[ExtensionName] = true
	case bayeux.MetaSubscribe:
		if e.isSupported() {
			ext := ms.GetExt(true)
			ext[ExtensionName] = e.replayStore.AsMap()
		}
	}
}

// Incoming attaches any additional metadata to a message
func (e *Extension) Incoming(ms *bayeux.Message) {
	switch ms.Channel.Type() {
	case bayeux.MetaChannel:
		switch ms.Channel {
		case bayeux.MetaHandshake:
			ext := ms.GetExt(false)
			if ext != nil {
				isSupported, ok := ext[ExtensionName].(bool)
				if ok && isSupported {
					atomic.CompareAndSwapInt32(e.supportedByServer, unsupported, supported)
				}
			}
			return
		case bayeux.MetaUnsubscribe:
			for _, channel := range ms.Subscription {
				e.replayStore.Delete(string(channel))
			}
			return
		case bayeux.MetaConnect, bayeux.MetaSubscribe:
			return
		}
	case bayeux.BroadcastChannel:
		e.updateReplayID(ms)
	case bayeux.ServiceChannel:
		return
	}
}

// Registered is called after an extension has been successfully registered
func (e *Extension) Registered(extensionName string, client *bayeux.BayeuxClient) {
}

// Unregistered is called when an extension is unregistered
func (e *Extension) Unregistered() {
}

func (e *Extension) updateReplayID(ms *bayeux.Message) {
	data := make(map[string]interface{})

	var md *MessageData
	if err := json.Unmarshal(ms.Data, &md); err != nil {
		return
	}

	if err := json.Unmarshal([]byte(md.Data), &data); err != nil {
		return
	}
	event, ok := data[eventKey]
	if !ok {
		return
	}
	eventMap, ok := event.(map[string]interface{})
	if !ok {
		return
	}
	replayIDVal, ok := eventMap[replayIDKey]
	if !ok {
		return
	}

	replayID, ok := replayIDVal.(float64)
	if !ok {
		return
	}
	e.replayStore.Set(string(ms.Channel), int(replayID))
}

func (e *Extension) isSupported() bool {
	return atomic.LoadInt32(e.supportedByServer) == supported
}

// MessageData represents the JSON object which contains the binary
// representation of the data in a Bayeux Message.
//
// See also: https://docs.cometd.org/current/reference/#_concepts_binary_data
type MessageData struct {
	// Data is the actual binary representation of the data
	Data string `json:"data,omitempty"`
	// Last tells whether the `data` field is the last chunk of binary data
	Last bool `json:"last,omitempty"`
	// Meta is an optional field that caries related additional metadata
	Meta map[string]string `json:"meta,omitempty"`
}

// MapStorage implements the IDStore interface over a regular map with a
// RWMutex protecting the access
type MapStorage struct {
	store map[string]int
	lock  sync.RWMutex
}

// NewMapStorage creates a new MapStorage instance
func NewMapStorage() *MapStorage {
	return &MapStorage{store: make(map[string]int)}
}

// Set implements the IDStore interface
func (s *MapStorage) Set(channel string, replayID int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.store[channel] = replayID
}

// Get implements the IDStore interface
func (s *MapStorage) Get(channel string) (replayID int, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	replayID, ok = s.store[channel]
	return
}

// Delete implements the IDStore interface
func (s *MapStorage) Delete(channel string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.store, channel)
}

// AsMap implements the IDStore interface
func (s *MapStorage) AsMap() map[string]int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	replay := make(map[string]int)
	for k, v := range s.store {
		replay[k] = v
	}
	return replay
}
