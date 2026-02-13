package licenseedict

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	eventsChannelSize        = 16
)

// HeartbeatOptions configures the background heartbeat.
type HeartbeatOptions struct {
	InstanceID string
	UserHash   string
	Hostname   string
	IP         string
	UserAgent  string
}

// heartbeatState holds the running heartbeat goroutine's control channels.
type heartbeatState struct {
	mu       sync.Mutex
	running  bool
	stopCh   chan struct{}
	doneCh   chan struct{}
	opts     HeartbeatOptions
	interval time.Duration
}

// StartHeartbeat starts a background goroutine that sends periodic heartbeats.
// It returns the Events channel for listening to heartbeat status updates.
//
// HeartbeatOptions can be provided to configure instance details. If omitted,
// the instance ID from WithInstanceID is used and the heartbeat interval from
// WithHeartbeatInterval is applied.
//
// Events are delivered non-blocking to the returned channel. If the channel
// buffer is full, events are dropped silently.
func (c *Client) StartHeartbeat(opts ...HeartbeatOptions) (<-chan Event, error) {
	if c.closed {
		return nil, ErrClientClosed
	}

	c.hb.mu.Lock()
	defer c.hb.mu.Unlock()

	if c.hb.running {
		return nil, ErrAlreadyRunning
	}

	serverURL := c.resolveServerURL()
	if serverURL == "" {
		return nil, ErrNoServerURL
	}

	c.mu.RLock()
	token := c.signedToken
	c.mu.RUnlock()

	if token == "" {
		token = c.cfg.token
	}
	if token == "" {
		return nil, ErrNoToken
	}

	// Merge options
	var hbOpts HeartbeatOptions
	if len(opts) > 0 {
		hbOpts = opts[0]
	}
	if hbOpts.InstanceID == "" && c.cfg.instanceID != "" {
		hbOpts.InstanceID = c.cfg.instanceID
	}

	interval := c.cfg.heartbeatInterval
	if interval == 0 {
		interval = defaultHeartbeatInterval
	}

	c.hb.running = true
	c.hb.opts = hbOpts
	c.hb.interval = interval
	c.hb.stopCh = make(chan struct{})
	c.hb.doneCh = make(chan struct{})

	go c.heartbeatLoop(serverURL, token, c.hb.stopCh, c.hb.doneCh)
	return c.Events, nil
}

// StopHeartbeat stops the background heartbeat goroutine.
func (c *Client) StopHeartbeat() {
	c.hb.mu.Lock()
	defer c.hb.mu.Unlock()

	if !c.hb.running {
		return
	}

	close(c.hb.stopCh)
	<-c.hb.doneCh
	c.hb.running = false
}

// Checkout releases the seat on the server and stops the heartbeat.
func (c *Client) Checkout() error {
	if c.closed {
		return ErrClientClosed
	}

	// Stop heartbeat first
	c.StopHeartbeat()

	serverURL := c.resolveServerURL()
	if serverURL == "" {
		return ErrNoServerURL
	}

	c.mu.RLock()
	token := c.signedToken
	c.mu.RUnlock()

	if token == "" {
		return ErrNoToken
	}

	body := map[string]interface{}{
		"signed_token": token,
		"instance_id":  c.hb.opts.InstanceID,
	}
	if c.hb.opts.UserHash != "" {
		body["user_hash"] = c.hb.opts.UserHash
	}

	var resp struct {
		Status string `json:"status"`
	}

	url := fmt.Sprintf("%s/api/v1/concurrency/checkout", serverURL)
	statusCode, err := c.http.deleteJSON(url, body, &resp)
	if err != nil {
		return &ValidationError{Code: ServerUnreachable, Message: "checkout request failed", Err: err}
	}

	if statusCode != http.StatusOK {
		return &ValidationError{Code: ServerUnreachable, Message: fmt.Sprintf("checkout returned status %d", statusCode)}
	}

	c.emitEvent(Event{Type: EventSeatReleased, Message: "seat released"})
	return nil
}

func (c *Client) heartbeatLoop(serverURL, token string, stopCh, doneCh chan struct{}) {
	defer close(doneCh)

	// Send initial heartbeat immediately
	c.sendHeartbeat(serverURL, token)

	ticker := time.NewTicker(c.hb.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			c.sendHeartbeat(serverURL, token)

			// Adapt ticker if interval changed
			c.hb.mu.Lock()
			currentInterval := c.hb.interval
			c.hb.mu.Unlock()
			ticker.Reset(currentInterval)
		}
	}
}

func (c *Client) sendHeartbeat(serverURL, token string) {
	body := map[string]interface{}{
		"signed_token": token,
		"instance_id":  c.hb.opts.InstanceID,
		"metadata": map[string]string{
			"hostname":   c.hb.opts.Hostname,
			"ip":         c.hb.opts.IP,
			"user_agent": c.hb.opts.UserAgent,
			"user_hash":  c.hb.opts.UserHash,
		},
	}

	var resp HeartbeatStatus
	url := fmt.Sprintf("%s/api/v1/concurrency/heartbeat", serverURL)
	statusCode, err := c.http.postJSON(url, body, &resp)

	if err != nil {
		c.emitEvent(Event{Type: EventHeartbeatError, Message: err.Error()})
		return
	}

	switch statusCode {
	case http.StatusOK:
		c.emitEvent(Event{Type: EventHeartbeatOK, Message: "heartbeat accepted", Data: resp})
		// Adapt interval from server response
		if resp.HeartbeatInterval > 0 {
			newInterval := time.Duration(resp.HeartbeatInterval) * time.Second
			c.hb.mu.Lock()
			c.hb.interval = newInterval
			c.hb.mu.Unlock()
		}
	case http.StatusTooManyRequests:
		c.emitEvent(Event{Type: EventHeartbeatRejected, Message: "seat limit reached", Data: resp})
	default:
		c.emitEvent(Event{Type: EventHeartbeatError, Message: fmt.Sprintf("heartbeat returned status %d", statusCode), Data: resp})
	}
}

func (c *Client) emitEvent(e Event) {
	select {
	case c.Events <- e:
	default:
		// Drop event if channel is full (non-blocking)
	}
}
