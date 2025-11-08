package permissions

import (
	"context"
	"fmt"

	"github.com/gate149/core/pkg"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	_ "embed"
)

// Resource types
const (
	ResourceContest  = "contest"
	ResourceProblem  = "problem"
	ResourceSolution = "solution"
)

// Relations
const (
	RelationOwner       = "owner"
	RelationModerator   = "moderator"
	RelationParticipant = "participant"
)

// Permission represents a permission record in the database
type Permission struct {
	ID           uuid.UUID `db:"id"`
	ResourceType string    `db:"resource_type"`
	ResourceID   uuid.UUID `db:"resource_id"`
	UserID       uuid.UUID `db:"user_id"`
	Relation     string    `db:"relation"`
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// validateResourceType checks if the resource type is valid
func validateResourceType(resourceType string) error {
	switch resourceType {
	case ResourceContest, ResourceProblem, ResourceSolution:
		return nil
	default:
		return fmt.Errorf("invalid resource type: %s", resourceType)
	}
}

// validateRelation checks if the relation is valid
func validateRelation(relation string) error {
	switch relation {
	case RelationOwner, RelationModerator, RelationParticipant:
		return nil
	default:
		return fmt.Errorf("invalid relation: %s", relation)
	}
}

//go:embed sql/create_permission.sql
var CreatePermissionQuery string

// CreatePermission grants a permission to a user
func (r *Repository) CreatePermission(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	userID uuid.UUID,
	relation string,
) error {
	const op = "Repository.CreatePermission"

	if err := validateResourceType(resourceType); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource type")
	}

	if err := validateRelation(relation); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid relation")
	}

	id := uuid.New()
	_, err := r.db.ExecContext(ctx, CreatePermissionQuery, id, resourceType, resourceID, userID, relation)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/delete_permission.sql
var DeletePermissionQuery string

// DeletePermission revokes a permission from a user
func (r *Repository) DeletePermission(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	userID uuid.UUID,
	relation string,
) error {
	const op = "Repository.DeletePermission"

	if err := validateResourceType(resourceType); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource type")
	}

	if err := validateRelation(relation); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid relation")
	}

	_, err := r.db.ExecContext(ctx, DeletePermissionQuery, resourceType, resourceID, userID, relation)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/has_permission.sql
var HasPermissionQuery string

// HasPermission checks if a user has a specific relation on a resource
func (r *Repository) HasPermission(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	userID uuid.UUID,
	relation string,
) (bool, error) {
	const op = "Repository.HasPermission"

	if err := validateResourceType(resourceType); err != nil {
		return false, pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource type")
	}

	if err := validateRelation(relation); err != nil {
		return false, pkg.Wrap(pkg.ErrBadInput, err, op, "invalid relation")
	}

	var exists bool
	err := r.db.GetContext(ctx, &exists, HasPermissionQuery, resourceType, resourceID, userID, relation)
	if err != nil {
		return false, pkg.HandlePgErr(err, op)
	}

	return exists, nil
}

//go:embed sql/get_user_permissions.sql
var GetUserPermissionsQuery string

// GetUserPermissions returns all relations a user has on a resource
func (r *Repository) GetUserPermissions(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	userID uuid.UUID,
) ([]string, error) {
	const op = "Repository.GetUserPermissions"

	if err := validateResourceType(resourceType); err != nil {
		return nil, pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource type")
	}

	var relations []string
	err := r.db.SelectContext(ctx, &relations, GetUserPermissionsQuery, resourceType, resourceID, userID)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return relations, nil
}

//go:embed sql/has_any_relation.sql
var HasAnyRelationQuery string

// HasAnyRelation checks if a user has any of the specified relations on a resource
func (r *Repository) HasAnyRelation(
	ctx context.Context,
	resourceType string,
	resourceID uuid.UUID,
	userID uuid.UUID,
	relations []string,
) (bool, error) {
	const op = "Repository.HasAnyRelation"

	if err := validateResourceType(resourceType); err != nil {
		return false, pkg.Wrap(pkg.ErrBadInput, err, op, "invalid resource type")
	}

	for _, relation := range relations {
		if err := validateRelation(relation); err != nil {
			return false, pkg.Wrap(pkg.ErrBadInput, err, op, "invalid relation")
		}
	}

	var exists bool
	err := r.db.GetContext(ctx, &exists, HasAnyRelationQuery, resourceType, resourceID, userID, pq.Array(relations))
	if err != nil {
		return false, pkg.HandlePgErr(err, op)
	}

	return exists, nil
}
