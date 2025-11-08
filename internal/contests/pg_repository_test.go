package contests_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gate149/core/internal/contests"
	"github.com/gate149/core/internal/models"
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

func TestRepository_CreateContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := contests.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		contest := models.Contest{
			Id:    uuid.New(),
			Title: "Test Contest",
		}

		contestCreation := models.ContestCreation{
			Title: contest.Title,
		}

		mock.ExpectQuery(contests.CreateContestQuery).
			WithArgs(contestCreation.Title).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contest.Id))

		id, err := repo.CreateContest(ctx, contestCreation)
		assert.NoError(t, err)
		assert.Equal(t, contest.Id, id)
	})
}

func TestRepository_GetContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := contests.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		contest := models.Contest{
			Id:        uuid.New(),
			Title:     "Test Contest",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectQuery(contests.GetContestQuery).
			WithArgs(contest.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at", "updated_at"}).
				AddRow(contest.Id, contest.Title, contest.CreatedAt, contest.UpdatedAt))

		result, err := repo.GetContest(ctx, contest.Id)
		assert.NoError(t, err)
		assert.EqualExportedValues(t, &contest, result)
	})
}

func TestRepository_UpdateContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := contests.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		contestId := uuid.New()
		update := models.ContestUpdate{
			Title: sp("Updated Contest"),
		}

		// UpdateContest uses static SQL with COALESCE
		expectedQuery := "UPDATE contests SET title = COALESCE($2, title), is_private = COALESCE($3, is_private), monitor_enabled = COALESCE($4, monitor_enabled) WHERE id = $1"
		mock.ExpectExec(expectedQuery).
			WithArgs(contestId, update.Title, update.IsPrivate, update.MonitorEnabled).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateContest(ctx, contestId, update)
		assert.NoError(t, err)
	})
}

func TestRepository_DeleteContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := contests.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		contestId := uuid.New()
		mock.ExpectExec(contests.DeleteContestQuery).
			WithArgs(contestId).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteContest(ctx, contestId)
		assert.NoError(t, err)
	})
}

func sp(s string) *string {
	return &s
}
