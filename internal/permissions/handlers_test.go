package permissions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations

type MockPermissionsUC struct {
	mock.Mock
}

func (m *MockPermissionsUC) CreatePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error {
	args := m.Called(ctx, resourceType, resourceID, userID, relation)
	return args.Error(0)
}

func (m *MockPermissionsUC) DeletePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error {
	args := m.Called(ctx, resourceType, resourceID, userID, relation)
	return args.Error(0)
}

func (m *MockPermissionsUC) CanViewContest(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error) {
	args := m.Called(ctx, userID, contest)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionsUC) CanEditContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, contestID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionsUC) CanAdminContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, contestID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionsUC) CanViewMonitor(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error) {
	args := m.Called(ctx, userID, contest)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionsUC) CanViewProblem(ctx context.Context, userID uuid.UUID, problem *models.Problem) (bool, error) {
	args := m.Called(ctx, userID, problem)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionsUC) CanEditProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, problemID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionsUC) CanAdminProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, problemID)
	return args.Bool(0), args.Error(1)
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

const sessionKey = "session"

// Tests

func TestCheckPermission_Contest_Edit_Success(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	userID := uuid.New()
	contestID := uuid.New()
	userIDStr := userID.String()
	kratosID := userIDStr

	mockPermissionsUC.On("CanEditContest", mock.Anything, userID, contestID).Return(true, nil)

	reqBody := CheckPermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   contestID.String(),
		Permission:   "edit",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/check", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CheckPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/check", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response CheckPermissionResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.True(t, response.Allowed)
	mockPermissionsUC.AssertExpectations(t)
}

func TestCheckPermission_Contest_Admin_Success(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	userID := uuid.New()
	contestID := uuid.New()
	userIDStr := userID.String()
	kratosID := userIDStr

	mockPermissionsUC.On("CanAdminContest", mock.Anything, userID, contestID).Return(true, nil)

	reqBody := CheckPermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   contestID.String(),
		Permission:   "admin",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/check", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CheckPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/check", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response CheckPermissionResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.True(t, response.Allowed)
	mockPermissionsUC.AssertExpectations(t)
}

func TestCheckPermission_Problem_Edit_Success(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := userIDStr

	mockPermissionsUC.On("CanEditProblem", mock.Anything, userID, problemID).Return(true, nil)

	reqBody := CheckPermissionRequest{
		ResourceType: ResourceProblem,
		ResourceID:   problemID.String(),
		Permission:   "edit",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/check", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CheckPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/check", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response CheckPermissionResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.True(t, response.Allowed)
	mockPermissionsUC.AssertExpectations(t)
}

func TestCheckPermission_NoPermission(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	userID := uuid.New()
	contestID := uuid.New()
	userIDStr := userID.String()
	kratosID := userIDStr

	mockPermissionsUC.On("CanEditContest", mock.Anything, userID, contestID).Return(false, nil)

	reqBody := CheckPermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   contestID.String(),
		Permission:   "edit",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/check", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CheckPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/check", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response CheckPermissionResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.False(t, response.Allowed)
	mockPermissionsUC.AssertExpectations(t)
}

func TestCheckPermission_InvalidResourceType(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	userID := uuid.New()
	userIDStr := userID.String()
	kratosID := userIDStr

	reqBody := CheckPermissionRequest{
		ResourceType: "invalid",
		ResourceID:   uuid.New().String(),
		Permission:   "edit",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/check", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CheckPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/check", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestCheckPermission_InvalidPermission(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	userID := uuid.New()
	contestID := uuid.New()
	userIDStr := userID.String()
	kratosID := userIDStr

	reqBody := CheckPermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   contestID.String(),
		Permission:   "invalid",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/check", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CheckPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/check", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestCheckPermission_Unauthenticated(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	reqBody := CheckPermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   uuid.New().String(),
		Permission:   "edit",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/check", func(c *fiber.Ctx) error {
		// No session in context
		return handlers.CheckPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/check", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestGrantPermission_Success(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	adminID := uuid.New()
	contestID := uuid.New()
	targetUserID := uuid.New()
	adminIDStr := adminID.String()

	mockPermissionsUC.On("CreatePermission", mock.Anything, ResourceContest, contestID, targetUserID, RelationModerator).Return(nil)

	reqBody := GrantPermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   contestID.String(),
		UserID:       targetUserID.String(),
		Relation:     RelationModerator,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/grant", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(adminIDStr))
		return handlers.GrantPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/grant", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockPermissionsUC.AssertExpectations(t)
}

func TestGrantPermission_InvalidResourceID(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	adminID := uuid.New()
	adminIDStr := adminID.String()

	reqBody := GrantPermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   "invalid-uuid",
		UserID:       uuid.New().String(),
		Relation:     RelationModerator,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/grant", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(adminIDStr))
		return handlers.GrantPermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/grant", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	mockPermissionsUC.AssertNotCalled(t, "CreatePermission")
}

func TestRevokePermission_Success(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	adminID := uuid.New()
	contestID := uuid.New()
	targetUserID := uuid.New()
	adminIDStr := adminID.String()

	mockPermissionsUC.On("DeletePermission", mock.Anything, ResourceContest, contestID, targetUserID, RelationModerator).Return(nil)

	reqBody := RevokePermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   contestID.String(),
		UserID:       targetUserID.String(),
		Relation:     RelationModerator,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/revoke", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(adminIDStr))
		return handlers.RevokePermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/revoke", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockPermissionsUC.AssertExpectations(t)
}

func TestRevokePermission_InvalidUserID(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	adminID := uuid.New()
	adminIDStr := adminID.String()

	reqBody := RevokePermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   uuid.New().String(),
		UserID:       "invalid-uuid",
		Relation:     RelationModerator,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/revoke", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(adminIDStr))
		return handlers.RevokePermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/revoke", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	mockPermissionsUC.AssertNotCalled(t, "DeletePermission")
}

func TestRevokePermission_Unauthenticated(t *testing.T) {
	app := setupFiberApp()
	mockPermissionsUC := new(MockPermissionsUC)

	handlers := NewHandlers(mockPermissionsUC)

	reqBody := RevokePermissionRequest{
		ResourceType: ResourceContest,
		ResourceID:   uuid.New().String(),
		UserID:       uuid.New().String(),
		Relation:     RelationModerator,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Post("/permissions/revoke", func(c *fiber.Ctx) error {
		// No session in context
		return handlers.RevokePermission(c)
	})

	req := httptest.NewRequest("POST", "/permissions/revoke", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	mockPermissionsUC.AssertNotCalled(t, "DeletePermission")
}
