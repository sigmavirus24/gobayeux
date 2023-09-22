// Package gobayeux provides both a low-level protocol client and a
// higher-level client that improves the ergonomics of talking to a server
// implementing the Bayeux Protocol.
//
// The best way to create a high-level client is with `NewClient`. Provided a
// server address for the server you're using, you can create a client like so
//
//	serverAddress := "https://localhost:8080/"
//	client := gobayeux.NewClient(serverAddress)
//
// You can also register customer HTTP transports with your client
//
//	transport := &http.Transport{
//		DialContext: (&net.Dialer{
//	  	Timeout:   3 * time.Second,
//	  	KeepAlive: 10 * time.Second,
//	  }).DialContext,
//	}
//	client := gobayeux.NewClient(serverAddress, gobayeux.WithHTTPTransport(transport))
//
// You can subscribe to a Bayeux Channel with a chan to receive messages on
//
//	recv := make(chan []gobayeux.Message)
//	client.Subscribe("example-channel", recv)
//
// You can also register extensions that you'd like to use with the server
// by implementing the MessageExtender interface and then passing it to the
// client
//
//	type Example struct {}
//	func (e *Example) Registered(name string, client *gobayeux.BayeuxClient) {}
//	func (e *Example) Unregistered() {}
//	func (e *Example) Outgoing(m *gobayeux.Message) {
//	   switch m.Channel {
//	   case gobayeux.MetaHandshake:
//	   	ext := m.GetExt(true)
//	   	ext["example"] = true
//	   }
//	}
//	func (e *Example) Incoming(m *gobayeux.Message) {}
//
//	e := &Example{}
//	client.UseExtension(e)
package gobayeux
