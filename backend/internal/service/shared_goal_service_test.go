package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
)

type mockSharedGoalRepository struct {
	goals map[uuid.UUID]*domain.SharedGoal
}

func newMockSharedGoalRepository() *mockSharedGoalRepository {
	return &mockSharedGoalRepository{
		goals: make(map[uuid.UUID]*domain.SharedGoal),
	}
}

func (m *mockSharedGoalRepository) Create(goal *domain.SharedGoal) error {
	goal.ID = uuid.New()
	m.goals[goal.ID] = goal
	return nil
}

func (m *mockSharedGoalRepository) GetByID(id uuid.UUID) (*domain.SharedGoal, error) {
	if goal, ok := m.goals[id]; ok {
		return goal, nil
	}
	return nil, domain.ErrSharedGoalNotFound
}

func (m *mockSharedGoalRepository) GetByConnection(connectionID uuid.UUID) ([]*domain.SharedGoal, error) {
	var result []*domain.SharedGoal
	for _, goal := range m.goals {
		if goal.ConnectionID == connectionID {
			result = append(result, goal)
		}
	}
	return result, nil
}

func (m *mockSharedGoalRepository) GetByUserID(userID uuid.UUID) ([]*domain.SharedGoal, error) {
	var result []*domain.SharedGoal
	for _, goal := range m.goals {
		result = append(result, goal)
	}
	return result, nil
}

func (m *mockSharedGoalRepository) UpdateProgress(id uuid.UUID, amount int) error {
	goal, ok := m.goals[id]
	if !ok {
		return domain.ErrSharedGoalNotFound
	}
	goal.CurrentValue += amount
	if goal.CurrentValue >= goal.TargetValue {
		goal.Status = domain.SharedGoalStatusCompleted
	}
	return nil
}

func (m *mockSharedGoalRepository) Update(goal *domain.SharedGoal) error {
	m.goals[goal.ID] = goal
	return nil
}

type mockConnectionService struct {
	connections map[uuid.UUID]*domain.Connection
}

func newMockConnectionService() *mockConnectionService {
	return &mockConnectionService{
		connections: make(map[uuid.UUID]*domain.Connection),
	}
}

func (m *mockConnectionService) CreateInvitation(userID uuid.UUID) (*domain.Connection, string, error) {
	return nil, "", nil
}

func (m *mockConnectionService) ValidateInvitation(token string) (*domain.Connection, error) {
	return nil, nil
}

func (m *mockConnectionService) AcceptInvitation(userID uuid.UUID, token string) error {
	return nil
}

func (m *mockConnectionService) RejectInvitation(userID uuid.UUID, token string) error {
	return nil
}

func (m *mockConnectionService) GetConnections(userID uuid.UUID) ([]*domain.Connection, error) {
	var result []*domain.Connection
	for _, conn := range m.connections {
		if conn.UserAID == userID || conn.UserBID == userID {
			result = append(result, conn)
		}
	}
	return result, nil
}

func (m *mockConnectionService) Disconnect(connectionID uuid.UUID, userID uuid.UUID) error {
	return nil
}

func (m *mockConnectionService) GenerateQRCode(userID uuid.UUID) (*domain.QRCodePayload, error) {
	return nil, nil
}

func (m *mockConnectionService) ScanQRCode(scannerID uuid.UUID, payload *domain.QRCodePayload) (*domain.Connection, error) {
	return nil, nil
}

func (m *mockConnectionService) ListConnections(userID uuid.UUID) ([]domain.Connection, error) {
	var result []domain.Connection
	for _, conn := range m.connections {
		result = append(result, *conn)
	}
	return result, nil
}

func (m *mockConnectionService) RemoveConnection(userID, connectionID uuid.UUID) error {
	return nil
}

type mockConnRepoForGoal struct {
	connections map[uuid.UUID]*domain.Connection
}

func newMockConnRepoForGoal() *mockConnRepoForGoal {
	return &mockConnRepoForGoal{
		connections: make(map[uuid.UUID]*domain.Connection),
	}
}

func (m *mockConnRepoForGoal) Create(connection *domain.Connection) error {
	m.connections[connection.ID] = connection
	return nil
}

func (m *mockConnRepoForGoal) GetByID(id uuid.UUID) (*domain.Connection, error) {
	if conn, ok := m.connections[id]; ok {
		return conn, nil
	}
	return nil, domain.ErrConnectionNotFound
}

func (m *mockConnRepoForGoal) GetByToken(token string) (*domain.Connection, error) {
	return nil, nil
}

func (m *mockConnRepoForGoal) GetByUserID(userID uuid.UUID) ([]*domain.Connection, error) {
	var result []*domain.Connection
	for _, conn := range m.connections {
		if conn.UserAID == userID || conn.UserBID == userID {
			result = append(result, conn)
		}
	}
	return result, nil
}

func (m *mockConnRepoForGoal) GetByUserPair(userAID, userBID uuid.UUID) (*domain.Connection, error) {
	return nil, nil
}

func (m *mockConnRepoForGoal) Update(connection *domain.Connection) error {
	m.connections[connection.ID] = connection
	return nil
}

func (m *mockConnRepoForGoal) Delete(id uuid.UUID) error {
	delete(m.connections, id)
	return nil
}

type mockGamificationService struct {
	xpAwards map[uuid.UUID]int
}

func newMockGamificationService() *mockGamificationService {
	return &mockGamificationService{
		xpAwards: make(map[uuid.UUID]int),
	}
}

func (m *mockGamificationService) CheckAndAwardBadges(userID uuid.UUID) ([]*domain.Badge, error) {
	return nil, nil
}

func (m *mockGamificationService) EvaluateBadgeCriteria(userID uuid.UUID, badge *domain.Badge) (bool, error) {
	return false, nil
}

func (m *mockGamificationService) GetUserBadges(userID uuid.UUID) ([]*domain.BadgeWithEarned, error) {
	return nil, nil
}

func (m *mockGamificationService) AwardBadge(userID uuid.UUID, badgeID uuid.UUID) (*domain.UserBadge, error) {
	return nil, nil
}

func (m *mockGamificationService) AwardXP(userID uuid.UUID, amount int, reason string) error {
	m.xpAwards[userID] += amount
	return nil
}

func (m *mockGamificationService) GetUserStats(userID uuid.UUID) (*domain.UserStats, error) {
	return nil, nil
}

func (m *mockGamificationService) OnTodoCompleted(userID uuid.UUID, completedAt time.Time) {
}

func (m *mockGamificationService) OnStreakUpdated(userID uuid.UUID, streakDays int) {
}

func (m *mockGamificationService) OnConnectionAdded(userID uuid.UUID) {
}

func (m *mockGamificationService) GetPointsHistory(userID uuid.UUID, limit int) ([]*domain.PointsTransaction, error) {
	return nil, nil
}

type mockNotificationServiceForGoal struct{}

func (m *mockNotificationServiceForGoal) QueueNotification(userID uuid.UUID, notificationType domain.NotificationType, title string, body string, data map[string]interface{}, priority int) error {
	return nil
}
func (m *mockNotificationServiceForGoal) QueueConnectionRequest(userID uuid.UUID, fromUserID uuid.UUID, fromUserName string, connectionID uuid.UUID) error {
	return nil
}
func (m *mockNotificationServiceForGoal) QueueConnectionAccepted(userID uuid.UUID, acceptedByID uuid.UUID, acceptedByName string, connectionID uuid.UUID) error {
	return nil
}
func (m *mockNotificationServiceForGoal) QueueTodoAssigned(userID uuid.UUID, todoID uuid.UUID, title string, assignerID uuid.UUID, assignerName string) error {
	return nil
}
func (m *mockNotificationServiceForGoal) QueueTodoCompleted(userID uuid.UUID, todoID uuid.UUID, title string, completedByID uuid.UUID, completedByName string) error {
	return nil
}
func (m *mockNotificationServiceForGoal) QueueBadgeEarned(userID uuid.UUID, badge *domain.Badge) error {
	return nil
}
func (m *mockNotificationServiceForGoal) QueueGoalCompleted(userID uuid.UUID, goalID uuid.UUID, goalName string, reward string) error {
	return nil
}
func (m *mockNotificationServiceForGoal) ProcessQueue(batchSize int) error {
	return nil
}
func (m *mockNotificationServiceForGoal) RetryFailed() error {
	return nil
}

func setupSharedGoalTestService(t *testing.T) (*sharedGoalService, *mockSharedGoalRepository, *mockConnRepoForGoal, *mockConnectionService, *mockGamificationService) {
	goalRepo := newMockSharedGoalRepository()
	connRepo := newMockConnRepoForGoal()
	connSvc := newMockConnectionService()
	gamificationSvc := newMockGamificationService()
	notifSvc := &mockNotificationServiceForGoal{}
	service := NewSharedGoalService(goalRepo, connRepo, connSvc, gamificationSvc, notifSvc).(*sharedGoalService)
	return service, goalRepo, connRepo, connSvc, gamificationSvc
}

func TestCreateGoal_Success(t *testing.T) {
	service, _, _, _, _ := setupSharedGoalTestService(t)
	connectionID := uuid.New()

	goal, err := service.CreateGoal(
		connectionID,
		domain.SharedGoalTargetTypeTodosCompleted,
		10,
		"Complete 10 todos together",
	)

	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}

	if goal == nil {
		t.Fatal("Expected goal, got nil")
	}

	if goal.ConnectionID != connectionID {
		t.Errorf("Expected connection ID %s, got %s", connectionID.String(), goal.ConnectionID.String())
	}

	if goal.TargetType != domain.SharedGoalTargetTypeTodosCompleted {
		t.Errorf("Expected target type todos_completed, got %s", goal.TargetType)
	}

	if goal.TargetValue != 10 {
		t.Errorf("Expected target value 10, got %d", goal.TargetValue)
	}

	if goal.CurrentValue != 0 {
		t.Errorf("Expected current value 0, got %d", goal.CurrentValue)
	}

	if goal.Status != domain.SharedGoalStatusActive {
		t.Errorf("Expected status active, got %s", goal.Status)
	}
}

func TestCreateGoal_InvalidTargetType(t *testing.T) {
	service, _, _, _, _ := setupSharedGoalTestService(t)
	connectionID := uuid.New()

	_, err := service.CreateGoal(
		connectionID,
		domain.SharedGoalTargetType("invalid_type"),
		10,
		"Test goal",
	)

	if !errors.Is(err, domain.ErrInvalidTargetType) {
		t.Errorf("Expected ErrInvalidTargetType, got %v", err)
	}
}

func TestCreateGoal_InvalidTargetValue(t *testing.T) {
	service, _, _, _, _ := setupSharedGoalTestService(t)
	connectionID := uuid.New()

	testCases := []int{0, -1, -10}
	for _, value := range testCases {
		_, err := service.CreateGoal(
			connectionID,
			domain.SharedGoalTargetTypeTodosCompleted,
			value,
			"Test goal",
		)

		if !errors.Is(err, domain.ErrInvalidTargetValue) {
			t.Errorf("Expected ErrInvalidTargetValue for value %d, got %v", value, err)
		}
	}
}

func TestCreateGoal_StreakDaysType(t *testing.T) {
	service, _, _, _, _ := setupSharedGoalTestService(t)
	connectionID := uuid.New()

	goal, err := service.CreateGoal(
		connectionID,
		domain.SharedGoalTargetTypeStreakDays,
		7,
		"Maintain 7-day streak together",
	)

	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}

	if goal.TargetType != domain.SharedGoalTargetTypeStreakDays {
		t.Errorf("Expected target type streak_days, got %s", goal.TargetType)
	}

	if goal.TargetValue != 7 {
		t.Errorf("Expected target value 7, got %d", goal.TargetValue)
	}
}

func TestUpdateProgress(t *testing.T) {
	service, goalRepo, connRepo, _, _ := setupSharedGoalTestService(t)
	connectionID := uuid.New()

	connRepo.connections[connectionID] = &domain.Connection{
		ID:      connectionID,
		UserAID: uuid.New(),
		UserBID: uuid.New(),
		Status:  domain.ConnectionStatusAccepted,
	}

	goal := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  10,
		CurrentValue: 5,
		Status:       domain.SharedGoalStatusActive,
	}
	goalRepo.goals[goal.ID] = goal

	err := service.UpdateProgress(connectionID, 3)
	if err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	if goal.CurrentValue != 8 {
		t.Errorf("Expected current value 8, got %d", goal.CurrentValue)
	}
}

func TestUpdateProgress_CompletesGoal(t *testing.T) {
	service, goalRepo, connRepo, _, gamificationSvc := setupSharedGoalTestService(t)
	connectionID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	connRepo.connections[connectionID] = &domain.Connection{
		ID:      connectionID,
		UserAID: userA,
		UserBID: userB,
		Status:  domain.ConnectionStatusAccepted,
	}

	goal := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  10,
		CurrentValue: 8,
		Status:       domain.SharedGoalStatusActive,
	}
	goalRepo.goals[goal.ID] = goal

	err := service.UpdateProgress(connectionID, 5)
	if err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	if goal.Status != domain.SharedGoalStatusCompleted {
		t.Errorf("Expected status completed, got %s", goal.Status)
	}

	if gamificationSvc.xpAwards[userA] != domain.SharedGoalCompletionBonusXP {
		t.Errorf("Expected user A to receive %d XP, got %d", domain.SharedGoalCompletionBonusXP, gamificationSvc.xpAwards[userA])
	}

	if gamificationSvc.xpAwards[userB] != domain.SharedGoalCompletionBonusXP {
		t.Errorf("Expected user B to receive %d XP, got %d", domain.SharedGoalCompletionBonusXP, gamificationSvc.xpAwards[userB])
	}
}

func TestCheckCompletion_NotCompleted(t *testing.T) {
	service, goalRepo, _, _, _ := setupSharedGoalTestService(t)
	goalID := uuid.New()

	goal := &domain.SharedGoal{
		ID:           goalID,
		ConnectionID: uuid.New(),
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  10,
		CurrentValue: 5,
		Status:       domain.SharedGoalStatusActive,
	}
	goalRepo.goals[goalID] = goal

	result, err := service.CheckCompletion(goalID)
	if err != nil {
		t.Fatalf("CheckCompletion failed: %v", err)
	}

	if result.Status != domain.SharedGoalStatusActive {
		t.Errorf("Expected status active, got %s", result.Status)
	}
}

func TestCheckCompletion_AlreadyCompleted(t *testing.T) {
	service, goalRepo, _, _, _ := setupSharedGoalTestService(t)
	goalID := uuid.New()

	goal := &domain.SharedGoal{
		ID:           goalID,
		ConnectionID: uuid.New(),
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  10,
		CurrentValue: 10,
		Status:       domain.SharedGoalStatusCompleted,
	}
	goalRepo.goals[goalID] = goal

	result, err := service.CheckCompletion(goalID)
	if err != nil {
		t.Fatalf("CheckCompletion failed: %v", err)
	}

	if result.Status != domain.SharedGoalStatusCompleted {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestListGoals(t *testing.T) {
	service, goalRepo, _, _, _ := setupSharedGoalTestService(t)
	userID := uuid.New()
	connectionID := uuid.New()

	goal1 := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  10,
		Status:       domain.SharedGoalStatusActive,
	}
	goal2 := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		TargetType:   domain.SharedGoalTargetTypeStreakDays,
		TargetValue:  7,
		Status:       domain.SharedGoalStatusCompleted,
	}
	goalRepo.goals[goal1.ID] = goal1
	goalRepo.goals[goal2.ID] = goal2

	goals, err := service.ListGoals(userID)
	if err != nil {
		t.Fatalf("ListGoals failed: %v", err)
	}

	if len(goals) != 2 {
		t.Errorf("Expected 2 goals, got %d", len(goals))
	}
}

func TestOnTodoCompleted_SharedGoal(t *testing.T) {
	service, goalRepo, _, connSvc, _ := setupSharedGoalTestService(t)
	userID := uuid.New()
	connectionID := uuid.New()

	connSvc.connections[connectionID] = &domain.Connection{
		ID:      connectionID,
		UserAID: userID,
		UserBID: uuid.New(),
		Status:  domain.ConnectionStatusAccepted,
	}

	goal := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  10,
		CurrentValue: 0,
		Status:       domain.SharedGoalStatusActive,
	}
	goalRepo.goals[goal.ID] = goal

	service.OnTodoCompleted(userID)

	if goal.CurrentValue != 1 {
		t.Errorf("Expected current value 1, got %d", goal.CurrentValue)
	}
}

func TestOnTodoCompleted_StreakGoalNotAffected(t *testing.T) {
	service, goalRepo, _, connSvc, _ := setupSharedGoalTestService(t)
	userID := uuid.New()
	connectionID := uuid.New()

	connSvc.connections[connectionID] = &domain.Connection{
		ID:      connectionID,
		UserAID: userID,
		UserBID: uuid.New(),
		Status:  domain.ConnectionStatusAccepted,
	}

	goal := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		TargetType:   domain.SharedGoalTargetTypeStreakDays,
		TargetValue:  7,
		CurrentValue: 0,
		Status:       domain.SharedGoalStatusActive,
	}
	goalRepo.goals[goal.ID] = goal

	service.OnTodoCompleted(userID)

	if goal.CurrentValue != 0 {
		t.Errorf("Expected streak goal current value to remain 0, got %d", goal.CurrentValue)
	}
}

func TestOnTodoCompleted_NoConnections(t *testing.T) {
	service, _, _, _, _ := setupSharedGoalTestService(t)
	userID := uuid.New()

	service.OnTodoCompleted(userID)
}

func TestOnTodoCompleted_MultipleConnections(t *testing.T) {
	service, goalRepo, _, connSvc, _ := setupSharedGoalTestService(t)
	userID := uuid.New()
	connectionID1 := uuid.New()
	connectionID2 := uuid.New()

	connSvc.connections[connectionID1] = &domain.Connection{
		ID:      connectionID1,
		UserAID: userID,
		UserBID: uuid.New(),
		Status:  domain.ConnectionStatusAccepted,
	}
	connSvc.connections[connectionID2] = &domain.Connection{
		ID:      connectionID2,
		UserAID: uuid.New(),
		UserBID: userID,
		Status:  domain.ConnectionStatusAccepted,
	}

	goal1 := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID1,
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  5,
		CurrentValue: 0,
		Status:       domain.SharedGoalStatusActive,
	}
	goal2 := &domain.SharedGoal{
		ID:           uuid.New(),
		ConnectionID: connectionID2,
		TargetType:   domain.SharedGoalTargetTypeTodosCompleted,
		TargetValue:  5,
		CurrentValue: 0,
		Status:       domain.SharedGoalStatusActive,
	}
	goalRepo.goals[goal1.ID] = goal1
	goalRepo.goals[goal2.ID] = goal2

	service.OnTodoCompleted(userID)

	if goal1.CurrentValue != 1 {
		t.Errorf("Expected goal1 current value 1, got %d", goal1.CurrentValue)
	}

	if goal2.CurrentValue != 1 {
		t.Errorf("Expected goal2 current value 1, got %d", goal2.CurrentValue)
	}
}

func TestSharedGoal_GetProgressPercentage(t *testing.T) {
	testCases := []struct {
		name         string
		currentValue int
		targetValue  int
		expected     float64
	}{
		{"0% progress", 0, 10, 0.0},
		{"50% progress", 5, 10, 50.0},
		{"100% progress", 10, 10, 100.0},
		{"over 100% capped", 15, 10, 100.0},
		{"zero target", 5, 0, 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goal := &domain.SharedGoal{
				CurrentValue: tc.currentValue,
				TargetValue:  tc.targetValue,
			}

			progress := goal.GetProgressPercentage()
			if progress != tc.expected {
				t.Errorf("Expected progress %f, got %f", tc.expected, progress)
			}
		})
	}
}

func TestSharedGoal_CanBeUpdated(t *testing.T) {
	testCases := []struct {
		name      string
		status    domain.SharedGoalStatus
		expectErr error
	}{
		{"active goal", domain.SharedGoalStatusActive, nil},
		{"completed goal", domain.SharedGoalStatusCompleted, domain.ErrGoalAlreadyCompleted},
		{"cancelled goal", domain.SharedGoalStatusCancelled, domain.ErrGoalCancelled},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goal := &domain.SharedGoal{
				Status: tc.status,
			}

			err := goal.CanBeUpdated()
			if err != tc.expectErr {
				t.Errorf("Expected error %v, got %v", tc.expectErr, err)
			}
		})
	}
}

func TestSharedGoal_MarkAsCompleted(t *testing.T) {
	goal := &domain.SharedGoal{
		ID:           uuid.New(),
		TargetValue:  10,
		CurrentValue: 8,
		Status:       domain.SharedGoalStatusActive,
	}

	goal.MarkAsCompleted()

	if goal.Status != domain.SharedGoalStatusCompleted {
		t.Errorf("Expected status completed, got %s", goal.Status)
	}

	if goal.CurrentValue != 10 {
		t.Errorf("Expected current value to be set to target value 10, got %d", goal.CurrentValue)
	}

	if goal.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}
