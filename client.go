package gobayeux

import (
	"context"
	"fmt"
	"sync"
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
		connectMessageChannel:     make(chan []Message),
		handshakeRequestChannel:   make(chan struct{}, 1),
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
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return err
			}
			break _poll_loop
		case subReq := <-c.subscribeRequestChannel:
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

type subscriptionsMap struct {
	lock sync.RWMutex
	subs map[Channel]chan []Message
}

func newSubscriptionsMap() *subscriptionsMap {
	return &subscriptionsMap{subs: make(map[Channel]chan []Message)}
}

func (sm *subscriptionsMap) Add(channel Channel, ms chan []Message) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	if _, ok := sm.subs[channel]; !ok {
		sm.subs[channel] = ms
		return nil
	}
	return fmt.Errorf("channel '%s' already subscribed", channel)
}

func (sm *subscriptionsMap) Remove(channel Channel) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	delete(sm.subs, channel)
}

func (sm *subscriptionsMap) Get(channel Channel) (chan []Message, error) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	ms, ok := sm.subs[channel]
	if !ok {
		return nil, fmt.Errorf("channel '%s' has no subscriptions", channel)
	}
	return ms, nil
}
