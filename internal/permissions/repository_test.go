package permissions_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gate149/core/internal/permissions"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestRepository_CreatePermission(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := permissions.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		mock.ExpectExec(permissions.CreatePermissionQuery).
			WithArgs(sqlmock.AnyArg(), permissions.ResourceContest, resourceID, userID, permissions.RelationOwner).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.CreatePermission(ctx, permissions.ResourceContest, resourceID, userID, permissions.RelationOwner)
		assert.NoError(t, err)
	})

	t.Run("invalid resource type", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		err := repo.CreatePermission(ctx, "invalid", resourceID, userID, permissions.RelationOwner)
		assert.Error(t, err)
	})

	t.Run("invalid relation", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		err := repo.CreatePermission(ctx, permissions.ResourceContest, resourceID, userID, "invalid")
		assert.Error(t, err)
	})
}

func TestRepository_DeletePermission(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := permissions.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		mock.ExpectExec(permissions.DeletePermissionQuery).
			WithArgs(permissions.ResourceContest, resourceID, userID, permissions.RelationOwner).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeletePermission(ctx, permissions.ResourceContest, resourceID, userID, permissions.RelationOwner)
		assert.NoError(t, err)
	})

	t.Run("invalid resource type", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		err := repo.DeletePermission(ctx, "invalid", resourceID, userID, permissions.RelationOwner)
		assert.Error(t, err)
	})
}

func TestRepository_HasPermission(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := permissions.NewRepository(db)

	t.Run("has permission", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		mock.ExpectQuery(permissions.HasPermissionQuery).
			WithArgs(permissions.ResourceContest, resourceID, userID, permissions.RelationOwner).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		hasPermission, err := repo.HasPermission(ctx, permissions.ResourceContest, resourceID, userID, permissions.RelationOwner)
		assert.NoError(t, err)
		assert.True(t, hasPermission)
	})

	t.Run("no permission", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		mock.ExpectQuery(permissions.HasPermissionQuery).
			WithArgs(permissions.ResourceContest, resourceID, userID, permissions.RelationOwner).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		hasPermission, err := repo.HasPermission(ctx, permissions.ResourceContest, resourceID, userID, permissions.RelationOwner)
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})

	t.Run("invalid resource type", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		_, err := repo.HasPermission(ctx, "invalid", resourceID, userID, permissions.RelationOwner)
		assert.Error(t, err)
	})
}

func TestRepository_GetUserPermissions(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := permissions.NewRepository(db)

	t.Run("multiple permissions", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		mock.ExpectQuery(permissions.GetUserPermissionsQuery).
			WithArgs(permissions.ResourceContest, resourceID, userID).
			WillReturnRows(sqlmock.NewRows([]string{"relation"}).
				AddRow(permissions.RelationOwner).
				AddRow(permissions.RelationModerator))

		relations, err := repo.GetUserPermissions(ctx, permissions.ResourceContest, resourceID, userID)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(relations))
		assert.Contains(t, relations, permissions.RelationOwner)
		assert.Contains(t, relations, permissions.RelationModerator)
	})

	t.Run("no permissions", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()

		mock.ExpectQuery(permissions.GetUserPermissionsQuery).
			WithArgs(permissions.ResourceContest, resourceID, userID).
			WillReturnRows(sqlmock.NewRows([]string{"relation"}))

		relations, err := repo.GetUserPermissions(ctx, permissions.ResourceContest, resourceID, userID)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(relations))
	})
}

func TestRepository_HasAnyRelation(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := permissions.NewRepository(db)

	t.Run("has one of relations", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()
		relations := []string{permissions.RelationOwner, permissions.RelationModerator}

		mock.ExpectQuery(permissions.HasAnyRelationQuery).
			WithArgs(permissions.ResourceContest, resourceID, userID, pq.Array(relations)).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		hasRelation, err := repo.HasAnyRelation(ctx, permissions.ResourceContest, resourceID, userID, relations)
		assert.NoError(t, err)
		assert.True(t, hasRelation)
	})

	t.Run("no relations", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()
		relations := []string{permissions.RelationOwner, permissions.RelationModerator}

		mock.ExpectQuery(permissions.HasAnyRelationQuery).
			WithArgs(permissions.ResourceContest, resourceID, userID, pq.Array(relations)).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		hasRelation, err := repo.HasAnyRelation(ctx, permissions.ResourceContest, resourceID, userID, relations)
		assert.NoError(t, err)
		assert.False(t, hasRelation)
	})

	t.Run("invalid relation in list", func(t *testing.T) {
		ctx := context.Background()

		resourceID := uuid.New()
		userID := uuid.New()
		relations := []string{permissions.RelationOwner, "invalid"}

		_, err := repo.HasAnyRelation(ctx, permissions.ResourceContest, resourceID, userID, relations)
		assert.Error(t, err)
	})
}
