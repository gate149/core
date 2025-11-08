package users

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	testerv1 "github.com/gate149/contracts/core/v1"
	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations

type MockUsersUC struct {
	mock.Mock
}

func (m *MockUsersUC) CreateUser(ctx context.Context, user *models.UserCreation) (uuid.UUID, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUsersUC) ReadUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUsersUC) ReadUserByKratosId(ctx context.Context, kratosId string) (*models.User, error) {
	args := m.Called(ctx, kratosId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUsersUC) SearchUsers(ctx context.Context, searchQuery, role string, page, pageSize int) ([]*models.User, int, error) {
	args := m.Called(ctx, searchQuery, role, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}

func (m *MockUsersUC) UpdateUserInIndex(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Helper functions

func setupFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			statusCode := pkg.ToREST(err)
			return c.Status(statusCode).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
	return app
}

func createMockSession(userID string) *ory.Session {
	active := true
	identity := &ory.Identity{
		Id: userID,
	}
	session := ory.Session{
		Active:   &active,
		Identity: identity,
	}
	return (*ory.Session)(&session)
}

func createTestUser(id uuid.UUID, kratosId string) *models.User {
	return &models.User{
		Id:        id,
		Username:  "testuser",
		Role:      "user",
		KratosId:  &kratosId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

const sessionKey = "session"

// Tests

func TestGetMe_Success(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	userID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)

	app.Get("/me", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.GetMe(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.GetUserResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, userID, response.User.Id)
	assert.Equal(t, "testuser", response.User.Username)
	mockUsersUC.AssertExpectations(t)
}

func TestGetMe_NoSession(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	app.Get("/me", func(c *fiber.Ctx) error {
		// No session in context
		return handlers.GetMe(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	mockUsersUC.AssertNotCalled(t, "ReadUserByKratosId")
}

func TestGetUser_Success(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	userID := uuid.New()
	userIDStr := userID.String()
	kratosID := userIDStr // GetUser uses the UUID as kratosID

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		return handlers.GetUser(c, userID)
	})

	req := httptest.NewRequest("GET", "/users/"+userIDStr, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.GetUserResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, userID, response.User.Id)
	assert.Equal(t, "testuser", response.User.Username)
	mockUsersUC.AssertExpectations(t)
}

func TestGetUsers_Success(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	user1 := createTestUser(uuid.New(), "kratos-1")
	user2 := createTestUser(uuid.New(), "kratos-2")

	users := []*models.User{user1, user2}
	totalPages := 1

	mockUsersUC.On("SearchUsers", mock.Anything, "", "", 1, 10).Return(users, totalPages, nil)

	params := testerv1.GetUsersParams{
		Page:     1,
		PageSize: 10,
	}

	app.Get("/users", func(c *fiber.Ctx) error {
		return handlers.GetUsers(c, params)
	})

	req := httptest.NewRequest("GET", "/users?page=1&page_size=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.ListUsersResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, 2, len(response.Users))
	assert.Equal(t, int32(1), response.Pagination.Page)
	assert.Equal(t, int32(1), response.Pagination.Total)
	mockUsersUC.AssertExpectations(t)
}

func TestGetUsers_WithSearchQuery(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	user := createTestUser(uuid.New(), "kratos-1")
	users := []*models.User{user}
	totalPages := 1

	searchQuery := "testuser"
	mockUsersUC.On("SearchUsers", mock.Anything, searchQuery, "", 1, 10).Return(users, totalPages, nil)

	params := testerv1.GetUsersParams{
		Page:     1,
		PageSize: 10,
		Search:   &searchQuery,
	}

	app.Get("/users", func(c *fiber.Ctx) error {
		return handlers.GetUsers(c, params)
	})

	req := httptest.NewRequest("GET", "/users?page=1&page_size=10&search=testuser", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.ListUsersResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, 1, len(response.Users))
	assert.Equal(t, "testuser", response.Users[0].Username)
	mockUsersUC.AssertExpectations(t)
}

func TestGetUsers_WithRoleFilter(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	adminUser := &models.User{
		Id:        uuid.New(),
		Username:  "admin",
		Role:      "admin",
		KratosId:  sp("kratos-admin"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users := []*models.User{adminUser}
	totalPages := 1

	role := "admin"
	mockUsersUC.On("SearchUsers", mock.Anything, "", role, 1, 10).Return(users, totalPages, nil)

	params := testerv1.GetUsersParams{
		Page:     1,
		PageSize: 10,
		Role:     &role,
	}

	app.Get("/users", func(c *fiber.Ctx) error {
		return handlers.GetUsers(c, params)
	})

	req := httptest.NewRequest("GET", "/users?page=1&page_size=10&role=admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.ListUsersResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, 1, len(response.Users))
	assert.Equal(t, "admin", response.Users[0].Role)
	mockUsersUC.AssertExpectations(t)
}

func TestGetUsers_InvalidPage(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	// Page defaults to 1 when < 1
	users := []*models.User{}
	totalPages := 0

	mockUsersUC.On("SearchUsers", mock.Anything, "", "", 1, 10).Return(users, totalPages, nil)

	params := testerv1.GetUsersParams{
		Page:     0, // Invalid page
		PageSize: 10,
	}

	app.Get("/users", func(c *fiber.Ctx) error {
		return handlers.GetUsers(c, params)
	})

	req := httptest.NewRequest("GET", "/users?page=0&page_size=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) // Should still succeed with defaulted page

	mockUsersUC.AssertExpectations(t)
}

func TestGetUsers_InvalidPageSize(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockUsersUC)

	// PageSize defaults to 10 when invalid
	users := []*models.User{}
	totalPages := 0

	mockUsersUC.On("SearchUsers", mock.Anything, "", "", 1, 10).Return(users, totalPages, nil)

	params := testerv1.GetUsersParams{
		Page:     1,
		PageSize: 200, // Invalid page size (> 100)
	}

	app.Get("/users", func(c *fiber.Ctx) error {
		return handlers.GetUsers(c, params)
	})

	req := httptest.NewRequest("GET", "/users?page=1&page_size=200", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode) // Should still succeed with defaulted page size

	mockUsersUC.AssertExpectations(t)
}

func sp(s string) *string {
	return &s
}
