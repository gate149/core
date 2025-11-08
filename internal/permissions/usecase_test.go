package permissions

import (
	"context"
	"testing"

	"github.com/gate149/core/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations

type MockPermissionsRepo struct {
	mock.Mock
}

func (m *MockPermissionsRepo) CreatePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error {
	args := m.Called(ctx, resourceType, resourceID, userID, relation)
	return args.Error(0)
}

func (m *MockPermissionsRepo) DeletePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error {
	args := m.Called(ctx, resourceType, resourceID, userID, relation)
	return args.Error(0)
}

func (m *MockPermissionsRepo) HasPermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) (bool, error) {
	args := m.Called(ctx, resourceType, resourceID, userID, relation)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionsRepo) GetUserPermissions(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, resourceType, resourceID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPermissionsRepo) HasAnyRelation(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relations []string) (bool, error) {
	args := m.Called(ctx, resourceType, resourceID, userID, relations)
	return args.Bool(0), args.Error(1)
}

type MockUsersRepo struct {
	mock.Mock
}

func (m *MockUsersRepo) GetUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

type MockContestsReader struct {
	mock.Mock
}

func (m *MockContestsReader) GetContest(ctx context.Context, id uuid.UUID) (*models.Contest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contest), args.Error(1)
}

// Tests

func TestUseCase_CreatePermission(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	resourceID := uuid.New()
	userID := uuid.New()

	mockPermissionsRepo.On("CreatePermission", ctx, ResourceContest, resourceID, userID, RelationOwner).Return(nil)

	err := uc.CreatePermission(ctx, ResourceContest, resourceID, userID, RelationOwner)
	assert.NoError(t, err)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_DeletePermission(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	resourceID := uuid.New()
	userID := uuid.New()

	mockPermissionsRepo.On("DeletePermission", ctx, ResourceContest, resourceID, userID, RelationOwner).Return(nil)

	err := uc.DeletePermission(ctx, ResourceContest, resourceID, userID, RelationOwner)
	assert.NoError(t, err)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanViewContest_GlobalAdmin(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	contest := &models.Contest{
		Id:        contestID,
		IsPrivate: true,
	}

	adminUser := &models.User{
		Id:   userID,
		Role: "admin",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(adminUser, nil)

	canView, err := uc.CanViewContest(ctx, userID, contest)
	assert.NoError(t, err)
	assert.True(t, canView)
	mockUsersRepo.AssertExpectations(t)
}

func TestUseCase_CanViewContest_Owner(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	contest := &models.Contest{
		Id:        contestID,
		IsPrivate: true,
	}

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contestID, userID, RelationOwner).Return(true, nil)

	canView, err := uc.CanViewContest(ctx, userID, contest)
	assert.NoError(t, err)
	assert.True(t, canView)
	mockUsersRepo.AssertExpectations(t)
}

func TestUseCase_CanViewContest_PublicContest(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	contest := &models.Contest{
		Id:        contestID,
		IsPrivate: false,
	}

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contestID, userID, RelationOwner).Return(false, nil)

	canView, err := uc.CanViewContest(ctx, userID, contest)
	assert.NoError(t, err)
	assert.True(t, canView)
	mockUsersRepo.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanViewContest_PrivateNoPermission(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	contest := &models.Contest{
		Id:        contestID,
		IsPrivate: true,
	}

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contestID, userID, RelationOwner).Return(false, nil)
	mockPermissionsRepo.On("HasAnyRelation", ctx, ResourceContest, contestID, userID, []string{RelationModerator, RelationParticipant}).Return(false, nil)

	canView, err := uc.CanViewContest(ctx, userID, contest)
	assert.NoError(t, err)
	assert.False(t, canView)
	mockUsersRepo.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanEditContest_Owner(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contestID, userID, RelationOwner).Return(true, nil)

	canEdit, err := uc.CanEditContest(ctx, userID, contestID)
	assert.NoError(t, err)
	assert.True(t, canEdit)
	mockUsersRepo.AssertExpectations(t)
	mockContestsReader.AssertExpectations(t)
}

func TestUseCase_CanEditContest_Moderator(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contestID, userID, RelationOwner).Return(false, nil)
	mockPermissionsRepo.On("HasAnyRelation", ctx, ResourceContest, contestID, userID, []string{RelationModerator}).Return(true, nil)

	canEdit, err := uc.CanEditContest(ctx, userID, contestID)
	assert.NoError(t, err)
	assert.True(t, canEdit)
	mockUsersRepo.AssertExpectations(t)
	mockContestsReader.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanAdminContest_Owner(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contestID, userID, RelationOwner).Return(true, nil)

	canAdmin, err := uc.CanAdminContest(ctx, userID, contestID)
	assert.NoError(t, err)
	assert.True(t, canAdmin)
	mockUsersRepo.AssertExpectations(t)
	mockContestsReader.AssertExpectations(t)
}

func TestUseCase_CanAdminContest_NotOwner(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contestID := uuid.New()
	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contestID, userID, RelationOwner).Return(false, nil)

	canAdmin, err := uc.CanAdminContest(ctx, userID, contestID)
	assert.NoError(t, err)
	assert.False(t, canAdmin)
	mockUsersRepo.AssertExpectations(t)
	mockContestsReader.AssertExpectations(t)
}

func TestUseCase_CanViewProblem_PublicProblem(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	problem := &models.Problem{
		Id:        uuid.New(),
		IsPrivate: false,
	}

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)

	canView, err := uc.CanViewProblem(ctx, userID, problem)
	assert.NoError(t, err)
	assert.True(t, canView)
	mockUsersRepo.AssertExpectations(t)
}

func TestUseCase_CanViewProblem_PrivateNoPermission(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	problemID := uuid.New()
	problem := &models.Problem{
		Id:        problemID,
		IsPrivate: true,
	}

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasAnyRelation", ctx, ResourceProblem, problemID, userID, []string{RelationOwner, RelationModerator}).Return(false, nil)

	canView, err := uc.CanViewProblem(ctx, userID, problem)
	assert.NoError(t, err)
	assert.False(t, canView)
	mockUsersRepo.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanEditProblem_Owner(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	problemID := uuid.New()

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasAnyRelation", ctx, ResourceProblem, problemID, userID, []string{RelationOwner, RelationModerator}).Return(true, nil)

	canEdit, err := uc.CanEditProblem(ctx, userID, problemID)
	assert.NoError(t, err)
	assert.True(t, canEdit)
	mockUsersRepo.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanAdminProblem_Owner(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	problemID := uuid.New()

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceProblem, problemID, userID, RelationOwner).Return(true, nil)

	canAdmin, err := uc.CanAdminProblem(ctx, userID, problemID)
	assert.NoError(t, err)
	assert.True(t, canAdmin)
	mockUsersRepo.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanViewMonitor_MonitorEnabled(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contest := &models.Contest{
		Id:             uuid.New(),
		IsPrivate:      false,
		MonitorEnabled: true,
	}

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contest.Id, userID, RelationOwner).Return(false, nil)

	canView, err := uc.CanViewMonitor(ctx, userID, contest)
	assert.NoError(t, err)
	assert.True(t, canView) // Monitor enabled + public contest = can view
	mockUsersRepo.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}

func TestUseCase_CanViewMonitor_MonitorDisabled_NotOwner(t *testing.T) {
	mockPermissionsRepo := new(MockPermissionsRepo)
	mockUsersRepo := new(MockUsersRepo)
	mockContestsReader := new(MockContestsReader)

	uc := NewUseCase(mockPermissionsRepo, mockUsersRepo, mockContestsReader)
	ctx := context.Background()

	userID := uuid.New()
	contest := &models.Contest{
		Id:             uuid.New(),
		IsPrivate:      false,
		MonitorEnabled: false,
	}

	regularUser := &models.User{
		Id:   userID,
		Role: "user",
	}

	mockUsersRepo.On("GetUserById", ctx, userID).Return(regularUser, nil)
	mockPermissionsRepo.On("HasPermission", ctx, ResourceContest, contest.Id, userID, RelationOwner).Return(false, nil)
	mockPermissionsRepo.On("HasAnyRelation", ctx, ResourceContest, contest.Id, userID, []string{RelationModerator}).Return(false, nil)

	canView, err := uc.CanViewMonitor(ctx, userID, contest)
	assert.NoError(t, err)
	assert.False(t, canView)
	mockUsersRepo.AssertExpectations(t)
	mockPermissionsRepo.AssertExpectations(t)
}
