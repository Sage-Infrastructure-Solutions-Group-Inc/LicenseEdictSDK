package licenseedict

import (
	"errors"
	"fmt"
)

// Failure codes for license validation errors.
// These are backward-compatible with GoLicenseCheck where applicable.
const (
	LicenseDecodeError      = "LICENSE_DECODE_ERROR"
	PubKeyDecodeError       = "PUBKEY_DECODE_ERROR"
	InvalidLicenseSignature = "INVALID_LICENSE_SIGNATURE"
	LicenseNotValidBefore   = "LICENSE_NOT_VALID_BEFORE"
	LicenseNotValidAfter    = "LICENSE_NOT_VALID_AFTER"
	LicenseRevoked          = "LICENSE_REVOKED"
	ServerUnreachable       = "SERVER_UNREACHABLE"
	SeatLimitReached        = "SEAT_LIMIT_REACHED"
	RenewalFailed           = "RENEWAL_FAILED"
)

// ValidationError is returned when license validation fails.
// It includes a machine-readable Code field for programmatic handling.
type ValidationError struct {
	Code    string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// Is enables errors.Is matching against documented sentinel errors.
// For example, errors.Is(err, ErrInvalidSignature) returns true when
// the ValidationError has Code == InvalidLicenseSignature.
func (e *ValidationError) Is(target error) bool {
	switch target {
	case ErrInvalidSignature:
		return e.Code == InvalidLicenseSignature
	case ErrTokenMalformed:
		return e.Code == LicenseDecodeError
	case ErrLicenseExpired:
		return e.Code == LicenseNotValidAfter
	case ErrLicenseRevoked:
		return e.Code == LicenseRevoked
	case ErrServerUnreachable:
		return e.Code == ServerUnreachable
	case ErrSeatLimitReached:
		return e.Code == SeatLimitReached
	case ErrRenewalDenied:
		return e.Code == RenewalFailed
	}
	return false
}

// Sentinel errors for common failure cases.
var (
	ErrNoPublicKey    = errors.New("licenseedict: no public key configured")
	ErrNoToken        = errors.New("licenseedict: no signed token provided")
	ErrNoServerURL    = errors.New("licenseedict: no server URL available")
	ErrClientClosed   = errors.New("licenseedict: client is closed")
	ErrAlreadyRunning = errors.New("licenseedict: heartbeat already running")
	ErrNotRunning     = errors.New("licenseedict: heartbeat not running")

	// Documented sentinel errors for errors.Is matching against ValidationError codes.
	ErrInvalidSignature  = errors.New("licenseedict: invalid license signature")
	ErrTokenMalformed    = errors.New("licenseedict: token could not be decoded")
	ErrLicenseExpired    = errors.New("licenseedict: license has expired")
	ErrLicenseRevoked    = errors.New("licenseedict: license has been revoked")
	ErrServerUnreachable = errors.New("licenseedict: server unreachable")
	ErrSeatLimitReached  = errors.New("licenseedict: seat limit reached")
	ErrRenewalDenied     = errors.New("licenseedict: renewal denied")
)
