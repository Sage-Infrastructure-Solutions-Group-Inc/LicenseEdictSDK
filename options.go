package licenseedict

import (
	"crypto/ed25519"
	"log/slog"
	"net/http"
	"time"
)

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	publicKey         ed25519.PublicKey
	publicKeyStr      string // base64-encoded, for convenience API
	token             string // stored token for Validate()
	serverURL         string
	appName           string
	appPublisher      string
	httpClient        *http.Client
	httpTimeout       time.Duration
	cacheDir          string
	disableCache      bool
	offlineOnly       bool
	userAgent         string
	instanceID        string
	heartbeatInterval time.Duration
	renewBefore       time.Duration
	disableAutoRenew  bool
	onRenew           func(*License)
	logger            *slog.Logger
}

// WithPublicKey sets the Ed25519 public key for offline verification.
// Accepts a base64-encoded string which is decoded internally.
func WithPublicKey(key string) Option {
	return func(c *clientConfig) {
		c.publicKeyStr = key
		decoded, err := DecodePublicKey(key)
		if err == nil {
			c.publicKey = decoded
		}
	}
}

// WithPublicKeyRaw sets an already-decoded Ed25519 public key.
func WithPublicKeyRaw(key ed25519.PublicKey) Option {
	return func(c *clientConfig) {
		c.publicKey = key
	}
}

// WithToken sets the license token used by Validate() when called without arguments.
func WithToken(token string) Option {
	return func(c *clientConfig) {
		c.token = token
	}
}

// WithServerURL overrides the server URL extracted from the token.
func WithServerURL(url string) Option {
	return func(c *clientConfig) {
		c.serverURL = url
	}
}

// WithAppInfo sets the application name and publisher for cache directory naming.
func WithAppInfo(name, publisher string) Option {
	return func(c *clientConfig) {
		c.appName = name
		c.appPublisher = publisher
	}
}

// WithHTTPClient sets a custom HTTP client for server communication.
func WithHTTPClient(client *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

// WithHTTPTimeout sets the timeout for HTTP requests (default 10s).
func WithHTTPTimeout(d time.Duration) Option {
	return func(c *clientConfig) {
		c.httpTimeout = d
	}
}

// WithCacheDir overrides the cache directory path.
func WithCacheDir(dir string) Option {
	return func(c *clientConfig) {
		c.cacheDir = dir
	}
}

// WithoutCache disables license caching entirely.
func WithoutCache() Option {
	return func(c *clientConfig) {
		c.disableCache = true
	}
}

// WithOfflineOnly disables all server communication.
func WithOfflineOnly() Option {
	return func(c *clientConfig) {
		c.offlineOnly = true
	}
}

// WithUserAgent sets a custom User-Agent string for HTTP requests.
func WithUserAgent(ua string) Option {
	return func(c *clientConfig) {
		c.userAgent = ua
	}
}

// WithInstanceID sets a custom instance ID for seat tracking.
// If not set, an ID is auto-generated.
func WithInstanceID(id string) Option {
	return func(c *clientConfig) {
		c.instanceID = id
	}
}

// WithHeartbeatInterval sets the heartbeat interval (default: 60s).
func WithHeartbeatInterval(d time.Duration) Option {
	return func(c *clientConfig) {
		c.heartbeatInterval = d
	}
}

// WithRenewBefore sets the auto-renewal threshold (default: 7 days before expiry).
func WithRenewBefore(d time.Duration) Option {
	return func(c *clientConfig) {
		c.renewBefore = d
	}
}

// WithoutAutoRenew disables automatic license renewal.
func WithoutAutoRenew() Option {
	return func(c *clientConfig) {
		c.disableAutoRenew = true
	}
}

// WithOnRenew registers a callback invoked after successful auto-renewal.
func WithOnRenew(fn func(*License)) Option {
	return func(c *clientConfig) {
		c.onRenew = fn
	}
}

// WithLogger sets a custom structured logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *clientConfig) {
		c.logger = l
	}
}
