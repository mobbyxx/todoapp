package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

const qrCodeValidityWindow = 5 * time.Minute

type connectionService struct {
	repo             domain.ConnectionRepository
	userRepo         domain.UserRepository
	notificationRepo domain.NotificationQueueRepository
	gamificationSvc  domain.GamificationService
	secret           string
}

func NewConnectionService(
	repo domain.ConnectionRepository,
	userRepo domain.UserRepository,
	notificationRepo domain.NotificationQueueRepository,
	gamificationSvc domain.GamificationService,
	secret string,
) domain.ConnectionService {
	return &connectionService{
		repo:             repo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		gamificationSvc:  gamificationSvc,
		secret:           secret,
	}
}

func (s *connectionService) CreateInvitation(userID uuid.UUID) (*domain.Connection, string, error) {
	userAID, userBID := domain.NormalizeUserPair(userID, uuid.Nil)

	_, err := s.repo.GetByUserPair(userAID, userBID)
	if err == nil {
		return nil, "", domain.ErrConnectionAlreadyExists
	}
	if err != domain.ErrConnectionNotFound {
		return nil, "", fmt.Errorf("failed to check existing connection: %w", err)
	}

	token := domain.GenerateInvitationToken()
	expiresAt := domain.CalculateExpirationTime()

	connection := &domain.Connection{
		UserAID:         userAID,
		UserBID:         userBID,
		Status:          domain.ConnectionStatusPending,
		RequestedBy:     userID,
		InvitationToken: token,
		ExpiresAt:       &expiresAt,
	}

	if err := s.repo.Create(connection); err != nil {
		return nil, "", fmt.Errorf("failed to create connection: %w", err)
	}

	return connection, token, nil
}

func (s *connectionService) ValidateInvitation(token string) (*domain.Connection, error) {
	connection, err := s.repo.GetByToken(token)
	if err != nil {
		if err == domain.ErrConnectionNotFound {
			return nil, domain.ErrInvalidInvitationToken
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	if !connection.IsPending() {
		return nil, domain.ErrInvalidInvitationToken
	}

	if connection.IsExpired() {
		return nil, domain.ErrInvitationExpired
	}

	return connection, nil
}

func (s *connectionService) AcceptInvitation(userID uuid.UUID, token string) error {
	connection, err := s.ValidateInvitation(token)
	if err != nil {
		return err
	}

	if connection.RequestedBy == userID {
		return domain.ErrUnauthorizedAction
	}

	if !connection.CanTransitionTo(domain.ConnectionStatusAccepted) {
		return domain.ErrInvalidStatusTransition
	}

	now := time.Now()
	connection.Status = domain.ConnectionStatusAccepted
	connection.UserBID = userID
	connection.UserAID, connection.UserBID = domain.NormalizeUserPair(connection.UserAID, connection.UserBID)
	connection.AcceptedAt = &now

	if err := s.repo.Update(connection); err != nil {
		return fmt.Errorf("failed to accept connection: %w", err)
	}

	s.queueConnectionAcceptedNotification(connection, userID)

	if s.gamificationSvc != nil {
		s.gamificationSvc.OnConnectionAdded(connection.RequestedBy)
		s.gamificationSvc.OnConnectionAdded(userID)
	}

	return nil
}

func (s *connectionService) RejectInvitation(userID uuid.UUID, token string) error {
	connection, err := s.ValidateInvitation(token)
	if err != nil {
		return err
	}

	if connection.RequestedBy == userID {
		return domain.ErrUnauthorizedAction
	}

	if !connection.CanTransitionTo(domain.ConnectionStatusRejected) {
		return domain.ErrInvalidStatusTransition
	}

	now := time.Now()
	connection.Status = domain.ConnectionStatusRejected
	connection.RejectedAt = &now

	if err := s.repo.Update(connection); err != nil {
		return fmt.Errorf("failed to reject connection: %w", err)
	}

	return nil
}

func (s *connectionService) GetConnections(userID uuid.UUID) ([]*domain.Connection, error) {
	connections, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	var result []*domain.Connection
	for _, conn := range connections {
		if conn.IsAccepted() {
			result = append(result, conn)
		}
	}

	return result, nil
}

func (s *connectionService) Disconnect(connectionID uuid.UUID, userID uuid.UUID) error {
	connection, err := s.repo.GetByID(connectionID)
	if err != nil {
		if err == domain.ErrConnectionNotFound {
			return err
		}
		return fmt.Errorf("failed to get connection: %w", err)
	}

	if !connection.IsParticipant(userID) {
		return domain.ErrUnauthorizedAction
	}

	if err := s.repo.Delete(connectionID); err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	return nil
}

func (s *connectionService) queueConnectionAcceptedNotification(connection *domain.Connection, acceptedBy uuid.UUID) {
	requesterID := connection.RequestedBy

	payload := domain.ConnectionAcceptedPayload{
		ConnectionID: connection.ID.String(),
		UserID:       acceptedBy.String(),
		UserName:     "",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	item := &domain.NotificationQueueItem{
		UserID:   requesterID,
		Type:     domain.NotificationTypePush,
		Priority: 5,
		Payload:  payloadBytes,
		Status:   domain.NotificationStatusPending,
	}

	s.notificationRepo.Enqueue(item)
}

func (s *connectionService) GenerateQRCode(userID uuid.UUID) (*domain.QRCodePayload, error) {
	timestamp := time.Now().Unix()

	payload := &domain.QRCodePayload{
		UserID:    userID.String(),
		Timestamp: timestamp,
	}

	signature, err := s.signPayload(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign payload: %w", err)
	}

	payload.Signature = signature

	return payload, nil
}

func (s *connectionService) ScanQRCode(scannerID uuid.UUID, payload *domain.QRCodePayload) (*domain.Connection, error) {
	if err := s.validatePayload(payload); err != nil {
		return nil, err
	}

	targetID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return nil, domain.ErrInvalidQRCode
	}

	if scannerID == targetID {
		return nil, domain.ErrSelfConnection
	}

	existing, err := s.repo.GetByUserPair(scannerID, targetID)
	if err != nil && err != domain.ErrConnectionNotFound {
		return nil, fmt.Errorf("failed to check existing connection: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrConnectionAlreadyExists
	}

	now := time.Now()
	userAID, userBID := domain.NormalizeUserPair(scannerID, targetID)
	connection := &domain.Connection{
		UserAID:     userAID,
		UserBID:     userBID,
		Status:      domain.ConnectionStatusAccepted,
		RequestedBy: scannerID,
		AcceptedAt:  &now,
	}

	if err := s.repo.Create(connection); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	if s.gamificationSvc != nil {
		s.gamificationSvc.OnConnectionAdded(scannerID)
		s.gamificationSvc.OnConnectionAdded(targetID)
	}

	return connection, nil
}

func (s *connectionService) ListConnections(userID uuid.UUID) ([]domain.Connection, error) {
	connections, err := s.GetConnections(userID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Connection, len(connections))
	for i, conn := range connections {
		result[i] = *conn
	}

	return result, nil
}

func (s *connectionService) RemoveConnection(userID, connectionID uuid.UUID) error {
	return s.Disconnect(connectionID, userID)
}

func (s *connectionService) signPayload(payload *domain.QRCodePayload) (string, error) {
	data := fmt.Sprintf("%s:%d", payload.UserID, payload.Timestamp)

	h := hmac.New(sha256.New, []byte(s.secret))
	if _, err := h.Write([]byte(data)); err != nil {
		return "", err
	}

	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func (s *connectionService) validatePayload(payload *domain.QRCodePayload) error {
	if payload.UserID == "" || payload.Timestamp == 0 || payload.Signature == "" {
		return domain.ErrInvalidQRCode
	}

	payloadTime := time.Unix(payload.Timestamp, 0)
	if time.Since(payloadTime) > qrCodeValidityWindow {
		return domain.ErrQRCodeExpired
	}

	expectedSig, err := s.signPayload(&domain.QRCodePayload{
		UserID:    payload.UserID,
		Timestamp: payload.Timestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to compute signature: %w", err)
	}

	if !hmac.Equal([]byte(expectedSig), []byte(payload.Signature)) {
		return domain.ErrInvalidSignature
	}

	return nil
}
