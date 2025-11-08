package users

import (
	"context"

	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/google/uuid"
)

type Repo interface {
	CreateUser(ctx context.Context, user *models.UserCreation) (uuid.UUID, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByKratosId(ctx context.Context, kratosId string) (*models.User, error)
	ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int, error)
	SearchUsers(ctx context.Context, searchQuery, role string, page, pageSize int) ([]*models.User, int, error)
}

type UsersUseCase struct {
	usersRepo Repo
}

func NewUseCase(usersRepo Repo) *UsersUseCase {
	return &UsersUseCase{
		usersRepo: usersRepo,
	}
}

func (u *UsersUseCase) CreateUser(ctx context.Context, user *models.UserCreation) (uuid.UUID, error) {
	const op = "UseCase.CreateUser"

	id, err := u.usersRepo.CreateUser(ctx, user)
	if err != nil {
		return uuid.Nil, pkg.Wrap(nil, err, op, "can't create user")
	}

	return id, nil
}

func (u *UsersUseCase) ReadUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "UseCase.ReadUserById"

	user, err := u.usersRepo.GetUserById(ctx, id)
	if err != nil {
		return nil, pkg.Wrap(nil, err, op, "can't read user by id")
	}
	return user, nil
}

func (u *UsersUseCase) ReadUserByKratosId(ctx context.Context, kratosId string) (*models.User, error) {
	const op = "UseCase.ReadUserByKratosId"

	user, err := u.usersRepo.GetUserByKratosId(ctx, kratosId)
	if err != nil {
		return nil, pkg.Wrap(nil, err, op, "can't read user by kratos id")
	}
	return user, nil
}

func (u *UsersUseCase) SearchUsers(ctx context.Context, searchQuery, role string, page, pageSize int) ([]*models.User, int, error) {
	const op = "UseCase.SearchUsers"

	users, total, err := u.usersRepo.SearchUsers(ctx, searchQuery, role, page, pageSize)
	if err != nil {
		return nil, 0, pkg.Wrap(nil, err, op, "can't search users in database")
	}
	return users, total, nil
}
