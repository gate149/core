package permissions

import (
	"context"

	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
)

type PermissionsUC interface {
	CreatePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error
	DeletePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error
	CanViewContest(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error)
	CanEditContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error)
	CanAdminContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error)
	CanViewMonitor(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error)
	CanViewProblem(ctx context.Context, userID uuid.UUID, problem *models.Problem) (bool, error)
	CanEditProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error)
	CanAdminProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error)
}

type Handlers struct {
	permissionsUC PermissionsUC
}

func NewHandlers(permissionsUC PermissionsUC) *Handlers {
	return &Handlers{
		permissionsUC: permissionsUC,
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

type CheckPermissionRequest struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Permission   string `json:"permission"`
}

type CheckPermissionResponse struct {
	Allowed bool `json:"allowed"`
}

// CheckPermission is an endpoint for frontend to check if user has permission
// POST /permissions/check
func (h *Handlers) CheckPermission(c *fiber.Ctx) error {
	const op = "PermissionsHandlers.CheckPermission"
	ctx := c.Context()

	// Get user from session
	userIDStr, err := getUserFromSession(c)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid user ID")
	}

	// Parse request body
	var req CheckPermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to parse request body")
	}

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource ID")
	}

	// Check permission based on resource type and permission
	var allowed bool
	switch req.ResourceType {
	case ResourceContest:
		switch req.Permission {
		case "view":
			// Need to get contest to check if it's public
			// For now, just check edit permission
			allowed, err = h.permissionsUC.CanEditContest(ctx, userID, resourceID)
		case "edit":
			allowed, err = h.permissionsUC.CanEditContest(ctx, userID, resourceID)
		case "admin":
			allowed, err = h.permissionsUC.CanAdminContest(ctx, userID, resourceID)
		default:
			return pkg.Wrap(pkg.ErrBadInput, nil, op, "invalid permission")
		}
	case ResourceProblem:
		switch req.Permission {
		case "view":
			// Need to get problem to check if it's public
			// For now, just check edit permission
			allowed, err = h.permissionsUC.CanEditProblem(ctx, userID, resourceID)
		case "edit":
			allowed, err = h.permissionsUC.CanEditProblem(ctx, userID, resourceID)
		case "admin":
			allowed, err = h.permissionsUC.CanAdminProblem(ctx, userID, resourceID)
		default:
			return pkg.Wrap(pkg.ErrBadInput, nil, op, "invalid permission")
		}
	default:
		return pkg.Wrap(pkg.ErrBadInput, nil, op, "invalid resource type")
	}

	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check permission")
	}

	return c.JSON(CheckPermissionResponse{Allowed: allowed})
}

// GrantPermissionRequest is the request body for granting permissions
type GrantPermissionRequest struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	UserID       string `json:"user_id"`
	Relation     string `json:"relation"`
}

// GrantPermission is an admin-only endpoint to grant permissions
// POST /permissions/grant
func (h *Handlers) GrantPermission(c *fiber.Ctx) error {
	const op = "PermissionsHandlers.GrantPermission"
	ctx := c.Context()

	// Get user from session
	_, err := getUserFromSession(c)
	if err != nil {
		return err
	}

	// TODO: Check if user is admin

	// Parse request body
	var req GrantPermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to parse request body")
	}

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource ID")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid user ID")
	}

	err = h.permissionsUC.CreatePermission(ctx, req.ResourceType, resourceID, userID, req.Relation)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to grant permission")
	}

	return c.SendStatus(fiber.StatusOK)
}

// RevokePermissionRequest is the request body for revoking permissions
type RevokePermissionRequest struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	UserID       string `json:"user_id"`
	Relation     string `json:"relation"`
}

// RevokePermission is an admin-only endpoint to revoke permissions
// POST /permissions/revoke
func (h *Handlers) RevokePermission(c *fiber.Ctx) error {
	const op = "PermissionsHandlers.RevokePermission"
	ctx := c.Context()

	// Get user from session
	_, err := getUserFromSession(c)
	if err != nil {
		return err
	}

	// TODO: Check if user is admin

	// Parse request body
	var req RevokePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to parse request body")
	}

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource ID")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid user ID")
	}

	err = h.permissionsUC.DeletePermission(ctx, req.ResourceType, resourceID, userID, req.Relation)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to revoke permission")
	}

	return c.SendStatus(fiber.StatusOK)
}
