package problems_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/internal/problems"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// setupTestDB creates a mocked sqlx.DB and sqlmock instance for runner.
func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestRepository_CreateProblem(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := problems.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		problem := models.Problem{
			Id:    uuid.New(),
			Title: "Test Problem",
		}

		mock.ExpectQuery(problems.CreateProblemQuery).
			WithArgs(problem.Title).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(problem.Id))

		id, err := repo.CreateProblem(ctx, db, problem.Title)
		assert.NoError(t, err)
		assert.Equal(t, problem.Id, id)
	})
}

func TestRepository_GetProblemById(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := problems.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		expected := &models.Problem{
			Id:               uuid.New(),
			Title:            "Test Problem",
			TimeLimit:        1000,
			MemoryLimit:      1024,
			Legend:           "Test Legend",
			InputFormat:      "Test Input Format",
			OutputFormat:     "Test Output Format",
			Notes:            "Test Notes",
			Scoring:          "Test Scoring",
			LegendHtml:       "Test Legend HTML",
			InputFormatHtml:  "Test Input Format HTML",
			OutputFormatHtml: "Test Output Format HTML",
			NotesHtml:        "Test Notes HTML",
			ScoringHtml:      "Test Scoring HTML",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		columns := []string{
			"id",
			"title",
			"time_limit",
			"memory_limit",

			"legend",
			"input_format",
			"output_format",
			"notes",
			"scoring",

			"legend_html",
			"input_format_html",
			"output_format_html",
			"notes_html",
			"scoring_html",

			"created_at",
			"updated_at",
		}

		rows := sqlmock.NewRows(columns).
			AddRow(
				expected.Id,
				expected.Title,
				expected.TimeLimit,
				expected.MemoryLimit,

				expected.Legend,
				expected.InputFormat,
				expected.OutputFormat,
				expected.Notes,
				expected.Scoring,

				expected.LegendHtml,
				expected.InputFormatHtml,
				expected.OutputFormatHtml,
				expected.NotesHtml,
				expected.ScoringHtml,

				expected.CreatedAt,
				expected.UpdatedAt)

		mock.ExpectQuery(problems.GetProblemByIdQuery).WithArgs(expected.Id).WillReturnRows(rows)

		problem, err := repo.GetProblemById(ctx, db, expected.Id)
		assert.NoError(t, err)
		assert.EqualExportedValues(t, expected, problem)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.Background()

		id := uuid.New()

		mock.ExpectQuery(problems.GetProblemByIdQuery).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := repo.GetProblemById(ctx, db, id)
		assert.Error(t, err)
	})
}

func TestRepository_DeleteProblem(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := problems.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		id := uuid.New()

		mock.ExpectExec(problems.DeleteProblemQuery).
			WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteProblem(ctx, db, id)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.Background()
		id := uuid.New()

		mock.ExpectExec(problems.DeleteProblemQuery).WithArgs(id).WillReturnError(sql.ErrNoRows)

		err := repo.DeleteProblem(ctx, db, id)
		assert.Error(t, err)
	})
}

//func TestRepository_ListProblems(t *testing.T) {
//	db, mock := setupTestDB(t)
//	defer db.Close()
//
//	repo := repository.NewRepository(db)
//
//	t.Run("success", func(t *testing.T) {
//		ctx := context.Background()
//
//		expected := make([]*models.ProblemsListItem, 0)
//		for i := 0; i < 10; i++ {
//			problem := &models.ProblemsListItem{
//				Id:          int32(i + 1),
//				Title:       fmt.Sprintf("Test Problem %d", i+1),
//				TimeLimit:   1000,
//				MemoryLimit: 1024,
//				CreatedAt:   time.Now(),
//				UpdatedAt:   time.Now(),
//			}
//
//			expected = append(expected, problem)
//		}
//
//		filter := models.ProblemsFilter{
//			Page:     1,
//			PageSize: 10,
//		}
//
//		var totalCount int32 = 10
//
//		columns := []string{
//			"id",
//			"title",
//			"time_limit",
//			"memory_limit",
//			"created_at",
//			"updated_at",
//		}
//
//		rows := sqlmock.NewRows(columns)
//		for _, problem := range expected {
//			rows = rows.AddRow(
//				problem.Id,
//				problem.Title,
//				problem.TimeLimit,
//				problem.MemoryLimit,
//				problem.CreatedAt,
//				problem.UpdatedAt,
//			)
//		}
//
//		mock.ExpectQuery(repository.ListProblemsQuery).WillReturnRows(rows)
//		mock.ExpectQuery(repository.CountProblemsQuery).
//			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalCount))
//
//		problems, err := repo.ListProblems(ctx, db, filter)
//		assert.NoError(t, err)
//		assert.Equal(t, expected, problems.Problems)
//		assert.Equal(t, models.Pagination{
//			Page:  1,
//			Total: 1,
//		}, problems.Pagination)
//	})
//}

//func TestRepository_UpdateProblem(t *testing.T) {
//	db, mock := setupTestDB(t)
//	defer db.Close()
//
//	repo := repository.NewRepository(db)
//
//	t.Run("success", func(t *testing.T) {
//		ctx := context.Background()
//		var id int32 = 1
//
//		update := &models.ProblemUpdate{
//			Title:            sp("Test Problem"),
//			TimeLimit:        ip(1000),
//			MemoryLimit:      ip(1024),
//			Legend:           sp("Test Legend"),
//			InputFormat:      sp("Test Input Format"),
//			OutputFormat:     sp("Test Output Format"),
//			Notes:            sp("Test Notes"),
//			Scoring:          sp("Test Scoring"),
//			LegendHtml:       sp("Test Legend HTML"),
//			InputFormatHtml:  sp("Test Input Format HTML"),
//			OutputFormatHtml: sp("Test Output Format HTML"),
//			NotesHtml:        sp("Test Notes HTML"),
//			ScoringHtml:      sp("Test Scoring HTML"),
//			Meta:             &models.Meta{},
//			Samples:          &[]models.Sample{},
//		}
//
//		mock.ExpectExec(repository.UpdateProblemQuery).WithArgs(
//			id,
//
//			update.Title,
//			update.TimeLimit,
//			update.MemoryLimit,
//
//			update.Legend,
//			update.InputFormat,
//			update.OutputFormat,
//			update.Notes,
//			update.Scoring,
//
//			update.LegendHtml,
//			update.InputFormatHtml,
//			update.OutputFormatHtml,
//			update.NotesHtml,
//			update.ScoringHtml,
//
//			update.Meta,
//			update.Samples,
//		).WillReturnResult(sqlmock.NewResult(1, 1))
//
//		err := repo.UpdateProblem(ctx, db, id, update)
//		assert.NoError(t, err)
//	})
//}
//
//func sp(s string) *string {
//	return &s
//}
//
//func ip(s int32) *int32 {
//	return &s
//}
