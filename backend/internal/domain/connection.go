package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Connection-related errors
var (
	ErrConnectionNotFound      = errors.New("connection not found")
	ErrConnectionAlreadyExists = errors.New("connection already exists")
	ErrInvalidInvitationToken  = errors.New("invalid invitation token")
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrSelfConnection          = errors.New("cannot connect to yourself")
	ErrInvalidQRCode           = errors.New("invalid QR code")
	ErrQRCodeExpired           = errors.New("QR code expired")
	ErrInvalidSignature        = errors.New("invalid signature")
)

// ConnectionStatus defines the possible statuses for a connection
type ConnectionStatus string

const (
	ConnectionStatusPending  ConnectionStatus = "pending"
	ConnectionStatusAccepted ConnectionStatus = "accepted"
	ConnectionStatusRejected ConnectionStatus = "rejected"
	ConnectionStatusBlocked  ConnectionStatus = "blocked"
)

// Connection represents a 1:1 connection between two users
type Connection struct {
	ID              uuid.UUID        `json:"id"`
	UserAID         uuid.UUID        `json:"user_a_id"`
	UserBID         uuid.UUID        `json:"user_b_id"`
	Status          ConnectionStatus `json:"status"`
	RequestedBy     uuid.UUID        `json:"requested_by"`
	InvitationToken string           `json:"invitation_token,omitempty"`
	ExpiresAt       *time.Time       `json:"expires_at,omitempty"`
	AcceptedAt      *time.Time       `json:"accepted_at,omitempty"`
	RejectedAt      *time.Time       `json:"rejected_at,omitempty"`
	BlockedAt       *time.Time       `json:"blocked_at,omitempty"`
	BlockReason     string           `json:"block_reason,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// QRCodePayload represents the data encoded in a QR code (deprecated, kept for compatibility)
type QRCodePayload struct {
	UserID    string `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
}

// ConnectionRepository defines the interface for connection persistence
type ConnectionRepository interface {
	Create(connection *Connection) error
	GetByID(id uuid.UUID) (*Connection, error)
	GetByToken(token string) (*Connection, error)
	GetByUserID(userID uuid.UUID) ([]*Connection, error)
	GetByUserPair(userAID, userBID uuid.UUID) (*Connection, error)
	Update(connection *Connection) error
	Delete(id uuid.UUID) error
}

// ConnectionService defines the interface for connection business logic
type ConnectionService interface {
	// Invitation-based connection methods (new)
	CreateInvitation(userID uuid.UUID) (*Connection, string, error)
	ValidateInvitation(token string) (*Connection, error)
	AcceptInvitation(userID uuid.UUID, token string) error
	RejectInvitation(userID uuid.UUID, token string) error
	GetConnections(userID uuid.UUID) ([]*Connection, error)
	Disconnect(connectionID uuid.UUID, userID uuid.UUID) error

	// QR code methods (deprecated, kept for compatibility)
	GenerateQRCode(userID uuid.UUID) (*QRCodePayload, error)
	ScanQRCode(scannerID uuid.UUID, payload *QRCodePayload) (*Connection, error)
	ListConnections(userID uuid.UUID) ([]Connection, error)
	RemoveConnection(userID, connectionID uuid.UUID) error
}

// GenerateQRRequest represents a QR code generation request (deprecated)
type GenerateQRRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// ScanQRRequest represents a QR code scan request (deprecated)
type ScanQRRequest struct {
	Payload QRCodePayload `json:"payload" validate:"required"`
}

// ConnectionResponse represents a connection in API responses
type ConnectionResponse struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	FriendID   string     `json:"friend_id"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

// ExtendedConnectionResponse includes invitation details
type ExtendedConnectionResponse struct {
	ID              string     `json:"id"`
	UserAID         string     `json:"user_a_id"`
	UserBID         string     `json:"user_b_id"`
	Status          string     `json:"status"`
	RequestedBy     string     `json:"requested_by"`
	InvitationToken string     `json:"invitation_token,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// QRCodeResponse represents a QR code payload response (deprecated)
type QRCodeResponse struct {
	Payload   *QRCodePayload `json:"payload"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// InvitationResponse represents an invitation creation response
type InvitationResponse struct {
	ConnectionID   string    `json:"connection_id"`
	Token          string    `json:"token"`
	InvitationLink string    `json:"invitation_link"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// IsExpired checks if the invitation has expired
func (c *Connection) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsPending checks if the connection is in pending status
func (c *Connection) IsPending() bool {
	return c.Status == ConnectionStatusPending
}

// IsAccepted checks if the connection is accepted
func (c *Connection) IsAccepted() bool {
	return c.Status == ConnectionStatusAccepted
}

// IsRejected checks if the connection is rejected
func (c *Connection) IsRejected() bool {
	return c.Status == ConnectionStatusRejected
}

// IsBlocked checks if the connection is blocked
func (c *Connection) IsBlocked() bool {
	return c.Status == ConnectionStatusBlocked
}

// CanTransitionTo checks if the connection can transition to the given status
func (c *Connection) CanTransitionTo(newStatus ConnectionStatus) bool {
	switch c.Status {
	case ConnectionStatusPending:
		return newStatus == ConnectionStatusAccepted ||
			newStatus == ConnectionStatusRejected ||
			newStatus == ConnectionStatusBlocked
	case ConnectionStatusAccepted:
		return newStatus == ConnectionStatusBlocked
	case ConnectionStatusRejected:
		return newStatus == ConnectionStatusPending // Can resend invitation
	case ConnectionStatusBlocked:
		return false // Blocked connections cannot change status
	default:
		return false
	}
}

// GetOtherUserID returns the ID of the other user in the connection
func (c *Connection) GetOtherUserID(userID uuid.UUID) uuid.UUID {
	if c.UserAID == userID {
		return c.UserBID
	}
	return c.UserAID
}

// IsParticipant checks if the given user is a participant in this connection
func (c *Connection) IsParticipant(userID uuid.UUID) bool {
	return c.UserAID == userID || c.UserBID == userID
}

// GenerateInvitationToken generates a new UUID v4 token for invitations
func GenerateInvitationToken() string {
	return uuid.New().String()
}

// InvitationExpirationDuration is the duration after which an invitation expires
const InvitationExpirationDuration = 24 * time.Hour

// CalculateExpirationTime calculates the expiration time for a new invitation
func CalculateExpirationTime() time.Time {
	return time.Now().Add(InvitationExpirationDuration)
}

// InvitationLinkBaseURL is the base URL for invitation links
const InvitationLinkBaseURL = "https://app.todo.com/invite"

// GenerateInvitationLink generates a full invitation link from a token
func GenerateInvitationLink(token string) string {
	return InvitationLinkBaseURL + "/" + token
}

// NormalizeUserPair ensures consistent ordering of user IDs
// Returns (smallerUUID, largerUUID) for unique constraint
func NormalizeUserPair(userAID, userBID uuid.UUID) (uuid.UUID, uuid.UUID) {
	// Compare UUID strings for consistent ordering
	if userAID.String() < userBID.String() {
		return userAID, userBID
	}
	return userBID, userAID
}
