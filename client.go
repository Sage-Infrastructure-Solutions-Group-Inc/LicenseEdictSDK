package licenseedict

import (
	"sync"
)

// Client is the main SDK entry point for full-featured license management.
// Use NewClient to create an instance, and defer client.Close().
type Client struct {
	cfg         clientConfig
	cache       *cacheManager
	http        *httpClient
	license     *License
	signedToken string
	mu          sync.RWMutex
	hb          heartbeatState
	closed      bool

	// Events receives asynchronous status updates from background operations
	// such as heartbeats and renewals. Events are delivered non-blocking;
	// if the channel buffer is full, events are dropped silently.
	Events chan Event
}

// NewClient creates a new Client configured with the provided options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := clientConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	c := &Client{
		cfg:    cfg,
		cache:  newCacheManager(cfg.appName, cfg.appPublisher, cfg.cacheDir, cfg.disableCache),
		http:   newHTTPClient(cfg.httpClient, cfg.httpTimeout, cfg.userAgent),
		Events: make(chan Event, eventsChannelSize),
	}

	// If token is pre-configured, store it for later use by Validate()
	if cfg.token != "" {
		c.signedToken = cfg.token
	}

	return c, nil
}

// License returns the most recently validated license, or nil.
func (c *Client) License() *License {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.license
}

// Close releases resources. It stops the heartbeat but does NOT auto-checkout
// (the seat will expire via TTL on the server).
func (c *Client) Close() error {
	if c.closed {
		return nil
	}
	c.StopHeartbeat()
	c.closed = true
	close(c.Events)
	return nil
}

// resolveServerURL returns the server URL from config or the license token.
func (c *Client) resolveServerURL() string {
	if c.cfg.serverURL != "" {
		return c.cfg.serverURL
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.license != nil {
		return c.license.ServerURL
	}
	return ""
}
