package problems

import (
	"context"
	"database/sql"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gate149/core/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) BeginTx(ctx context.Context) (Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Tx), args.Error(1)
}

func (m *MockRepo) DB() Querier {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(Querier)
}

func (m *MockRepo) CreateProblem(ctx context.Context, q Querier, title string) (uuid.UUID, error) {
	args := m.Called(ctx, q, title)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRepo) GetProblemById(ctx context.Context, q Querier, id uuid.UUID) (*models.Problem, error) {
	args := m.Called(ctx, q, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Problem), args.Error(1)
}

func (m *MockRepo) DeleteProblem(ctx context.Context, q Querier, id uuid.UUID) error {
	args := m.Called(ctx, q, id)
	return args.Error(0)
}

func (m *MockRepo) ListProblems(ctx context.Context, q Querier, filter models.ProblemsFilter) (*models.ProblemsList, error) {
	args := m.Called(ctx, q, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProblemsList), args.Error(1)
}

func (m *MockRepo) UpdateProblem(ctx context.Context, q Querier, id uuid.UUID, heading *models.ProblemUpdate) error {
	args := m.Called(ctx, q, id, heading)
	return args.Error(0)
}

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rebind(query string) string {
	args := m.Called(query)
	return args.String(0)
}

func (m *MockTx) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*sqlx.Rows), callArgs.Error(1)
}

func (m *MockTx) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	callArgs := m.Called(ctx, dest, query, args)
	return callArgs.Error(0)
}

func (m *MockTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	callArgs := m.Called(ctx, query, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockTx) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	callArgs := m.Called(ctx, dest, query, args)
	return callArgs.Error(0)
}

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) Rebind(query string) string {
	args := m.Called(query)
	return args.String(0)
}

func (m *MockQuerier) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*sqlx.Rows), callArgs.Error(1)
}

func (m *MockQuerier) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	callArgs := m.Called(ctx, dest, query, args)
	return callArgs.Error(0)
}

func (m *MockQuerier) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	callArgs := m.Called(ctx, query, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockQuerier) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	callArgs := m.Called(ctx, dest, query, args)
	return callArgs.Error(0)
}

type MockPandocClient struct {
	mock.Mock
}

func (m *MockPandocClient) ConvertLatexToHtml5(ctx context.Context, text string) (string, error) {
	args := m.Called(ctx, text)
	return args.String(0), args.Error(1)
}

func (m *MockPandocClient) BatchConvertLatexToHtml5(ctx context.Context, latex []string) ([]string, error) {
	args := m.Called(ctx, latex)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

type MockS3Repo struct {
	mock.Mock
}

func (m *MockS3Repo) UploadTestsFile(ctx context.Context, id uuid.UUID, reader io.Reader) (string, error) {
	args := m.Called(ctx, id, reader)
	return args.String(0), args.Error(1)
}

func (m *MockS3Repo) DownloadTestsFile(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// Tests

func TestUseCase_CreateProblem(t *testing.T) {
	mockRepo := new(MockRepo)
	mockPandoc := new(MockPandocClient)
	mockS3 := new(MockS3Repo)
	mockQuerier := new(MockQuerier)

	mockRepo.On("DB").Return(mockQuerier)

	uc, err := NewUseCase(mockRepo, mockPandoc, mockS3, "/tmp/test-cache")
	assert.NoError(t, err)

	ctx := context.Background()
	title := "Test Problem"
	expectedID := uuid.New()

	mockRepo.On("CreateProblem", ctx, mockQuerier, title).Return(expectedID, nil)

	id, err := uc.CreateProblem(ctx, title)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_GetProblemById(t *testing.T) {
	mockRepo := new(MockRepo)
	mockPandoc := new(MockPandocClient)
	mockS3 := new(MockS3Repo)
	mockQuerier := new(MockQuerier)

	mockRepo.On("DB").Return(mockQuerier)

	uc, err := NewUseCase(mockRepo, mockPandoc, mockS3, "/tmp/test-cache")
	assert.NoError(t, err)

	ctx := context.Background()
	id := uuid.New()
	expectedProblem := &models.Problem{
		Id:          id,
		Title:       "Test Problem",
		TimeLimit:   1000,
		MemoryLimit: 256,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetProblemById", ctx, mockQuerier, id).Return(expectedProblem, nil)

	problem, err := uc.GetProblemById(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, expectedProblem, problem)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_DeleteProblem(t *testing.T) {
	mockRepo := new(MockRepo)
	mockPandoc := new(MockPandocClient)
	mockS3 := new(MockS3Repo)
	mockQuerier := new(MockQuerier)

	mockRepo.On("DB").Return(mockQuerier)

	uc, err := NewUseCase(mockRepo, mockPandoc, mockS3, "/tmp/test-cache")
	assert.NoError(t, err)

	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("DeleteProblem", ctx, mockQuerier, id).Return(nil)

	err = uc.DeleteProblem(ctx, id)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_ListProblems(t *testing.T) {
	mockRepo := new(MockRepo)
	mockPandoc := new(MockPandocClient)
	mockS3 := new(MockS3Repo)
	mockQuerier := new(MockQuerier)

	mockRepo.On("DB").Return(mockQuerier)

	uc, err := NewUseCase(mockRepo, mockPandoc, mockS3, "/tmp/test-cache")
	assert.NoError(t, err)

	ctx := context.Background()
	filter := models.ProblemsFilter{
		Page:     1,
		PageSize: 10,
	}

	expectedList := &models.ProblemsList{
		Problems: []*models.ProblemsListItem{
			{
				Id:          uuid.New(),
				Title:       "Test Problem 1",
				TimeLimit:   1000,
				MemoryLimit: 256,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		Pagination: models.Pagination{
			Page:  1,
			Total: 1,
		},
	}

	mockRepo.On("ListProblems", ctx, mockQuerier, filter).Return(expectedList, nil)

	list, err := uc.ListProblems(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, expectedList, list)
	assert.Equal(t, 1, len(list.Problems))
	mockRepo.AssertExpectations(t)
}

func TestUseCase_UpdateProblem(t *testing.T) {
	mockRepo := new(MockRepo)
	mockPandoc := new(MockPandocClient)
	mockS3 := new(MockS3Repo)
	mockTx := new(MockTx)

	uc, err := NewUseCase(mockRepo, mockPandoc, mockS3, "/tmp/test-cache")
	assert.NoError(t, err)

	ctx := context.Background()
	id := uuid.New()
	newTitle := "Updated Problem"
	newLegend := "Updated legend"

	existingProblem := &models.Problem{
		Id:          id,
		Title:       "Old Title",
		TimeLimit:   1000,
		MemoryLimit: 256,
		Legend:      "Old legend",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	update := &models.ProblemUpdate{
		Title:  &newTitle,
		Legend: &newLegend,
	}

	mockRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockRepo.On("GetProblemById", ctx, mockTx, id).Return(existingProblem, nil)
	mockPandoc.On("BatchConvertLatexToHtml5", ctx, mock.MatchedBy(func(latex []string) bool {
		return len(latex) == 5 // legend, input, output, notes, scoring
	})).Return([]string{
		"<p>Updated legend HTML</p>",
		"", // input format
		"", // output format
		"", // notes
		"", // scoring
	}, nil)
	mockRepo.On("UpdateProblem", ctx, mockTx, id, mock.MatchedBy(func(u *models.ProblemUpdate) bool {
		return u.Title != nil && *u.Title == newTitle &&
			u.LegendHtml != nil && strings.Contains(*u.LegendHtml, "Updated legend HTML")
	})).Return(nil)
	mockTx.On("Commit").Return(nil)

	err = uc.UpdateProblem(ctx, id, update)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPandoc.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestUseCase_UpdateProblem_EmptyUpdate(t *testing.T) {
	mockRepo := new(MockRepo)
	mockPandoc := new(MockPandocClient)
	mockS3 := new(MockS3Repo)

	uc, err := NewUseCase(mockRepo, mockPandoc, mockS3, "/tmp/test-cache")
	assert.NoError(t, err)

	ctx := context.Background()
	id := uuid.New()
	emptyUpdate := &models.ProblemUpdate{}

	err = uc.UpdateProblem(ctx, id, emptyUpdate)
	assert.Error(t, err)
	// Check that it's a bad input error
	assert.Contains(t, err.Error(), "empty problem update")
}

func TestUseCase_DownloadTestsArchive(t *testing.T) {
	mockRepo := new(MockRepo)
	mockPandoc := new(MockPandocClient)
	mockS3 := new(MockS3Repo)

	uc, err := NewUseCase(mockRepo, mockPandoc, mockS3, "/tmp/test-cache")
	assert.NoError(t, err)

	ctx := context.Background()
	id := uuid.New()

	// Mock ReadCloser
	mockReader := &mockReadCloser{Reader: strings.NewReader("test data")}
	mockS3.On("DownloadTestsFile", ctx, id).Return(mockReader, nil)

	path, err := uc.DownloadTestsArchive(ctx, id)
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	mockS3.AssertExpectations(t)
}

// mockReadCloser is a helper type that implements io.ReadCloser
type mockReadCloser struct {
	*strings.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}
