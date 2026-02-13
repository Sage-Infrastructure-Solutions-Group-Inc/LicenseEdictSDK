package licenseedict

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"time"
)

// tokenPayload mirrors the server's LicenseTokenPayload.
type tokenPayload struct {
	LicenseID  string    `json:"license_id"`
	ProductID  string    `json:"product_id"`
	LicenseKey string    `json:"license_key"`
	Licensee   string    `json:"licensee,omitempty"`
	Plan       string    `json:"plan"`
	Features   []string  `json:"features,omitempty"`
	MaxSeats   int       `json:"max_seats"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
	ServerURL  string    `json:"server_url,omitempty"`
}

// verifyToken verifies the Ed25519 signature and returns the decoded payload.
// Token format: base64(signature_64bytes + json_payload)
func verifyToken(pubKey ed25519.PublicKey, signedToken string) (*tokenPayload, error) {
	combined, err := base64.StdEncoding.DecodeString(signedToken)
	if err != nil {
		return nil, &ValidationError{
			Code:    LicenseDecodeError,
			Message: "failed to base64-decode token",
			Err:     err,
		}
	}

	if len(combined) <= ed25519.SignatureSize {
		return nil, &ValidationError{
			Code:    LicenseDecodeError,
			Message: "token too short",
		}
	}

	signature := combined[:ed25519.SignatureSize]
	payloadBytes := combined[ed25519.SignatureSize:]

	if !ed25519.Verify(pubKey, payloadBytes, signature) {
		return nil, &ValidationError{
			Code:    InvalidLicenseSignature,
			Message: "Ed25519 signature verification failed",
		}
	}

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, &ValidationError{
			Code:    LicenseDecodeError,
			Message: "failed to decode token payload",
			Err:     err,
		}
	}

	return &payload, nil
}

// decodeTokenPayload extracts the payload without verifying the signature.
// Useful for extracting server_url or license_key before full verification.
func decodeTokenPayload(signedToken string) (*tokenPayload, error) {
	combined, err := base64.StdEncoding.DecodeString(signedToken)
	if err != nil {
		return nil, &ValidationError{
			Code:    LicenseDecodeError,
			Message: "failed to base64-decode token",
			Err:     err,
		}
	}

	if len(combined) <= ed25519.SignatureSize {
		return nil, &ValidationError{
			Code:    LicenseDecodeError,
			Message: "token too short",
		}
	}

	payloadBytes := combined[ed25519.SignatureSize:]

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, &ValidationError{
			Code:    LicenseDecodeError,
			Message: "failed to decode token payload",
			Err:     err,
		}
	}

	return &payload, nil
}

// DecodePublicKey decodes a base64-encoded Ed25519 public key.
func DecodePublicKey(encoded string) (ed25519.PublicKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, &ValidationError{
			Code:    PubKeyDecodeError,
			Message: "failed to base64-decode public key",
			Err:     err,
		}
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, &ValidationError{
			Code:    PubKeyDecodeError,
			Message: "invalid public key length",
		}
	}
	return ed25519.PublicKey(decoded), nil
}

// payloadToLicense converts an internal token payload to a public License.
func payloadToLicense(p *tokenPayload, signedToken string, valid bool) *License {
	features := p.Features
	if features == nil {
		features = []string{}
	}
	return &License{
		Valid:       valid,
		LicenseID:   p.LicenseID,
		ProductID:   p.ProductID,
		LicenseKey:  p.LicenseKey,
		Licensee:    p.Licensee,
		Plan:        p.Plan,
		Features:    features,
		MaxSeats:    p.MaxSeats,
		IssuedAt:    p.IssuedAt,
		ExpiresAt:   p.ExpiresAt,
		ServerURL:   p.ServerURL,
		SignedToken: signedToken,
	}
}
