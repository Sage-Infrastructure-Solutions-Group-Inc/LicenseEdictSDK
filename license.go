package licenseedict

import "time"

// License holds the decoded and validated license information.
type License struct {
	Valid       bool      `json:"valid"`
	LicenseID   string    `json:"license_id"`
	ProductID   string    `json:"product_id"`
	LicenseKey  string    `json:"license_key"`
	Licensee    string    `json:"licensee"`
	Plan        string    `json:"plan"`
	Features    []string  `json:"features"`
	MaxSeats    int       `json:"max_seats"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ServerURL   string    `json:"server_url"`
	SignedToken string    `json:"signed_token"`
}

// HasFeature returns true if the license includes the named feature.
func (l *License) HasFeature(feature string) bool {
	if l == nil {
		return false
	}
	for _, f := range l.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// IsExpired returns true if the license has an expiration date that has passed.
func (l *License) IsExpired() bool {
	if l == nil {
		return true
	}
	if l.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(l.ExpiresAt)
}
