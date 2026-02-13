// Package licenseedict provides a Go SDK for the LicenseEdict licensing platform.
//
// It supports three usage patterns:
//
// Pattern 1 -- Simplified functions (new API):
//
//	license, err := licenseedict.CheckLicense(publicKeyB64, token)
//	if license.Valid {
//	    hasPro, _ := licenseedict.CheckFeature(publicKeyB64, token, "PRO")
//	}
//
// Pattern 2 -- Legacy functions (GoLicenseCheck-compatible):
//
//	pubKey, _ := licenseedict.DecodePublicKey(publicKeyB64)
//	license, err := licenseedict.CheckLicenseLegacy(token, pubKey, "MyApp", "MyCompany")
//	if license.Valid {
//	    hasPro := licenseedict.CheckFeatureLegacy("PRO", license)
//	}
//
// Pattern 3 -- Client-based (full features with heartbeat and renewal):
//
//	client, _ := licenseedict.NewClient(
//	    licenseedict.WithPublicKey(publicKeyB64),
//	    licenseedict.WithAppInfo("MyApp", "MyCompany"),
//	)
//	defer client.Close()
//	license, _ := client.Validate(token)
package licenseedict

import (
	"crypto/ed25519"
	"time"
)

// CheckLicense verifies a signed license token with the given base64-encoded
// public key. This is the simplified entry point for the documented API.
//
// The returned License.Valid indicates whether the license passed all checks.
// Even when Valid is false, the License struct is populated with decoded data.
// A non-nil error indicates a fundamental failure (decode error, missing key).
func CheckLicense(publicKey string, token string) (*License, error) {
	if publicKey == "" {
		return &License{}, ErrNoPublicKey
	}
	if token == "" {
		return &License{}, ErrNoToken
	}

	pubKey, err := DecodePublicKey(publicKey)
	if err != nil {
		return &License{}, err
	}

	payload, err := verifyToken(pubKey, token)
	if err != nil {
		// Try cache fallback
		cm := newCacheManager("", "", "", false)
		cached, cacheErr := cm.load()
		if cacheErr == nil && cached != nil {
			return cached, nil
		}
		return &License{}, err
	}

	license := payloadToLicense(payload, token, true)

	// Temporal validity checks
	now := time.Now()
	if !payload.IssuedAt.IsZero() && now.Before(payload.IssuedAt) {
		license.Valid = false
	}
	if !payload.ExpiresAt.IsZero() && now.After(payload.ExpiresAt) {
		license.Valid = false
	}

	// Cache the license
	cm := newCacheManager("", "", "", false)
	_ = cm.save(license)

	return license, nil
}

// CheckFeature returns true if the license for the given token includes the
// named feature. This is a convenience function that validates the license and
// checks the feature in one call.
func CheckFeature(publicKey string, token string, feature string) (bool, error) {
	license, err := CheckLicense(publicKey, token)
	if err != nil {
		return false, err
	}
	return license.HasFeature(feature), nil
}

// CheckLicenseLegacy verifies a signed license token with the given public key.
// This is the GoLicenseCheck-compatible entry point that accepts a raw
// ed25519.PublicKey and application info for cache directory naming.
//
// The returned License.Valid indicates whether the license passed all checks.
// Even when Valid is false, the License struct is populated with decoded data.
// A non-nil error indicates a fundamental failure (decode error, missing key).
func CheckLicenseLegacy(signedToken string, publicKey ed25519.PublicKey, appName, appPublisher string) (*License, error) {
	if publicKey == nil {
		return &License{}, ErrNoPublicKey
	}
	if signedToken == "" {
		return &License{}, ErrNoToken
	}

	payload, err := verifyToken(publicKey, signedToken)
	if err != nil {
		// Try cache fallback
		cm := newCacheManager(appName, appPublisher, "", false)
		cached, cacheErr := cm.load()
		if cacheErr == nil && cached != nil {
			return cached, nil
		}
		return &License{}, err
	}

	license := payloadToLicense(payload, signedToken, true)

	// Temporal validity checks
	now := time.Now()
	if !payload.IssuedAt.IsZero() && now.Before(payload.IssuedAt) {
		license.Valid = false
	}
	if !payload.ExpiresAt.IsZero() && now.After(payload.ExpiresAt) {
		license.Valid = false
	}

	// Cache the license
	cm := newCacheManager(appName, appPublisher, "", false)
	_ = cm.save(license)

	return license, nil
}

// CheckFeatureLegacy returns true if the license includes the named feature.
// This is a convenience wrapper for License.HasFeature, preserving the original
// GoLicenseCheck-compatible API.
func CheckFeatureLegacy(feature string, license *License) bool {
	return license.HasFeature(feature)
}
