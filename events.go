package licenseedict

// EventType identifies the kind of asynchronous event.
type EventType int

const (
	// EventHeartbeatOK indicates a successful heartbeat.
	EventHeartbeatOK EventType = iota
	// EventHeartbeatRejected indicates the server rejected the heartbeat (seat limit).
	EventHeartbeatRejected
	// EventHeartbeatError indicates a heartbeat network or server error.
	EventHeartbeatError
	// EventSeatReleased indicates a seat was released via checkout.
	EventSeatReleased
	// EventLicenseRenewed indicates a license was successfully renewed.
	EventLicenseRenewed
	// EventServerUnreachable indicates the server could not be reached.
	EventServerUnreachable
)

// Event carries information about an asynchronous SDK operation.
type Event struct {
	Type    EventType
	Message string
	Data    interface{}
}

// HeartbeatStatus contains the server's response to a heartbeat.
type HeartbeatStatus struct {
	Status            string `json:"status"`
	ActiveSessions    int    `json:"active_sessions"`
	MaxSessions       int    `json:"max_sessions"`
	RemainingSessions int    `json:"remaining_sessions"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
	GracePeriod       int    `json:"grace_period"`
	LicenseID         string `json:"license_id"`
	ProductID         string `json:"product_id"`
}

// RenewalResult contains the server's response to a renewal request.
type RenewalResult struct {
	Status            string `json:"status"`
	SignedToken       string `json:"signed_token"`
	IssuedAt          string `json:"issued_at"`
	ExpiresAt         string `json:"expires_at"`
	PreviousExpiresAt string `json:"previous_expires_at"`
}
