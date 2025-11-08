package problems

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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

type MockProblemsUC struct {
	mock.Mock
}

func (m *MockProblemsUC) CreateProblem(ctx context.Context, title string) (uuid.UUID, error) {
	args := m.Called(ctx, title)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockProblemsUC) GetProblemById(ctx context.Context, id uuid.UUID) (*models.Problem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Problem), args.Error(1)
}

func (m *MockProblemsUC) DownloadTestsArchive(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockProblemsUC) DeleteProblem(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProblemsUC) ListProblems(ctx context.Context, filter models.ProblemsFilter) (*models.ProblemsList, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProblemsList), args.Error(1)
}

func (m *MockProblemsUC) UpdateProblem(ctx context.Context, id uuid.UUID, problemUpdate *models.ProblemUpdate) error {
	args := m.Called(ctx, id, problemUpdate)
	return args.Error(0)
}

func (m *MockProblemsUC) UploadProblem(ctx context.Context, id uuid.UUID, r io.ReaderAt, size int64) error {
	args := m.Called(ctx, id, r, size)
	return args.Error(0)
}

type MockPermissionsUC struct {
	mock.Mock
}

func (m *MockPermissionsUC) CreatePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error {
	args := m.Called(ctx, resourceType, resourceID, userID, relation)
	return args.Error(0)
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

type MockUsersUC struct {
	mock.Mock
}

func (m *MockUsersUC) ReadUserByKratosId(ctx context.Context, kratosId string) (*models.User, error) {
	args := m.Called(ctx, kratosId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
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

func createTestProblem(id uuid.UUID, isPrivate bool) *models.Problem {
	return &models.Problem{
		Id:          id,
		Title:       "Test Problem",
		TimeLimit:   1000,
		MemoryLimit: 256,
		IsPrivate:   isPrivate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

const sessionKey = "session"

// Tests

func TestCreateProblem_Success(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockProblemsUC.On("CreateProblem", mock.Anything, "Test Problem").Return(problemID, nil)
	mockPermissionsUC.On("CreatePermission", mock.Anything, "problem", problemID, userID, "owner").Return(nil)

	params := testerv1.CreateProblemParams{
		Title: "Test Problem",
	}

	app.Post("/problems", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CreateProblem(c, params)
	})

	req := httptest.NewRequest("POST", "/problems", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.CreationResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, problemID, response.Id)
	mockProblemsUC.AssertExpectations(t)
	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
}

func TestCreateProblem_EmptyTitle(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)

	params := testerv1.CreateProblemParams{
		Title: "",
	}

	app.Post("/problems", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.CreateProblem(c, params)
	})

	req := httptest.NewRequest("POST", "/problems", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	mockUsersUC.AssertExpectations(t)
	mockProblemsUC.AssertNotCalled(t, "CreateProblem")
}

func TestGetProblem_Success(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)
	problem := createTestProblem(problemID, false)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockProblemsUC.On("GetProblemById", mock.Anything, problemID).Return(problem, nil)
	mockPermissionsUC.On("CanViewProblem", mock.Anything, userID, problem).Return(true, nil)

	app.Get("/problems/:id", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.GetProblem(c, problemID)
	})

	req := httptest.NewRequest("GET", "/problems/"+problemID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.GetProblemResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, problemID, response.Problem.Id)
	assert.Equal(t, "Test Problem", response.Problem.Title)
	mockProblemsUC.AssertExpectations(t)
	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
}

func TestGetProblem_NoPermission(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)
	problem := createTestProblem(problemID, true) // private problem

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockProblemsUC.On("GetProblemById", mock.Anything, problemID).Return(problem, nil)
	mockPermissionsUC.On("CanViewProblem", mock.Anything, userID, problem).Return(false, nil)

	app.Get("/problems/:id", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.GetProblem(c, problemID)
	})

	req := httptest.NewRequest("GET", "/problems/"+problemID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)

	mockProblemsUC.AssertExpectations(t)
	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
}

func TestUpdateProblem_Success(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)
	newTitle := "Updated Problem"
	timeLimit := int32(2000)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockPermissionsUC.On("CanEditProblem", mock.Anything, userID, problemID).Return(true, nil)
	mockProblemsUC.On("UpdateProblem", mock.Anything, problemID, mock.MatchedBy(func(update *models.ProblemUpdate) bool {
		return update.Title != nil && *update.Title == newTitle && update.TimeLimit != nil && *update.TimeLimit == timeLimit
	})).Return(nil)

	reqBody := testerv1.UpdateProblemRequest{
		Title:     &newTitle,
		TimeLimit: &timeLimit,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Put("/problems/:id", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.UpdateProblem(c, problemID)
	})

	req := httptest.NewRequest("PUT", "/problems/"+problemID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockProblemsUC.AssertExpectations(t)
	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
}

func TestUpdateProblem_NoPermission(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)
	newTitle := "Updated Problem"

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockPermissionsUC.On("CanEditProblem", mock.Anything, userID, problemID).Return(false, nil)

	reqBody := testerv1.UpdateProblemRequest{
		Title: &newTitle,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	app.Put("/problems/:id", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.UpdateProblem(c, problemID)
	})

	req := httptest.NewRequest("PUT", "/problems/"+problemID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)

	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
	mockProblemsUC.AssertNotCalled(t, "UpdateProblem")
}

func TestDeleteProblem_Success(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockPermissionsUC.On("CanAdminProblem", mock.Anything, userID, problemID).Return(true, nil)
	mockProblemsUC.On("DeleteProblem", mock.Anything, problemID).Return(nil)

	app.Delete("/problems/:id", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.DeleteProblem(c, problemID)
	})

	req := httptest.NewRequest("DELETE", "/problems/"+problemID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockProblemsUC.AssertExpectations(t)
	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
}

func TestDeleteProblem_NoPermission(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockPermissionsUC.On("CanAdminProblem", mock.Anything, userID, problemID).Return(false, nil)

	app.Delete("/problems/:id", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.DeleteProblem(c, problemID)
	})

	req := httptest.NewRequest("DELETE", "/problems/"+problemID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)

	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
	mockProblemsUC.AssertNotCalled(t, "DeleteProblem")
}

func TestListProblems_Success(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)
	problemID := uuid.New()
	problem := &models.ProblemsListItem{
		Id:          problemID,
		Title:       "Test Problem",
		TimeLimit:   1000,
		MemoryLimit: 256,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	problemsList := &models.ProblemsList{
		Problems: []*models.ProblemsListItem{problem},
		Pagination: models.Pagination{
			Page:  1,
			Total: 1,
		},
	}

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockProblemsUC.On("ListProblems", mock.Anything, mock.MatchedBy(func(filter models.ProblemsFilter) bool {
		return filter.Page == 1 && filter.PageSize == 10
	})).Return(problemsList, nil)

	params := testerv1.ListProblemsParams{
		Page:     1,
		PageSize: 10,
	}

	app.Get("/problems", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.ListProblems(c, params)
	})

	req := httptest.NewRequest("GET", "/problems?page=1&page_size=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response testerv1.ListProblemsResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, 1, len(response.Problems))
	assert.Equal(t, problemID, response.Problems[0].Id)
	mockProblemsUC.AssertExpectations(t)
	mockUsersUC.AssertExpectations(t)
}

func TestUploadProblem_Success(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockPermissionsUC.On("CanEditProblem", mock.Anything, userID, problemID).Return(true, nil)
	mockProblemsUC.On("UploadProblem", mock.Anything, problemID, mock.Anything, mock.Anything).Return(nil)

	app.Post("/problems/:id/upload", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.UploadProblem(c, problemID)
	})

	// Create multipart form with file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("archive", "test.zip")
	part.Write([]byte("test archive content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/problems/"+problemID.String()+"/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockProblemsUC.AssertExpectations(t)
	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
}

func TestUploadProblem_NoPermission(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	userID := uuid.New()
	problemID := uuid.New()
	userIDStr := userID.String()
	kratosID := "kratos-" + userIDStr

	user := createTestUser(userID, kratosID)

	mockUsersUC.On("ReadUserByKratosId", mock.Anything, kratosID).Return(user, nil)
	mockPermissionsUC.On("CanEditProblem", mock.Anything, userID, problemID).Return(false, nil)

	app.Post("/problems/:id/upload", func(c *fiber.Ctx) error {
		c.Locals(sessionKey, createMockSession(kratosID))
		return handlers.UploadProblem(c, problemID)
	})

	// Create multipart form with file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("archive", "test.zip")
	part.Write([]byte("test archive content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/problems/"+problemID.String()+"/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)

	mockUsersUC.AssertExpectations(t)
	mockPermissionsUC.AssertExpectations(t)
	mockProblemsUC.AssertNotCalled(t, "UploadProblem")
}

func TestCreateProblem_Unauthenticated(t *testing.T) {
	app := setupFiberApp()
	mockProblemsUC := new(MockProblemsUC)
	mockPermissionsUC := new(MockPermissionsUC)
	mockUsersUC := new(MockUsersUC)

	handlers := NewHandlers(mockProblemsUC, mockPermissionsUC, mockUsersUC)

	params := testerv1.CreateProblemParams{
		Title: "Test Problem",
	}

	app.Post("/problems", func(c *fiber.Ctx) error {
		// No session in context
		return handlers.CreateProblem(c, params)
	})

	req := httptest.NewRequest("POST", "/problems", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	mockProblemsUC.AssertNotCalled(t, "CreateProblem")
	mockUsersUC.AssertNotCalled(t, "ReadUserByKratosId")
}
