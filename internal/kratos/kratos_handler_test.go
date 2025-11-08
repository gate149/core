package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gate149/core/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	return fiber.New()
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Set to error to avoid cluttering test output
	}))
}

// Tests

func TestHandleKratosWebhook_Success(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)
	logger := createTestLogger()

	handler := NewKratosHandler(mockUsersUC, logger)

	kratosID := "kratos-" + uuid.New().String()
	username := "testuser"

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(nil, nil)
	mockUsersUC.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.UserCreation) bool {
		return user.Username == username && *user.KratosId == kratosID && user.Role == "user"
	})).Return(uuid.New(), nil)

	reqBody := KratosWebhookRequest{
		UserId:   kratosID,
		Username: username,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/webhooks/kratos", handler.HandleKratosWebhook)

	req := httptest.NewRequest("POST", "/webhooks/kratos", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response KratosWebhookResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.True(t, response.Success)
	assert.Equal(t, "User created successfully", response.Message)
	mockUsersUC.AssertExpectations(t)
}

func TestHandleKratosWebhook_UserAlreadyExists(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)
	logger := createTestLogger()

	handler := NewKratosHandler(mockUsersUC, logger)

	kratosID := "kratos-" + uuid.New().String()
	username := "testuser"
	existingUser := &models.User{
		Id:       uuid.New(),
		Username: username,
		Role:     "user",
		KratosId: &kratosID,
	}

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(existingUser, nil)

	reqBody := KratosWebhookRequest{
		UserId:   kratosID,
		Username: username,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/webhooks/kratos", handler.HandleKratosWebhook)

	req := httptest.NewRequest("POST", "/webhooks/kratos", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response KratosWebhookResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.True(t, response.Success)
	assert.Equal(t, "User already exists", response.Message)
	mockUsersUC.AssertExpectations(t)
	mockUsersUC.AssertNotCalled(t, "CreateUser")
}

func TestHandleKratosWebhook_MissingUserId(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)
	logger := createTestLogger()

	handler := NewKratosHandler(mockUsersUC, logger)

	reqBody := KratosWebhookRequest{
		UserId:   "",
		Username: "testuser",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/webhooks/kratos", handler.HandleKratosWebhook)

	req := httptest.NewRequest("POST", "/webhooks/kratos", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var response KratosWebhookResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "Missing required fields")
	mockUsersUC.AssertNotCalled(t, "CreateUser")
}

func TestHandleKratosWebhook_MissingUsername(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)
	logger := createTestLogger()

	handler := NewKratosHandler(mockUsersUC, logger)

	reqBody := KratosWebhookRequest{
		UserId:   "kratos-" + uuid.New().String(),
		Username: "",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/webhooks/kratos", handler.HandleKratosWebhook)

	req := httptest.NewRequest("POST", "/webhooks/kratos", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var response KratosWebhookResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "Missing required fields")
	mockUsersUC.AssertNotCalled(t, "CreateUser")
}

func TestHandleKratosWebhook_InvalidBody(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)
	logger := createTestLogger()

	handler := NewKratosHandler(mockUsersUC, logger)

	app.Post("/webhooks/kratos", handler.HandleKratosWebhook)

	req := httptest.NewRequest("POST", "/webhooks/kratos", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var response KratosWebhookResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Error)
	mockUsersUC.AssertNotCalled(t, "CreateUser")
}

func TestHandleKratosWebhook_CreateUserError(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)
	logger := createTestLogger()

	handler := NewKratosHandler(mockUsersUC, logger)

	kratosID := "kratos-" + uuid.New().String()
	username := "testuser"

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(nil, nil)
	mockUsersUC.On("CreateUser", mock.Anything, mock.Anything).Return(uuid.Nil, assert.AnError)

	reqBody := KratosWebhookRequest{
		UserId:   kratosID,
		Username: username,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/webhooks/kratos", handler.HandleKratosWebhook)

	req := httptest.NewRequest("POST", "/webhooks/kratos", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	var response KratosWebhookResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "Failed to create user")
	mockUsersUC.AssertExpectations(t)
}

func TestHealthCheck_Success(t *testing.T) {
	app := setupFiberApp()
	mockUsersUC := new(MockUsersUC)
	logger := createTestLogger()

	handler := NewKratosHandler(mockUsersUC, logger)

	app.Get("/health", handler.HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "tester-private-server", response["service"])
	assert.NotNil(t, response["timestamp"])
}
