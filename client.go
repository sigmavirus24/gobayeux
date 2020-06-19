package gobayeux

import (
	"context"
	"time"
)

// Client is a high-level abstraction
type Client struct {
	client                    *BayeuxClient
	subscriptions             *subscriptionsMap
	timer                     *time.Timer
	subscribeRequestChannel   chan subscriptionRequest
	unsubscribeRequestChannel chan Channel
	connectMessageChannel     chan []Message
	connectRequestChannel     chan struct{}
	handshakeRequestChannel   chan struct{}
	shutdown                  chan struct{}
}

// NewClient creates a new high-level client
func NewClient(serverAddress string) (*Client, error) {
	bc, err := NewBayeuxClient(nil, serverAddress)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:                    bc,
		subscriptions:             newSubscriptionsMap(),
		timer:                     time.NewTimer(time.Millisecond),
		subscribeRequestChannel:   make(chan subscriptionRequest, 10),
		unsubscribeRequestChannel: make(chan Channel, 10),
		connectRequestChannel:     make(chan struct{}, 1),
		connectMessageChannel:     make(chan []Message, 5),
		handshakeRequestChannel:   make(chan struct{}),
		shutdown:                  make(chan struct{}),
	}, nil
}

// Subscribe queues a request to subscribe to a new channel from the server
func (c *Client) Subscribe(ch Channel, receiving chan []Message) {
	c.subscribeRequestChannel <- subscriptionRequest{ch, receiving}
}

// Start begins the background process that talks to the server
func (c *Client) Start(ctx context.Context) <-chan error {
	errors := make(chan error)
	go c.start(ctx, errors)
	return errors
}

// Disconnect issues a /meta/disconnect request to the Bayeux server and then
// cleans up channels and our timer.
func (c *Client) Disconnect(ctx context.Context) error {
	_, err := c.client.Disconnect(ctx)
	close(c.subscribeRequestChannel)
	close(c.unsubscribeRequestChannel)
	close(c.connectRequestChannel)
	close(c.connectMessageChannel)
	close(c.handshakeRequestChannel)
	c.timer.Stop()
	return err
}

// Publish is not yet implemented. When implemented, it will - in a separate thread
// from the polling task - publish messages to the Bayeux Server.
// See also
// https://docs.cometd.org/current/reference/#_two_connection_operation
func (c *Client) Publish(ctx context.Context, messages []Message) error {
	// TODO:
	// * Locking mechanism to ensure only one outstanding Publish request at a
	//   time
	// * Ensure that this separate from Start()/poll()
	// * Implement Publish() in *BayeuxClient
	panic("Publish() is not yet implemented")
}

func (c *Client) start(ctx context.Context, errors chan error) {
	if _, err := c.client.Handshake(ctx); err != nil {
		errors <- err
		return
	}

	subReqs, channels := c.getSubscriptionRequests()

	if _, err := c.client.Subscribe(ctx, channels); err != nil {
		errors <- err
		return
	}

	_ = c.subscriptions.Add(MetaConnect, c.connectMessageChannel)
	for _, subReq := range subReqs {
		if err := c.subscriptions.Add(subReq.subscription, subReq.msgChan); err != nil {
			errors <- err
			return
		}
	}

	c.connectRequestChannel <- struct{}{}

	if err := c.poll(ctx); err != nil {
		errors <- err
		return
	}

	if _, err := c.client.Disconnect(ctx); err != nil {
		errors <- err
		return
	}
}

func (c *Client) poll(ctx context.Context) error {
_poll_loop:
	for {
		select {
		case <-c.shutdown: // When the user calls the Disconnect() method
			break _poll_loop
		case <-ctx.Done(): // When the user cancel's the Start() context
			if err := ctx.Err(); err != nil {
				return err
			}
			break _poll_loop
		case subReq := <-c.subscribeRequestChannel:
			// Let's attempt to drain the channel before sending a
			// /meta/unsubscribe request to more efficiently use HTTP
			// requests
			subReqs, channels := c.getSubscriptionRequests()
			subReqs = append(subReqs, subReq)
			channels = append(channels, subReq.subscription)
			// TODO: Find a way to consolidate this logic and the logic in
			// start()
			if _, err := c.client.Subscribe(ctx, channels); err != nil {
				return err
			}

			for _, subReq := range subReqs {
				if err := c.subscriptions.Add(subReq.subscription, subReq.msgChan); err != nil {
					return err
				}
			}

			c.connectRequestChannel <- struct{}{}

		case unsubReq := <-c.unsubscribeRequestChannel:
			channels := c.getUnsubscriptionRequests()
			channels = append(channels, unsubReq)
			if _, err := c.client.Unsubscribe(ctx, channels); err != nil {
				return err
			}

			for _, channel := range channels {
				c.subscriptions.Remove(channel)
			}

		case ms := <-c.connectMessageChannel:
			for _, m := range ms {
				if m.Advice.ShouldHandshake() {
					c.handshakeRequestChannel <- struct{}{}
				}
				c.timer.Reset(m.Advice.IntervalAsDuration())
			}

		case <-c.connectRequestChannel:
			ms, err := c.client.Connect(ctx)
			if err != nil {
				return err
			}
			batch := make([]Message, 0)
			lastChannel := emptyChannel
			for _, m := range ms {
				switch lastChannel {
				case emptyChannel:
					lastChannel = m.Channel
					batch = append(batch, m)
				case m.Channel:
					batch = append(batch, m)
				default:
					msgChan, err := c.subscriptions.Get(lastChannel)
					if err != nil {
						return err
					}
					msgChan <- batch
					lastChannel = m.Channel
					batch = append([]Message(nil), m)
				}
			}
		case <-c.timer.C:
			c.connectRequestChannel <- struct{}{}
		}
	}
	return nil
}

func (c *Client) getSubscriptionRequests() ([]subscriptionRequest, []Channel) {
	subscriptionRequests := make([]subscriptionRequest, 0)
	channels := make([]Channel, 0)
	timer := time.After(1 * time.Second)

_get_subs_for_loop:
	for {
		select {
		case req := <-c.subscribeRequestChannel:
			subscriptionRequests = append(subscriptionRequests, req)
			channels = append(channels, req.subscription)
		case <-timer:
			break _get_subs_for_loop
		}
	}
	return subscriptionRequests, channels
}

func (c *Client) getUnsubscriptionRequests() []Channel {
	unsubscriptionRequests := make([]Channel, 0)
	timer := time.After(1 * time.Second)

_get_unsubs_for_loop:
	for {
		select {
		case req := <-c.unsubscribeRequestChannel:
			unsubscriptionRequests = append(unsubscriptionRequests, req)
		case <-timer:
			break _get_unsubs_for_loop
		}
	}
	return unsubscriptionRequests
}

type subscriptionRequest struct {
	subscription Channel
	msgChan      chan []Message
}
