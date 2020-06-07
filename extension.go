package gobayeux

// MessageExtender defines the interface that extensions are expected to
// implement
type MessageExtender interface {
	Outgoing(*Message)
	Incoming(*Message)
	Registered(extensionName string, client *BayeuxClient)
	Unregistered()
}
