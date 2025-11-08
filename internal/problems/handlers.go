package problems

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	testerv1 "github.com/gate149/contracts/core/v1"
	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/internal/permissions"
	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
)

type UC interface {
	CreateProblem(ctx context.Context, title string) (uuid.UUID, error)
	GetProblemById(ctx context.Context, id uuid.UUID) (*models.Problem, error)
	DownloadTestsArchive(ctx context.Context, id uuid.UUID) (string, error)
	DeleteProblem(ctx context.Context, id uuid.UUID) error
	ListProblems(ctx context.Context, filter models.ProblemsFilter) (*models.ProblemsList, error)
	UpdateProblem(ctx context.Context, id uuid.UUID, problemUpdate *models.ProblemUpdate) error
	UploadProblem(ctx context.Context, id uuid.UUID, r io.ReaderAt, size int64) error
}

type PermissionsUC interface {
	CreatePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error
	CanViewProblem(ctx context.Context, userID uuid.UUID, problem *models.Problem) (bool, error)
	CanEditProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error)
	CanAdminProblem(ctx context.Context, userID uuid.UUID, problemID uuid.UUID) (bool, error)
}

type UsersUC interface {
	ReadUserByKratosId(ctx context.Context, kratosId string) (*models.User, error)
}

type ProblemsHandlers struct {
	problemsUC    UC
	permissionsUC PermissionsUC
	usersUC       UsersUC
	jwtSecret     string
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

// getUserID gets the user's database ID from Kratos session
func (h *ProblemsHandlers) getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	kratosID, err := getUserFromSession(c)
	if err != nil {
		return uuid.Nil, err
	}

	user, err := h.usersUC.ReadUserByKratosId(c.Context(), kratosID)
	if err != nil {
		return uuid.Nil, err
	}

	return user.Id, nil
}

func NewHandlers(problemsUC UC, permissionsUC PermissionsUC, usersUC UsersUC) *ProblemsHandlers {
	return &ProblemsHandlers{
		problemsUC:    problemsUC,
		permissionsUC: permissionsUC,
		usersUC:       usersUC,
	}
}

func (h *ProblemsHandlers) ListProblems(c *fiber.Ctx, params testerv1.ListProblemsParams) error {
	const op = "ProblemsHandlers.ListProblems"
	ctx := c.Context()

	// Get user database ID
	userID, err := h.getUserID(c)
	if err != nil {
		return err
	}

	// Build filter
	filter := models.ProblemsFilter{
		Page:     params.Page,
		PageSize: params.PageSize,
		Title:    params.Title,
		Search:   params.Search,
		Order:    params.Order,
	}

	// Add owner filter if provided (for user's private problems)
	if params.Owner != nil && *params.Owner == "me" {
		filter.OwnerId = &userID
	}

	// List problems
	problemsList, err := h.problemsUC.ListProblems(ctx, filter)
	if err != nil {
		return err
	}

	resp := testerv1.ListProblemsResponse{
		Problems:   make([]testerv1.ProblemsListItem, len(problemsList.Problems)),
		Pagination: PaginationDTO(problemsList.Pagination),
	}

	for i, problem := range problemsList.Problems {
		resp.Problems[i] = ProblemsListItemDTO(*problem)
	}
	return c.JSON(resp)
}

func (h *ProblemsHandlers) CreateProblem(c *fiber.Ctx, params testerv1.CreateProblemParams) error {
	const op = "ProblemsHandlers.CreateProblem"
	ctx := c.Context()

	// Get user database ID
	userID, err := h.getUserID(c)
	if err != nil {
		return err
	}

	if params.Title == "" {
		return pkg.Wrap(pkg.ErrBadInput, nil, op, "empty title")
	}

	// Create problem
	problemID, err := h.problemsUC.CreateProblem(ctx, params.Title)
	if err != nil {
		return err
	}

	// Set user as owner of the problem
	err = h.permissionsUC.CreatePermission(ctx, permissions.ResourceProblem, problemID, userID, permissions.RelationOwner)
	if err != nil {
		// Log error but continue - this is a permission setup step, not critical for creation
		fmt.Printf("Warning: failed to create owner relation: %v\n", err)
	}

	return c.JSON(&testerv1.CreationResponse{Id: problemID})
}

func (h *ProblemsHandlers) DeleteProblem(c *fiber.Ctx, id uuid.UUID) error {
	const op = "ProblemsHandlers.DeleteProblem"
	ctx := c.Context()

	// Get user database ID
	userID, err := h.getUserID(c)
	if err != nil {
		return err
	}

	// Only admin can delete problem
	canAdmin, err := h.permissionsUC.CanAdminProblem(ctx, userID, id)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check admin permission")
	}
	if !canAdmin {
		return pkg.Wrap(pkg.NoPermission, nil, op, "insufficient permissions to delete problem")
	}

	err = h.problemsUC.DeleteProblem(ctx, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ProblemsHandlers) GetProblem(c *fiber.Ctx, id uuid.UUID) error {
	const op = "ProblemsHandlers.GetProblem"
	ctx := c.Context()

	// Get user database ID
	userID, err := h.getUserID(c)
	if err != nil {
		return err
	}

	// Get problem
	problem, err := h.problemsUC.GetProblemById(ctx, id)
	if err != nil {
		return err
	}

	// Check if user can view this problem (considering is_private)
	canView, err := h.permissionsUC.CanViewProblem(ctx, userID, problem)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check view permission")
	}
	if !canView {
		return pkg.Wrap(pkg.NoPermission, nil, op, "cannot view this problem")
	}

	return c.JSON(
		testerv1.GetProblemResponse{Problem: *ProblemDTO(problem)},
	)
}

func (h *ProblemsHandlers) UpdateProblem(c *fiber.Ctx, id uuid.UUID) error {
	const op = "ProblemsHandlers.UpdateProblem"
	ctx := c.Context()

	// Get logger safely
	var logger *slog.Logger
	loggerVal := c.Locals("logger")
	if loggerVal != nil {
		if l, ok := loggerVal.(*slog.Logger); ok {
			logger = l
		}
	}

	// Log at the very start
	if logger != nil {
		logger.Info("UpdateProblem handler called", slog.String("problem_id", id.String()))
	} else {
		fmt.Printf("UpdateProblem handler called for problem_id: %s\n", id.String())
	}

	// Get user database ID
	userID, err := h.getUserID(c)
	if err != nil {
		if logger != nil {
			logger.Error("failed to get user ID", slog.Any("error", err))
		}
		return err
	}

	// Check if user can edit this problem
	canEdit, err := h.permissionsUC.CanEditProblem(ctx, userID, id)
	if err != nil {
		if logger != nil {
			logger.Error("failed to check edit permission", slog.Any("error", err))
		}
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check edit permission")
	}
	if !canEdit {
		if logger != nil {
			logger.Info("user does not have edit permission", slog.String("user_id", userID.String()))
		}
		return pkg.Wrap(pkg.NoPermission, nil, op, "insufficient permissions to update problem")
	}

	var req testerv1.UpdateProblemRequest

	// Log raw request before parsing
	fmt.Printf("🔍 UpdateProblem raw request:\n")
	fmt.Printf("   Content-Type: %s\n", c.Get("Content-Type"))
	fmt.Printf("   Body length: %d\n", len(c.Body()))
	fmt.Printf("   Body: %s\n", string(c.Body()))

	err = c.BodyParser(&req)
	if err != nil {
		// Log the raw body for debugging
		body := c.Body()
		if logger != nil {
			logger.Error("failed to parse update problem request body",
				slog.String("op", op),
				slog.String("raw_body", string(body)),
				slog.String("content_type", c.Get("Content-Type")),
				slog.Any("error", err),
			)
		} else {
			fmt.Printf("Failed to parse body: %v\nRaw body: %s\nContent-Type: %s\n", err, string(body), c.Get("Content-Type"))
		}
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to parse request body")
	}

	err = h.problemsUC.UpdateProblem(ctx, id, &models.ProblemUpdate{
		Title:       req.Title,
		MemoryLimit: req.MemoryLimit,
		TimeLimit:   req.TimeLimit,
		IsPrivate:   req.IsPrivate,

		Legend:       req.Legend,
		InputFormat:  req.InputFormat,
		OutputFormat: req.OutputFormat,
		Notes:        req.Notes,
		Scoring:      req.Scoring,
	})

	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ProblemsHandlers) UploadProblem(c *fiber.Ctx, id uuid.UUID) error {
	const op = "ProblemsHandlers.UploadProblem"
	ctx := c.Context()

	// Get user database ID
	userID, err := h.getUserID(c)
	if err != nil {
		return err
	}

	canEdit, err := h.permissionsUC.CanEditProblem(ctx, userID, id)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check edit permission")
	}
	if !canEdit {
		return pkg.Wrap(pkg.NoPermission, nil, op, "insufficient permissions to upload problem archive")
	}

	a, err := c.FormFile("archive")
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "no archive uploaded")
	}

	if a.Size == 0 || a.Size > 1024*1024*1024 {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid archive size")
	}

	f, err := a.Open()
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to open archive")
	}
	defer f.Close()

	err = h.problemsUC.UploadProblem(ctx, id, f, a.Size)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

func ProblemsListItemDTO(p models.ProblemsListItem) testerv1.ProblemsListItem {
	return testerv1.ProblemsListItem{
		Id:          p.Id,
		Title:       p.Title,
		MemoryLimit: p.MemoryLimit,
		TimeLimit:   p.TimeLimit,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ProblemDTO(p *models.Problem) *testerv1.Problem {
	return &testerv1.Problem{
		Id:          p.Id,
		Title:       p.Title,
		TimeLimit:   p.TimeLimit,
		MemoryLimit: p.MemoryLimit,

		Legend:       p.Legend,
		InputFormat:  p.InputFormat,
		OutputFormat: p.OutputFormat,
		Notes:        p.Notes,
		Scoring:      p.Scoring,

		LegendHtml:       p.LegendHtml,
		InputFormatHtml:  p.InputFormatHtml,
		OutputFormatHtml: p.OutputFormatHtml,
		NotesHtml:        p.NotesHtml,
		ScoringHtml:      p.ScoringHtml,

		//Meta:    MetaDTO(p.Meta),
		//Samples: SamplesDTO(p.Samples),

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

//func MetaDTO(m models.Meta) testerv1.Meta {
//	return testerv1.Meta{
//		Author: m.Author,
//	}
//}
