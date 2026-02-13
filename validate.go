package licenseedict

import (
	"time"
)

const defaultRenewBefore = 7 * 24 * time.Hour

// Validate verifies the signed token using the configured public key,
// checks temporal validity, caches the result, and returns a License.
//
// When called without arguments, Validate uses the token stored via WithToken
// or from a previous Validate call. A token can also be passed explicitly.
//
// This method follows the offline-first philosophy: it never returns an error
// that should block the host application. Check license.Valid instead.
// On verification failure, it falls back to the cached license if available.
func (c *Client) Validate(signedToken ...string) (*License, error) {
	if c.closed {
		return &License{}, ErrClientClosed
	}

	// Use provided token, fall back to stored token
	token := ""
	if len(signedToken) > 0 && signedToken[0] != "" {
		token = signedToken[0]
	}
	if token == "" {
		c.mu.RLock()
		token = c.signedToken
		c.mu.RUnlock()
	}
	if token == "" {
		token = c.cfg.token
	}

	if token == "" {
		return &License{}, ErrNoToken
	}

	if c.cfg.publicKey == nil {
		return &License{}, ErrNoPublicKey
	}

	// Verify signature
	payload, err := verifyToken(c.cfg.publicKey, token)
	if err != nil {
		// Attempt cache fallback
		cached, cacheErr := c.cache.load()
		if cacheErr == nil && cached != nil {
			return cached, nil
		}
		return &License{}, err
	}

	license := payloadToLicense(payload, token, true)

	// Temporal checks
	now := time.Now()
	if !payload.IssuedAt.IsZero() && now.Before(payload.IssuedAt) {
		license.Valid = false
	}
	if !payload.ExpiresAt.IsZero() && now.After(payload.ExpiresAt) {
		license.Valid = false
	}

	// Update server URL from token if not explicitly set
	if c.cfg.serverURL == "" && license.ServerURL != "" {
		c.cfg.serverURL = license.ServerURL
	}

	// Store the current license and token
	c.mu.Lock()
	c.license = license
	c.signedToken = token
	c.mu.Unlock()

	// Cache the license
	_ = c.cache.save(license)

	// Trigger auto-renewal if approaching expiry
	c.maybeAutoRenew(license)

	return license, nil
}

// ValidateFromCache loads and returns the cached license without network calls
// or re-verification. Returns nil if no cached license exists.
func (c *Client) ValidateFromCache() (*License, error) {
	if c.closed {
		return nil, ErrClientClosed
	}

	cached, err := c.cache.load()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.license = cached
	if cached.SignedToken != "" {
		c.signedToken = cached.SignedToken
	}
	c.mu.Unlock()

	return cached, nil
}

// maybeAutoRenew checks if the license is approaching expiry and triggers
// a background renewal if auto-renewal is enabled.
func (c *Client) maybeAutoRenew(license *License) {
	if c.cfg.disableAutoRenew || c.cfg.offlineOnly {
		return
	}
	if license == nil || !license.Valid || license.ExpiresAt.IsZero() {
		return
	}

	threshold := c.cfg.renewBefore
	if threshold == 0 {
		threshold = defaultRenewBefore
	}

	timeLeft := time.Until(license.ExpiresAt)
	if timeLeft > threshold {
		return
	}

	// Spawn background renewal
	go func() {
		result, err := c.Renew()
		if err != nil {
			return
		}
		if c.cfg.onRenew != nil && result != nil {
			c.cfg.onRenew(result)
		}
	}()
}
