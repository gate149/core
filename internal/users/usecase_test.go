package users

import (
	"context"
	"testing"

	"github.com/gate149/core/internal/models"
	"github.com/google/uuid"
)

// MockRepo is a mock implementation of Repo
type MockRepo struct {
	CreatedUsers []*models.UserCreation
	CreateError  error
}

func (m *MockRepo) CreateUser(ctx context.Context, user *models.UserCreation) (uuid.UUID, error) {
	if m.CreateError != nil {
		return uuid.Nil, m.CreateError
	}
	m.CreatedUsers = append(m.CreatedUsers, user)
	return user.Id, nil
}

func (m *MockRepo) GetUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, nil
}

func (m *MockRepo) GetUserByKratosId(ctx context.Context, kratosId string) (*models.User, error) {
	return nil, nil
}

func (m *MockRepo) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int, error) {
	return nil, 0, nil
}

func (m *MockRepo) SearchUsers(ctx context.Context, searchQuery, role string, page, pageSize int) ([]*models.User, int, error) {
	return nil, 0, nil
}

func TestCreateUser(t *testing.T) {
	// Setup
	mockRepo := &MockRepo{}
	useCase := NewUseCase(mockRepo)

	// Test data
	userCreation := &models.UserCreation{
		Id:       uuid.New(),
		Username: "testuser",
		Role:     "user",
		KratosId: stringPtr("kratos-123"),
	}

	// Execute
	ctx := context.Background()
	id, err := useCase.CreateUser(ctx, userCreation)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id != userCreation.Id {
		t.Fatalf("Expected ID %v, got %v", userCreation.Id, id)
	}

	// Check that user was created in repo
	if len(mockRepo.CreatedUsers) != 1 {
		t.Fatalf("Expected 1 user created in repo, got %d", len(mockRepo.CreatedUsers))
	}
}

func stringPtr(s string) *string {
	return &s
}
