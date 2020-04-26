package gobayeux

// Message represents a message received by a Bayeux client
type Message struct {
	// ID represents the identifier of the specific message
	ID string `json:"id"`
	// Channel is the Channel on which the message was sent
	Channel Channel `json:"channel"`
	// ClientID identifies a particular session via a session id token
	ClientID string `json:"clientId"`
	// Data contains the binary represnetation of the data
	Data MessageData `json:"data"`
	// Ext is TBD
	Ext map[string]interface{} `json:"ext,omitempty"`
}

// MessageData represents the JSON object which contains the binary
// representation of the data in a Bayeux Message.
// See also https://docs.cometd.org/current/reference/#_concepts_binary_data
type MessageData struct {
	// Data is the actual binary representation of the data
	Data string `json:"data"`
	// Last tells whether the `data` field is the last chunk of binary data
	Last bool `json:"last"`
	// Meta is an optional field that caries related additional metadata
	Meta map[string]string `json:"meta,omitempty"`
}
