package solutions

import (
	"context"

	testerv1 "github.com/gate149/contracts/core/v1"
	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
)

type SolutionsUC interface {
	GetSolution(ctx context.Context, id uuid.UUID) (*models.Solution, error)
	CreateSolution(ctx context.Context, creation *models.SolutionCreation) (uuid.UUID, error)
	UpdateSolution(ctx context.Context, id uuid.UUID, update *models.SolutionUpdate) error
	ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error)
}

type ContestsUC interface {
	CreateContest(ctx context.Context, creation models.ContestCreation) (uuid.UUID, error)
	GetContest(ctx context.Context, id uuid.UUID) (*models.Contest, error)
	ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error)
	UpdateContest(ctx context.Context, id uuid.UUID, contestUpdate models.ContestUpdate) error
	DeleteContest(ctx context.Context, id uuid.UUID) error
	CreateContestProblem(ctx context.Context, contestId, problemId uuid.UUID) error
	GetContestProblem(ctx context.Context, contestId, problemId uuid.UUID) (*models.ContestProblem, error)
	GetContestProblems(ctx context.Context, contestId uuid.UUID) ([]*models.ContestProblemsListItem, error)
	DeleteContestProblem(ctx context.Context, contestId, problemId uuid.UUID) error
	CreateParticipant(ctx context.Context, contestId, userId uuid.UUID) error
	IsParticipant(ctx context.Context, contestId, userId uuid.UUID) (bool, error)
	DeleteParticipant(ctx context.Context, contestId, userId uuid.UUID) error
	//ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error)
	GetMonitor(ctx context.Context, contestId uuid.UUID) (*models.Monitor, error)
}

type PermissionsUC interface {
	CanViewContest(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error)
	CanCreateSolution(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error)
}

type UsersUC interface {
	ReadUserByKratosId(ctx context.Context, kratosId string) (*models.User, error)
}

type SolutionsHandlers struct {
	solutionsUC   SolutionsUC
	contestsUC    ContestsUC
	permissionsUC PermissionsUC
	usersUC       UsersUC
}

func getUserID(h *SolutionsHandlers, c *fiber.Ctx) (uuid.UUID, error) {
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

func NewHandlers(
	solutionsUC SolutionsUC,
	contestsUC ContestsUC,
	permissionsUC PermissionsUC,
	usersUC UsersUC,
) *SolutionsHandlers {
	handlers := &SolutionsHandlers{
		solutionsUC:   solutionsUC,
		contestsUC:    contestsUC,
		permissionsUC: permissionsUC,
		usersUC:       usersUC,
	}

	return handlers
}

const (
	maxSolutionSize int64 = 10 * 1024 * 1024 // 10 MB
)

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

func (h *SolutionsHandlers) CreateSolution(c *fiber.Ctx, params testerv1.CreateSolutionParams) error {
	const op = "SolutionsHandlers.CreateSolution"
	ctx := c.Context()

	userID, err := getUserID(h, c)
	if err != nil {
		return err
	}

	// Get contest to check permissions
	contest, err := h.contestsUC.GetContest(ctx, params.ContestId)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to get contest")
	}

	// Check if user can create solution in this contest
	canCreate, err := h.permissionsUC.CanCreateSolution(ctx, userID, contest)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check create permission")
	}
	if !canCreate {
		return pkg.Wrap(pkg.NoPermission, nil, op, "insufficient permissions to create solution")
	}

	s, err := c.FormFile("solution")
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to get solution file")
	}

	if s.Size == 0 || s.Size > maxSolutionSize {
		return pkg.Wrap(pkg.ErrBadInput, nil, op, "invalid solution size")
	}

	f, err := s.Open()
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to open solution file")
	}
	defer f.Close()

	b := make([]byte, s.Size)
	_, err = f.Read(b)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to read solution file")
	}

	solution := string(b)

	langName := models.LanguageName(params.Language)
	if err := langName.Valid(); err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid language")
	}

	solutionCreation := &models.SolutionCreation{
		UserId:    userID,
		ProblemId: params.ProblemId,
		ContestId: params.ContestId,
		Language:  langName,
		Solution:  solution,
		Penalty:   20,
	}

	solutionID, err := h.solutionsUC.CreateSolution(ctx, solutionCreation)
	if err != nil {
		return err
	}

	return c.JSON(testerv1.CreationResponse{Id: solutionID})
}

func (h *SolutionsHandlers) GetSolution(c *fiber.Ctx, id uuid.UUID) error {
	const op = "SolutionsHandlers.GetSolution"
	ctx := c.Context()

	userID, err := getUserID(h, c)
	if err != nil {
		return err
	}

	solution, err := h.solutionsUC.GetSolution(ctx, id)
	if err != nil {
		return err
	}

	// User can only view their own solution (simplified for now)
	if solution.UserId != userID {
		return pkg.Wrap(pkg.NoPermission, nil, op, "insufficient permissions to view this solution")
	}

	return c.JSON(testerv1.GetSolutionResponse{Solution: SolutionDTO(*solution)})
}

func (h *SolutionsHandlers) ListSolutions(c *fiber.Ctx, params testerv1.ListSolutionsParams) error {
	const op = "SolutionsHandlers.ListSolutions"
	ctx := c.Context()

	userID, err := getUserID(h, c)
	if err != nil {
		return err
	}

	if params.ContestId == nil {
		return pkg.Wrap(pkg.ErrBadInput, nil, op, "contest id is required")
	}

	// Get contest to check permissions
	contest, err := h.contestsUC.GetContest(ctx, *params.ContestId)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to get contest")
	}

	// Check if user can view contest
	canView, err := h.permissionsUC.CanViewContest(ctx, userID, contest)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check contest view permission")
	}
	if !canView {
		return pkg.Wrap(pkg.NoPermission, nil, op, "insufficient permissions to view contest solutions")
	}

	filter := ListSolutionsParamsDTO(params)

	// Users can only view their own solutions (simplified)
	if params.UserId == nil || *params.UserId != userID {
		filter.UserId = &userID
	}

	solutionsList, err := h.solutionsUC.ListSolutions(ctx, filter)
	if err != nil {
		return err
	}

	return c.JSON(ListSolutionsResponseDTO(solutionsList))
}

func ListSolutionsParamsDTO(params testerv1.ListSolutionsParams) models.SolutionsFilter {
	var langName *models.LanguageName = nil
	if params.Language != nil {
		t := models.LanguageName(*params.Language)
		langName = &t
	}

	var state *models.State = nil
	if params.State != nil {
		t := models.State(*params.State)
		state = &t
	}

	return models.SolutionsFilter{
		ContestId: params.ContestId,
		Page:      params.Page,
		PageSize:  params.PageSize,
		// UserId:    params.UserId,
		ProblemId: params.ProblemId,
		Language:  langName,
		Order:     params.Order,
		State:     state,
	}
}

func ListSolutionsResponseDTO(solutionsList *models.SolutionsList) *testerv1.ListSolutionsResponse {
	resp := testerv1.ListSolutionsResponse{
		Solutions:  make([]testerv1.SolutionsListItem, len(solutionsList.Solutions)),
		Pagination: PaginationDTO(solutionsList.Pagination),
	}

	for i, solution := range solutionsList.Solutions {
		resp.Solutions[i] = SolutionsListItemDTO(*solution)
	}

	return &resp
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

func SolutionsListItemDTO(s models.SolutionsListItem) testerv1.SolutionsListItem {
	return testerv1.SolutionsListItem{
		Id: s.Id,

		// UserId:   s.UserId,
		Username: s.Username,

		State:      int32(s.State),
		Score:      s.Score,
		Penalty:    s.Penalty,
		TimeStat:   s.TimeStat,
		MemoryStat: s.MemoryStat,
		Language:   int32(s.Language),

		ProblemId:    s.ProblemId,
		ProblemTitle: s.ProblemTitle,

		Position: s.Position,

		ContestId:    s.ContestId,
		ContestTitle: s.ContestTitle,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func SolutionDTO(s models.Solution) testerv1.Solution {
	return testerv1.Solution{
		Id: s.Id,

		// UserId:   s.UserId,
		Username: s.Username,

		Solution: s.Solution,

		State:      int32(s.State),
		Score:      s.Score,
		Penalty:    s.Penalty,
		TimeStat:   s.TimeStat,
		MemoryStat: s.MemoryStat,
		Language:   int32(s.Language),

		ProblemId:    s.ProblemId,
		ProblemTitle: s.ProblemTitle,

		Position: s.Position,

		ContestId:    s.ContestId,
		ContestTitle: s.ContestTitle,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
