package licenseedict

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "LicenseEdictSDK-Go/1.0"
)

// httpClient wraps an *http.Client with SDK-specific defaults.
type httpClient struct {
	client    *http.Client
	userAgent string
}

func newHTTPClient(customClient *http.Client, timeout time.Duration, userAgent string) *httpClient {
	c := customClient
	if c == nil {
		t := timeout
		if t == 0 {
			t = defaultTimeout
		}
		c = &http.Client{Timeout: t}
	}

	ua := userAgent
	if ua == "" {
		ua = defaultUserAgent
	}

	return &httpClient{client: c, userAgent: ua}
}

func (h *httpClient) postJSON(url string, body interface{}, result interface{}) (int, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", h.userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}

	return resp.StatusCode, nil
}

func (h *httpClient) deleteJSON(url string, body interface{}, result interface{}) (int, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", h.userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}

	return resp.StatusCode, nil
}
