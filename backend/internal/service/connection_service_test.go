package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

type mockConnectionRepository struct {
	connections map[uuid.UUID]*domain.Connection
	byUsers     map[string]*domain.Connection
}

func newMockConnectionRepository() *mockConnectionRepository {
	return &mockConnectionRepository{
		connections: make(map[uuid.UUID]*domain.Connection),
		byUsers:     make(map[string]*domain.Connection),
	}
}

func (m *mockConnectionRepository) Create(connection *domain.Connection) error {
	m.connections[connection.ID] = connection
	key := fmt.Sprintf("%s:%s", connection.UserID, connection.FriendID)
	m.byUsers[key] = connection
	return nil
}

func (m *mockConnectionRepository) GetByID(id uuid.UUID) (*domain.Connection, error) {
	if conn, ok := m.connections[id]; ok {
		return conn, nil
	}
	return nil, domain.ErrConnectionNotFound
}

func (m *mockConnectionRepository) GetByUsers(userID, friendID uuid.UUID) (*domain.Connection, error) {
	key := fmt.Sprintf("%s:%s", userID, friendID)
	if conn, ok := m.byUsers[key]; ok {
		return conn, nil
	}
	return nil, domain.ErrConnectionNotFound
}

func (m *mockConnectionRepository) ListByUser(userID uuid.UUID) ([]domain.Connection, error) {
	var result []domain.Connection
	for _, conn := range m.connections {
		if conn.UserID == userID {
			result = append(result, *conn)
		}
	}
	return result, nil
}

func (m *mockConnectionRepository) Update(connection *domain.Connection) error {
	m.connections[connection.ID] = connection
	return nil
}

func (m *mockConnectionRepository) Delete(id uuid.UUID) error {
	if _, ok := m.connections[id]; !ok {
		return domain.ErrConnectionNotFound
	}
	delete(m.connections, id)
	return nil
}

func setupConnectionTestService(t *testing.T) (*ConnectionService, *mockConnectionRepository) {
	repo := newMockConnectionRepository()
	secret := "test-secret-key-for-qr-signing"
	service := NewConnectionService(repo, secret)
	return service, repo
}

func TestGenerateQRCode(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userID := uuid.New()

	payload, err := service.GenerateQRCode(userID)
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}

	if payload == nil {
		t.Fatal("Expected payload, got nil")
	}

	if payload.UserID != userID.String() {
		t.Errorf("Expected user ID %s, got %s", userID.String(), payload.UserID)
	}

	if payload.Timestamp == 0 {
		t.Error("Expected timestamp to be set")
	}

	if payload.Signature == "" {
		t.Error("Expected signature to be set")
	}

	now := time.Now().Unix()
	if payload.Timestamp > now || payload.Timestamp < now-5 {
		t.Errorf("Timestamp %d is not within expected range", payload.Timestamp)
	}
}

func TestGenerateQRCode_SignatureValid(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userID := uuid.New()

	payload, err := service.GenerateQRCode(userID)
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}

	err = service.VerifySignature(payload)
	if err != nil {
		t.Errorf("Expected signature to be valid, got error: %v", err)
	}
}

func TestScanQRCode_ValidPayload(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	payload, err := service.GenerateQRCode(userA)
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}

	connection, err := service.ScanQRCode(userB, payload)
	if err != nil {
		t.Fatalf("ScanQRCode failed: %v", err)
	}

	if connection == nil {
		t.Fatal("Expected connection, got nil")
	}

	if connection.UserID != userB {
		t.Errorf("Expected connection user ID %s, got %s", userB.String(), connection.UserID.String())
	}

	if connection.FriendID != userA {
		t.Errorf("Expected connection friend ID %s, got %s", userA.String(), connection.FriendID.String())
	}

	if connection.Status != "accepted" {
		t.Errorf("Expected status 'accepted', got %s", connection.Status)
	}

	if connection.AcceptedAt == nil {
		t.Error("Expected AcceptedAt to be set")
	}
}

func TestScanQRCode_CreatesBidirectionalConnection(t *testing.T) {
	service, repo := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	payload, _ := service.GenerateQRCode(userA)
	service.ScanQRCode(userB, payload)

	forward, err := repo.GetByUsers(userB, userA)
	if err != nil {
		t.Fatalf("Failed to get forward connection: %v", err)
	}
	if forward == nil {
		t.Error("Expected forward connection to exist")
	}

	reverse, err := repo.GetByUsers(userA, userB)
	if err != nil {
		t.Fatalf("Failed to get reverse connection: %v", err)
	}
	if reverse == nil {
		t.Error("Expected reverse connection to exist")
	}
}

func TestScanQRCode_ExpiredQRCode(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	expiredPayload := &domain.QRCodePayload{
		UserID:    userA.String(),
		Timestamp: time.Now().Add(-10 * time.Minute).Unix(),
	}

	sig, _ := service.signPayload(expiredPayload)
	expiredPayload.Signature = sig

	_, err := service.ScanQRCode(userB, expiredPayload)
	if !errors.Is(err, domain.ErrQRCodeExpired) {
		t.Errorf("Expected ErrQRCodeExpired, got %v", err)
	}
}

func TestScanQRCode_InvalidSignature(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	payload := &domain.QRCodePayload{
		UserID:    userA.String(),
		Timestamp: time.Now().Unix(),
		Signature: "invalid-signature",
	}

	_, err := service.ScanQRCode(userB, payload)
	if !errors.Is(err, domain.ErrInvalidSignature) {
		t.Errorf("Expected ErrInvalidSignature, got %v", err)
	}
}

func TestScanQRCode_TamperedPayload(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	payload, _ := service.GenerateQRCode(userA)
	payload.UserID = uuid.New().String()

	_, err := service.ScanQRCode(userB, payload)
	if !errors.Is(err, domain.ErrInvalidSignature) {
		t.Errorf("Expected ErrInvalidSignature for tampered payload, got %v", err)
	}
}

func TestScanQRCode_SelfConnection(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()

	payload, _ := service.GenerateQRCode(userA)

	_, err := service.ScanQRCode(userA, payload)
	if !errors.Is(err, domain.ErrSelfConnection) {
		t.Errorf("Expected ErrSelfConnection, got %v", err)
	}
}

func TestScanQRCode_DuplicateConnection(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	payload, _ := service.GenerateQRCode(userA)
	service.ScanQRCode(userB, payload)

	_, err := service.ScanQRCode(userB, payload)
	if !errors.Is(err, domain.ErrConnectionAlreadyExists) {
		t.Errorf("Expected ErrConnectionAlreadyExists, got %v", err)
	}
}

func TestScanQRCode_InvalidUserID(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userB := uuid.New()

	payload := &domain.QRCodePayload{
		UserID:    "invalid-uuid",
		Timestamp: time.Now().Unix(),
		Signature: "some-signature",
	}

	_, err := service.ScanQRCode(userB, payload)
	if !errors.Is(err, domain.ErrInvalidQRCode) {
		t.Errorf("Expected ErrInvalidQRCode, got %v", err)
	}
}

func TestScanQRCode_EmptyPayload(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userB := uuid.New()

	testCases := []struct {
		name      string
		payload   *domain.QRCodePayload
		wantError error
	}{
		{
			name:      "empty user_id",
			payload:   &domain.QRCodePayload{UserID: "", Timestamp: time.Now().Unix(), Signature: "sig"},
			wantError: domain.ErrInvalidQRCode,
		},
		{
			name:      "zero timestamp",
			payload:   &domain.QRCodePayload{UserID: uuid.New().String(), Timestamp: 0, Signature: "sig"},
			wantError: domain.ErrInvalidQRCode,
		},
		{
			name:      "empty signature",
			payload:   &domain.QRCodePayload{UserID: uuid.New().String(), Timestamp: time.Now().Unix(), Signature: ""},
			wantError: domain.ErrInvalidQRCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ScanQRCode(userB, tc.payload)
			if !errors.Is(err, tc.wantError) {
				t.Errorf("Expected %v, got %v", tc.wantError, err)
			}
		})
	}
}

func TestSignPayload(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userID := uuid.New()
	timestamp := time.Now().Unix()

	payload := &domain.QRCodePayload{
		UserID:    userID.String(),
		Timestamp: timestamp,
	}

	signature, err := service.signPayload(payload)
	if err != nil {
		t.Fatalf("signPayload failed: %v", err)
	}

	if signature == "" {
		t.Error("Expected signature to be non-empty")
	}

	expectedData := fmt.Sprintf("%s:%d", userID.String(), timestamp)
	h := hmac.New(sha256.New, []byte(service.secret))
	h.Write([]byte(expectedData))
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if signature != expectedSig {
		t.Error("Signature does not match expected HMAC-SHA256")
	}
}

func TestListConnections(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()
	userC := uuid.New()

	payload, _ := service.GenerateQRCode(userB)
	service.ScanQRCode(userA, payload)

	payload2, _ := service.GenerateQRCode(userC)
	service.ScanQRCode(userA, payload2)

	connections, err := service.ListConnections(userA)
	if err != nil {
		t.Fatalf("ListConnections failed: %v", err)
	}

	if len(connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(connections))
	}
}

func TestRemoveConnection(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	payload, _ := service.GenerateQRCode(userB)
	conn, _ := service.ScanQRCode(userA, payload)

	err := service.RemoveConnection(userA, conn.ID)
	if err != nil {
		t.Fatalf("RemoveConnection failed: %v", err)
	}

	connections, _ := service.ListConnections(userA)
	if len(connections) != 0 {
		t.Errorf("Expected 0 connections after removal, got %d", len(connections))
	}
}

func TestRemoveConnection_NotFound(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()

	err := service.RemoveConnection(userA, uuid.New())
	if !errors.Is(err, domain.ErrConnectionNotFound) {
		t.Errorf("Expected ErrConnectionNotFound, got %v", err)
	}
}

func TestRemoveConnection_Forbidden(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()
	userC := uuid.New()

	payload, _ := service.GenerateQRCode(userB)
	conn, _ := service.ScanQRCode(userA, payload)

	err := service.RemoveConnection(userC, conn.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestQRCode_ValidityWindow(t *testing.T) {
	service, _ := setupConnectionTestService(t)
	userA := uuid.New()
	userB := uuid.New()

	justExpiredPayload := &domain.QRCodePayload{
		UserID:    userA.String(),
		Timestamp: time.Now().Add(-5*time.Minute - 1*time.Second).Unix(),
	}
	sig, _ := service.signPayload(justExpiredPayload)
	justExpiredPayload.Signature = sig

	_, err := service.ScanQRCode(userB, justExpiredPayload)
	if !errors.Is(err, domain.ErrQRCodeExpired) {
		t.Errorf("Expected ErrQRCodeExpired for just-expired code, got %v", err)
	}

	stillValidPayload := &domain.QRCodePayload{
		UserID:    userA.String(),
		Timestamp: time.Now().Add(-4 * time.Minute).Unix(),
	}
	sig, _ = service.signPayload(stillValidPayload)
	stillValidPayload.Signature = sig

	err = service.VerifySignature(stillValidPayload)
	if err != nil {
		t.Errorf("Expected code to still be valid, got error: %v", err)
	}
}
