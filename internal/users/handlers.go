package users

import (
	"context"

	testerv1 "github.com/gate149/contracts/core/v1"
	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
)

type UsersUC interface {
	CreateUser(ctx context.Context, user *models.UserCreation) (uuid.UUID, error)
	ReadUserById(ctx context.Context, id uuid.UUID) (*models.User, error)
	ReadUserByKratosId(ctx context.Context, kratosId string) (*models.User, error)
	SearchUsers(ctx context.Context, searchQuery, role string, page, pageSize int) ([]*models.User, int, error)
}

type UsersHandlers struct {
	usersUC UsersUC
}

func NewHandlers(usersUC UsersUC) *UsersHandlers {
	return &UsersHandlers{
		usersUC: usersUC,
	}
}

func getUserFromSession(c *fiber.Ctx) (string, error) {
	session := c.Locals("session")
	if session == nil {
		return "", pkg.Wrap(pkg.ErrUnauthenticated, nil, "", "no session in context")
	}

	s, ok := session.(*ory.Session)
	if !ok {
		return "", pkg.Wrap(pkg.ErrUnauthenticated, nil, "", "invalid session type")
	}

	if !*s.Active {
		return "", pkg.Wrap(pkg.ErrUnauthenticated, nil, "", "session is not active")
	}

	return s.Identity.Id, nil
}

// TODO: Implement UpdateUser and DeleteUser methods when they are added to API
// These methods should check:
// - If userID == targetUserID: allow (user can edit themselves)
// - If userID != targetUserID: only allow if userID is global admin
// Example:
/*
func (h *UsersHandlers) UpdateUser(c *fiber.Ctx, id uuid.UUID) error {
	const op = "UsersHandlers.UpdateUser"
	ctx := c.Context()

	userID, err := getUserFromSession(c)
	if err != nil {
		return err
	}

	targetUserID := id.String()

	// Only global admin can edit others
	if userID != targetUserID {
		isAdmin, err := h.keto.IsGlobalAdmin(ctx, userID)
		if err != nil {
			return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
		}
		if !isAdmin {
			return pkg.Wrap(pkg.NoPermission, nil, op, "only admin can edit other users")
		}
	}

	// ... rest of update logic
}
*/

func (h *UsersHandlers) GetMe(c *fiber.Ctx) error {
	session := c.Locals("session")
	if session == nil {
		return pkg.Wrap(pkg.ErrUnauthenticated, nil, "", "no session in ctx")
	}

	s, ok := session.(*ory.Session)
	if !ok {
		return pkg.Wrap(pkg.ErrUnauthenticated, nil, "", "invalid session type")
	}

	// Use Kratos ID directly from session
	currentUserKratosID := s.Identity.Id

	// Find user by Kratos ID
	user, err := h.usersUC.ReadUserByKratosId(c.Context(), currentUserKratosID)
	if err != nil {
		return err
	}

	return c.JSON(testerv1.GetUserResponse{
		User: UserDTO(*user),
	})
}

func (h *UsersHandlers) GetUser(c *fiber.Ctx, id uuid.UUID) error {
	ctx := c.Context()

	user, err := h.usersUC.ReadUserByKratosId(ctx, id.String())
	if err != nil {
		return err
	}

	userDTO := UserDTO(*user)

	return c.JSON(testerv1.GetUserResponse{
		User: userDTO,
	})
}

func (h *UsersHandlers) GetUsers(c *fiber.Ctx, params testerv1.GetUsersParams) error {
	ctx := c.Context()

	// Extract parameters from the params object
	page := int(params.Page)
	pageSize := int(params.PageSize)

	// Handle optional string pointers
	searchQuery := ""
	if params.Search != nil {
		searchQuery = *params.Search
	}

	role := ""
	if params.Role != nil {
		role = *params.Role
	}

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Search users
	users, total, err := h.usersUC.SearchUsers(ctx, searchQuery, role, page, pageSize)
	if err != nil {
		return err
	}

	// Convert to DTOs
	userDTOs := make([]testerv1.User, len(users))
	for i, user := range users {
		userDTOs[i] = UserDTO(*user)
	}

	return c.JSON(testerv1.ListUsersResponse{
		Users: userDTOs,
		Pagination: testerv1.Pagination{
			Page:  params.Page,
			Total: int32(total),
		},
	})
}

func UserDTO(u models.User) testerv1.User {
	return testerv1.User{
		Id:        u.Id,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
