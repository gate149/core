package users

import (
	"context"

	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	_ "embed"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

//go:embed sql/create_user.sql
var CreateUserQuery string

func (r *Repository) CreateUser(ctx context.Context, user *models.UserCreation) (uuid.UUID, error) {
	const op = "UsersRepository.CreateUser"

	rows, err := r.db.QueryxContext(
		ctx,
		CreateUserQuery,
		user.Id,
		user.Username,
		user.Role,
		user.KratosId,
	)
	if err != nil {
		return uuid.Nil, pkg.HandlePgErr(err, op)
	}

	defer rows.Close()
	var id uuid.UUID
	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return uuid.Nil, pkg.HandlePgErr(err, op)
	}

	return id, nil
}

//go:embed sql/get_user_by_id.sql
var GetUserByIdQuery string

func (r *Repository) GetUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "UsersRepository.ReadUserById"

	var user models.User
	err := r.db.GetContext(ctx, &user, GetUserByIdQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	return &user, nil
}

//go:embed sql/get_user_by_kratos_id.sql
var GetUserByKratosIdQuery string

func (r *Repository) GetUserByKratosId(ctx context.Context, kratosId string) (*models.User, error) {
	const op = "UsersRepository.ReadUserByKratosId"

	var user models.User
	err := r.db.GetContext(ctx, &user, GetUserByKratosIdQuery, kratosId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	return &user, nil
}

//go:embed sql/list_users.sql
var ListUsersQuery string

//go:embed sql/count_users.sql
var CountUsersQuery string

func (r *Repository) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int, error) {
	const op = "UsersRepository.ListUsers"

	offset := (page - 1) * pageSize

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, CountUsersQuery)
	if err != nil {
		return nil, 0, pkg.HandlePgErr(err, op)
	}

	// Get users
	var users []*models.User
	err = r.db.SelectContext(ctx, &users, ListUsersQuery, pageSize, offset)
	if err != nil {
		return nil, 0, pkg.HandlePgErr(err, op)
	}

	return users, total, nil
}

//go:embed sql/search_users.sql
var SearchUsersQuery string

//go:embed sql/count_search_users.sql
var CountSearchUsersQuery string

// SearchUsers searches users by search query and role with pagination using PostgreSQL trigram
func (r *Repository) SearchUsers(ctx context.Context, searchQuery, role string, page, pageSize int) ([]*models.User, int, error) {
	const op = "UsersRepository.SearchUsers"

	offset := (page - 1) * pageSize

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, CountSearchUsersQuery, searchQuery, role)
	if err != nil {
		return nil, 0, pkg.HandlePgErr(err, op)
	}

	// Get users
	var users []*models.User
	err = r.db.SelectContext(ctx, &users, SearchUsersQuery, searchQuery, role, pageSize, offset)
	if err != nil {
		return nil, 0, pkg.HandlePgErr(err, op)
	}

	return users, total, nil
}
