package licenseedict

import (
	"fmt"
	"net/http"
)

// Renew exchanges the current signed token for a renewed one via the server.
// On success, the client's internal license and token are updated and the
// new License is returned. The RenewalResult details are emitted as an
// EventLicenseRenewed event on the Events channel.
func (c *Client) Renew() (*License, error) {
	if c.closed {
		return nil, ErrClientClosed
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

	body := map[string]string{
		"signed_token": token,
	}

	var result RenewalResult
	url := fmt.Sprintf("%s/api/v1/licenses/renew", serverURL)
	statusCode, err := c.http.postJSON(url, body, &result)
	if err != nil {
		return nil, &ValidationError{Code: RenewalFailed, Message: "renewal request failed", Err: err}
	}

	if statusCode != http.StatusOK {
		return nil, &ValidationError{Code: RenewalFailed, Message: fmt.Sprintf("renewal returned status %d", statusCode)}
	}

	// Re-validate with the new token
	if result.SignedToken != "" && c.cfg.publicKey != nil {
		newLicense, validateErr := c.Validate(result.SignedToken)
		if validateErr == nil && newLicense.Valid {
			c.emitEvent(Event{Type: EventLicenseRenewed, Message: "license renewed", Data: result})
			return newLicense, nil
		}
	}

	// If we can't re-validate, return a license with token info from result
	license := &License{
		SignedToken: result.SignedToken,
	}
	// Store the new token
	if result.SignedToken != "" {
		c.mu.Lock()
		c.signedToken = result.SignedToken
		c.mu.Unlock()
	}

	return license, nil
}

// RenewResult exchanges the current signed token for a renewed one via the
// server and returns the raw RenewalResult from the server response.
// This is the legacy return type; prefer Renew() which returns *License.
func (c *Client) RenewResult() (*RenewalResult, error) {
	if c.closed {
		return nil, ErrClientClosed
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

	body := map[string]string{
		"signed_token": token,
	}

	var result RenewalResult
	url := fmt.Sprintf("%s/api/v1/licenses/renew", serverURL)
	statusCode, err := c.http.postJSON(url, body, &result)
	if err != nil {
		return nil, &ValidationError{Code: RenewalFailed, Message: "renewal request failed", Err: err}
	}

	if statusCode != http.StatusOK {
		return nil, &ValidationError{Code: RenewalFailed, Message: fmt.Sprintf("renewal returned status %d", statusCode)}
	}

	// Re-validate with the new token to update internal state
	if result.SignedToken != "" && c.cfg.publicKey != nil {
		newLicense, validateErr := c.Validate(result.SignedToken)
		if validateErr == nil && newLicense.Valid {
			c.emitEvent(Event{Type: EventLicenseRenewed, Message: "license renewed", Data: result})
		}
	}

	return &result, nil
}
