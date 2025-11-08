package permissions

import (
	"context"

	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/google/uuid"
)

type PermissionsRepo interface {
	CreatePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error
	DeletePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error
	HasPermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) (bool, error)
	GetUserPermissions(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID) ([]string, error)
	HasAnyRelation(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relations []string) (bool, error)
}

type UsersRepo interface {
	GetUserById(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type ContestsReader interface {
	GetContest(ctx context.Context, id uuid.UUID) (*models.Contest, error)
}

type UseCase struct {
	permissionsRepo PermissionsRepo
	usersRepo       UsersRepo
	contestsReader  ContestsReader
}

func NewUseCase(permissionsRepo PermissionsRepo, usersRepo UsersRepo, contestsReader ContestsReader) *UseCase {
	return &UseCase{
		permissionsRepo: permissionsRepo,
		usersRepo:       usersRepo,
		contestsReader:  contestsReader,
	}
}

// CreatePermission grants a permission to a user
func (uc *UseCase) CreatePermission(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	userID uuid.UUID,
	relation string,
) error {
	return uc.permissionsRepo.CreatePermission(ctx, resourceType, resourceID, userID, relation)
}

// DeletePermission revokes a permission from a user
func (uc *UseCase) DeletePermission(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	userID uuid.UUID,
	relation string,
) error {
	return uc.permissionsRepo.DeletePermission(ctx, resourceType, resourceID, userID, relation)
}

// isGlobalAdmin checks if a user is a global admin
func (uc *UseCase) isGlobalAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := uc.usersRepo.GetUserById(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.IsAdmin(), nil
}

// CanViewContest checks if a user can view a contest
// Public contests can be viewed by any authenticated user
// Private contests require participant || owner || moderator || global admin
func (uc *UseCase) CanViewContest(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error) {
	const op = "UseCase.CanViewContest"

	// Check global admin first
	isAdmin, err := uc.isGlobalAdmin(ctx, userID)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
	}
	if isAdmin {
		return true, nil
	}

	// Check if user is owner via permissions table
	hasAccess, err := uc.permissionsRepo.HasPermission(
		ctx,
		ResourceContest,
		contest.Id,
		userID,
		RelationOwner,
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check owner permission")
	}
	if hasAccess {
		return true, nil
	}

	// If contest is public, any authenticated user can view
	if !contest.IsPrivate {
		return true, nil
	}

	// If contest is private, check ACL for moderator/participant
	hasAccess, err = uc.permissionsRepo.HasAnyRelation(
		ctx,
		ResourceContest,
		contest.Id,
		userID,
		[]string{RelationModerator, RelationParticipant},
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check permissions")
	}

	return hasAccess, nil
}

// CanEditContest checks if a user can edit a contest
// owner || moderator || global admin
func (uc *UseCase) CanEditContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error) {
	const op = "UseCase.CanEditContest"

	// Check global admin first
	isAdmin, err := uc.isGlobalAdmin(ctx, userID)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
	}
	if isAdmin {
		return true, nil
	}

	// Check if user is owner via permissions table
	hasAccess, err := uc.permissionsRepo.HasPermission(
		ctx,
		ResourceContest,
		contestID,
		userID,
		RelationOwner,
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check owner permission")
	}
	if hasAccess {
		return true, nil
	}

	// Check ACL for moderator
	hasAccess, err = uc.permissionsRepo.HasAnyRelation(
		ctx,
		ResourceContest,
		contestID,
		userID,
		[]string{RelationModerator},
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check permissions")
	}

	return hasAccess, nil
}

// CanAdminContest checks if a user has admin rights on a contest
// owner || global admin
func (uc *UseCase) CanAdminContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error) {
	const op = "UseCase.CanAdminContest"

	// Check global admin first
	isAdmin, err := uc.isGlobalAdmin(ctx, userID)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
	}
	if isAdmin {
		return true, nil
	}

	// Check if user is owner via permissions table (only owners can admin)
	hasAccess, err := uc.permissionsRepo.HasPermission(
		ctx,
		ResourceContest,
		contestID,
		userID,
		RelationOwner,
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check owner permission")
	}

	return hasAccess, nil
}

// CanViewOwnSolutions checks if a user can view their own solutions
// Same as CanViewContest
func (uc *UseCase) CanViewOwnSolutions(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error) {
	return uc.CanViewContest(ctx, userID, contest)
}

// CanViewOthersSolutions checks if a user can view others' solutions
// owner || moderator || global admin
func (uc *UseCase) CanViewOthersSolutions(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error) {
	return uc.CanEditContest(ctx, userID, contestID)
}

// CanCreateSolution checks if a user can create a solution
// Same as CanViewContest
func (uc *UseCase) CanCreateSolution(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error) {
	return uc.CanViewContest(ctx, userID, contest)
}

// CanViewMonitor checks if a user can view the contest monitor
// If monitor_enabled: same as view permission
// If monitor_disabled: owner || moderator || global admin
func (uc *UseCase) CanViewMonitor(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error) {
	const op = "UseCase.CanViewMonitor"

	// Check global admin first
	isAdmin, err := uc.isGlobalAdmin(ctx, userID)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
	}
	if isAdmin {
		return true, nil
	}

	// Check if user is owner via permissions table
	hasAccess, err := uc.permissionsRepo.HasPermission(
		ctx,
		ResourceContest,
		contest.Id,
		userID,
		RelationOwner,
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check owner permission")
	}
	if hasAccess {
		return true, nil
	}

	// If monitor is enabled, check view permission
	if contest.MonitorEnabled {
		return uc.CanViewContest(ctx, userID, contest)
	}

	// If monitor is disabled, only moderator can view (owner already checked above)
	hasAccess, err = uc.permissionsRepo.HasAnyRelation(
		ctx,
		ResourceContest,
		contest.Id,
		userID,
		[]string{RelationModerator},
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check permissions")
	}

	return hasAccess, nil
}

// CanViewProblem checks if a user can view a problem
// Public problems can be viewed by any authenticated user
// Private problems require owner || moderator || global admin
func (uc *UseCase) CanViewProblem(ctx context.Context, userID uuid.UUID, problem *models.Problem) (bool, error) {
	const op = "UseCase.CanViewProblem"

	// Check global admin first
	isAdmin, err := uc.isGlobalAdmin(ctx, userID)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
	}
	if isAdmin {
		return true, nil
	}

	// If problem is public, any authenticated user can view
	if !problem.IsPrivate {
		return true, nil
	}

	// If problem is private, check permissions
	hasAccess, err := uc.permissionsRepo.HasAnyRelation(
		ctx,
		ResourceProblem,
		problem.Id,
		userID,
		[]string{RelationOwner, RelationModerator},
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check permissions")
	}

	return hasAccess, nil
}

// CanEditProblem checks if a user can edit a problem
// owner || moderator || global admin
func (uc *UseCase) CanEditProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error) {
	const op = "UseCase.CanEditProblem"

	// Check global admin first
	isAdmin, err := uc.isGlobalAdmin(ctx, userID)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
	}
	if isAdmin {
		return true, nil
	}

	// Check if user is owner or moderator
	hasAccess, err := uc.permissionsRepo.HasAnyRelation(
		ctx,
		ResourceProblem,
		problemID,
		userID,
		[]string{RelationOwner, RelationModerator},
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check permissions")
	}

	return hasAccess, nil
}

// CanAdminProblem checks if a user has admin rights on a problem
// owner || global admin
func (uc *UseCase) CanAdminProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error) {
	const op = "UseCase.CanAdminProblem"

	// Check global admin first
	isAdmin, err := uc.isGlobalAdmin(ctx, userID)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin status")
	}
	if isAdmin {
		return true, nil
	}

	// Check if user is owner
	hasAccess, err := uc.permissionsRepo.HasPermission(
		ctx,
		ResourceProblem,
		problemID,
		userID,
		RelationOwner,
	)
	if err != nil {
		return false, pkg.Wrap(pkg.ErrInternal, err, op, "failed to check permissions")
	}

	return hasAccess, nil
}
