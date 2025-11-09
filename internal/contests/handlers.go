package contests

import (
	"context"
	"io"
	"unicode/utf8"

	testerv1 "github.com/gate149/contracts/core/v1"
	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/internal/permissions"
	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
)

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
	DeleteParticipant(ctx context.Context, contestId, userId uuid.UUID) error
	ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error)

	GetMonitor(ctx context.Context, contestId uuid.UUID) (*models.Monitor, error)
}

type ProblemsUC interface {
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
	DeletePermission(ctx context.Context, resourceType string, resourceID uuid.UUID, userID uuid.UUID, relation string) error
	CanViewContest(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error)
	CanEditContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error)
	CanAdminContest(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (bool, error)
	CanViewMonitor(ctx context.Context, userID uuid.UUID, contest *models.Contest) (bool, error)
}

type UsersUC interface {
	ReadUserByKratosId(ctx context.Context, kratosId string) (*models.User, error)
}

type ContestsHandlers struct {
	problemsUC    ProblemsUC
	contestsUC    ContestsUC
	permissionsUC PermissionsUC
	usersUC       UsersUC
}

func NewHandlers(problemsUC ProblemsUC, contestsUC ContestsUC, permissionsUC PermissionsUC, usersUC UsersUC) *ContestsHandlers {
	return &ContestsHandlers{
		problemsUC:    problemsUC,
		contestsUC:    contestsUC,
		permissionsUC: permissionsUC,
		usersUC:       usersUC,
	}
}

const sessionKey = "session"

func getKratosId(c *fiber.Ctx) (string, error) {
	session := c.Locals(sessionKey)
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

func (h *ContestsHandlers) getUser(c *fiber.Ctx) (*models.User, error) {
	kratosID, err := getKratosId(c)
	if err != nil {
		return nil, err
	}

	user, err := h.usersUC.ReadUserByKratosId(c.Context(), kratosID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func checkPermission(f func() (bool, error)) error {
	can, err := f()
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, "", "failed to check permission")
	}
	if !can {
		return pkg.Wrap(pkg.NoPermission, nil, "", "insufficient permission")
	}
	return nil
}

func validateCreateContestParams(params testerv1.CreateContestParams) error {
	if params.Title == "" {
		return pkg.Wrap(pkg.ErrBadInput, nil, "", "empty title")
	}

	titleLength := utf8.RuneCountInString(params.Title)
	if titleLength < 3 || titleLength > 64 {
		return pkg.Wrap(pkg.ErrBadInput, nil, "", "title must be between 3 and 64 characters")
	}

	return nil
}

func (h *ContestsHandlers) CreateContest(c *fiber.Ctx, params testerv1.CreateContestParams) error {
	const op = "ContestsHandlers.CreateContest"

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	err = validateCreateContestParams(params)
	if err != nil {
		return err
	}

	contestCreation := models.ContestCreation{
		Title: params.Title,
	}

	contestID, err := h.contestsUC.CreateContest(c.Context(), contestCreation)
	if err != nil {
		return err
	}

	// Create owner permission for the user who created the contest
	err = h.permissionsUC.CreatePermission(
		c.Context(),
		permissions.ResourceContest,
		contestID,
		user.Id,
		permissions.RelationOwner,
	)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to create owner permission")
	}

	return c.JSON(&testerv1.CreationResponse{Id: contestID})
}

func (h *ContestsHandlers) GetContest(c *fiber.Ctx, id uuid.UUID) error {
	const op = "ContestsHandlers.GetContest"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	contest, err := h.contestsUC.GetContest(ctx, id)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanViewContest(ctx, user.Id, contest)
	})
	if err != nil {
		return err
	}

	ps, err := h.contestsUC.GetContestProblems(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(GetContestResponseDTO(contest, ps))
}

func validateUpdateContestRequest(params testerv1.UpdateContestRequest) error {
	if params.Title != nil {
		titleLength := utf8.RuneCountInString(*params.Title)
		if titleLength < 3 || titleLength > 64 {
			return pkg.Wrap(pkg.ErrBadInput, nil, "", "title must be between 3 and 64 characters")
		}
	}

	return nil
}

func (h *ContestsHandlers) UpdateContest(c *fiber.Ctx, id uuid.UUID) error {
	const op = "ContestsHandlers.UpdateContest"
	ctx := c.Context()

	var req testerv1.UpdateContestRequest
	err := c.BodyParser(&req)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to parse request body")
	}

	err = validateUpdateContestRequest(req)
	if err != nil {
		return err
	}

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanEditContest(ctx, user.Id, id)
	})
	if err != nil {
		return err
	}

	err = h.contestsUC.UpdateContest(ctx, id, models.ContestUpdate{
		Title:          req.Title,
		IsPrivate:      req.IsPrivate,
		MonitorEnabled: req.MonitorEnabled,
	})
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ContestsHandlers) DeleteContest(c *fiber.Ctx, id uuid.UUID) error {
	const op = "ContestsHandlers.DeleteContest"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanAdminContest(ctx, user.Id, id)
	})
	if err != nil {
		return err
	}

	err = h.contestsUC.DeleteContest(ctx, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func validateListContestsParams(params testerv1.ListContestsParams) error {
	if params.Page < 1 {
		return pkg.Wrap(pkg.ErrBadInput, nil, "", "page must be greater than 0")
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		return pkg.Wrap(pkg.ErrBadInput, nil, "", "page size must be between 1 and 100")
	}

	if params.Owner != nil && *params.Owner != "me" {
		return pkg.Wrap(pkg.ErrBadInput, nil, "", "only 'me' is supported for owner")
	}

	if params.Title != nil {
		titleLength := utf8.RuneCountInString(*params.Title)
		if titleLength > 64 {
			return pkg.Wrap(pkg.ErrBadInput, nil, "", "title length must be less than 64 characters")
		}
	}

	return nil
}

func (h *ContestsHandlers) ListContests(c *fiber.Ctx, params testerv1.ListContestsParams) error {
	const op = "ContestsHandlers.ListContests"
	ctx := c.Context()

	err := validateListContestsParams(params)
	if err != nil {
		return err
	}

	filter := models.ContestsFilter{
		Page:     int64(params.Page),
		PageSize: int64(params.PageSize),
	}

	// Add search filter if provided
	if params.Title != nil && *params.Title != "" {
		filter.Search = params.Title
	}

	// Add owner filter if provided (for user's private contests)
	if params.Owner != nil {
		// For owner filter, we need authenticated user
		user, err := h.getUser(c)
		if err != nil {
			return err
		}
		filter.OwnerId = &user.Id
	}

	// Add descending sort order if provided (default is false = ASC)
	if params.Descending != nil {
		filter.Descending = *params.Descending
	}

	contestsList, err := h.contestsUC.ListContests(ctx, filter)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to list contests")
	}

	return c.JSON(ListContestsResponseDTO(contestsList))
}

func (h *ContestsHandlers) CreateContestProblem(c *fiber.Ctx, contestId uuid.UUID, params testerv1.CreateContestProblemParams) error {
	const op = "ContestsHandlers.CreateContestProblem"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanEditContest(ctx, user.Id, contestId)
	})
	if err != nil {
		return err
	}

	err = h.contestsUC.CreateContestProblem(ctx, contestId, params.ProblemId)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ContestsHandlers) GetContestProblem(c *fiber.Ctx, contestId uuid.UUID, problemId uuid.UUID) error {
	const op = "ContestsHandlers.GetContestProblem"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	// Get contest to check permissions
	contest, err := h.contestsUC.GetContest(ctx, contestId)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanViewContest(ctx, user.Id, contest)
	})
	if err != nil {
		return err
	}

	p, err := h.contestsUC.GetContestProblem(ctx, contestId, problemId)
	if err != nil {
		return err
	}

	return c.JSON(GetContestProblemResponseDTO(p))
}

func (h *ContestsHandlers) DeleteContestProblem(c *fiber.Ctx, contestId uuid.UUID, problemId uuid.UUID) error {
	const op = "ContestsHandlers.DeleteContestProblem"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	canEdit, err := h.permissionsUC.CanEditContest(ctx, user.Id, contestId)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to check edit permission")
	}
	if !canEdit {
		return pkg.Wrap(pkg.NoPermission, nil, op, "insufficient permissions to remove problem from contest")
	}

	err = h.contestsUC.DeleteContestProblem(ctx, contestId, problemId)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ContestsHandlers) CreateParticipant(c *fiber.Ctx, contestId uuid.UUID, params testerv1.CreateParticipantParams) error {
	const op = "ContestsHandlers.CreateParticipant"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanEditContest(ctx, user.Id, contestId)
	})
	if err != nil {
		return err
	}

	err = h.permissionsUC.CreatePermission(ctx, permissions.ResourceContest, contestId, params.UserId, permissions.RelationParticipant)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to create participant permission")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ContestsHandlers) DeleteParticipant(c *fiber.Ctx, contestId uuid.UUID, params testerv1.DeleteParticipantParams) error {
	const op = "ContestsHandlers.DeleteParticipant"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanEditContest(ctx, user.Id, contestId)
	})
	if err != nil {
		return err
	}

	err = h.permissionsUC.DeletePermission(ctx, permissions.ResourceContest, contestId, params.UserId, permissions.RelationParticipant)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to delete participant permission")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ContestsHandlers) ListParticipants(c *fiber.Ctx, contestId uuid.UUID, params testerv1.ListParticipantsParams) error {
	const op = "ContestsHandlers.ListParticipants"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	contest, err := h.contestsUC.GetContest(ctx, contestId)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanViewContest(ctx, user.Id, contest)
	})
	if err != nil {
		return err
	}

	participantsList, err := h.contestsUC.ListParticipants(ctx, models.ParticipantsFilter{
		Page:      params.Page,
		PageSize:  params.PageSize,
		ContestId: contestId,
	})
	if err != nil {
		return err
	}

	resp := testerv1.ListUsersResponse{
		Users:      make([]testerv1.User, len(participantsList.Users)),
		Pagination: PaginationDTO(participantsList.Pagination),
	}

	for i, user := range participantsList.Users {
		resp.Users[i] = UserDTO(*user)
	}

	return c.JSON(resp)
}

func (h *ContestsHandlers) GetMonitor(c *fiber.Ctx, contestId uuid.UUID) error {
	const op = "ContestsHandlers.GetMonitor"
	ctx := c.Context()

	user, err := h.getUser(c)
	if err != nil {
		return err
	}

	contest, err := h.contestsUC.GetContest(ctx, contestId)
	if err != nil {
		return err
	}

	err = checkPermission(func() (bool, error) {
		return h.permissionsUC.CanViewMonitor(ctx, user.Id, contest)
	})
	if err != nil {
		return err
	}

	monitor, err := h.contestsUC.GetMonitor(ctx, contestId)
	if err != nil {
		return err
	}
	return c.JSON(GetMonitorResponseDTO(monitor))
}

func GetContestResponseDTO(contest *models.Contest, problems []*models.ContestProblemsListItem) *testerv1.GetContestResponse {
	resp := testerv1.GetContestResponse{
		Contest:  ContestDTO(*contest),
		Problems: make([]testerv1.ContestProblemListItem, len(problems)),
	}

	for i, task := range problems {
		resp.Problems[i] = ContestProblemsListItemDTO(*task)
	}

	return &resp
}

func ListContestsResponseDTO(contestsList *models.ContestsList) *testerv1.ListContestsResponse {
	resp := testerv1.ListContestsResponse{
		Contests:   make([]testerv1.Contest, len(contestsList.Contests)),
		Pagination: PaginationDTO(contestsList.Pagination),
	}

	for i, contest := range contestsList.Contests {
		resp.Contests[i] = ContestDTO(*contest)
	}

	return &resp
}

func GetContestProblemResponseDTO(p *models.ContestProblem) *testerv1.GetContestProblemResponse {
	resp := testerv1.GetContestProblemResponse{
		Problem: testerv1.ContestProblem{
			ProblemId:   p.ProblemId,
			Title:       p.Title,
			MemoryLimit: p.MemoryLimit,
			TimeLimit:   p.TimeLimit,

			Position: p.Position,

			LegendHtml:       p.LegendHtml,
			InputFormatHtml:  p.InputFormatHtml,
			OutputFormatHtml: p.OutputFormatHtml,
			NotesHtml:        p.NotesHtml,
			ScoringHtml:      p.ScoringHtml,

			//Meta:             MetaDTO(p.Meta),
			//Samples:          SamplesDTO(p.Samples),

			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
	}

	return &resp
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

func ContestDTO(c models.Contest) testerv1.Contest {
	return testerv1.Contest{
		Id:        c.Id,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ContestProblemsListItemDTO(t models.ContestProblemsListItem) testerv1.ContestProblemListItem {
	return testerv1.ContestProblemListItem{
		ProblemId:   t.ProblemId,
		Position:    t.Position,
		Title:       t.Title,
		MemoryLimit: t.MemoryLimit,
		TimeLimit:   t.TimeLimit,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
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

func GetMonitorResponseDTO(m *models.Monitor) testerv1.GetMonitorResponse {
	resp := testerv1.GetMonitorResponse{
		Participants: make([]testerv1.ParticipantsStat, len(m.Participants)),
		Summary:      make([]testerv1.ProblemStatSummary, len(m.Summary)),
	}

	ProblemAttemptsDTO := func(p *models.ProblemAttempts) testerv1.ProblemAttempts {
		return testerv1.ProblemAttempts{
			ProblemId:      p.ProblemId,
			Position:       p.Position,
			State:          stateP(p.State),
			FailedAttempts: p.FAttempts,
		}
	}

	ParticipantsStatDTO := func(p models.ParticipantsStat) testerv1.ParticipantsStat {
		s := testerv1.ParticipantsStat{
			// UserId:   p.UserId,
			Username: p.Username,
			Solved:   p.Solved,
			Penalty:  p.Penalty,
			Attempts: make([]testerv1.ProblemAttempts, len(p.Attempts)),
		}

		for i, attempt := range p.Attempts {
			s.Attempts[i] = ProblemAttemptsDTO(attempt)
		}

		return s
	}

	ProblemStatSummaryDTO := func(p models.ProblemStatSummary) testerv1.ProblemStatSummary {
		return testerv1.ProblemStatSummary{
			ProblemId: p.ProblemId,
			Position:  p.Position,
			SAttempts: p.SAttempts,
			FAttempts: p.UnsAttempts,
			TAttempts: p.TAttempts,
		}
	}

	for i, user := range m.Participants {
		resp.Participants[i] = ParticipantsStatDTO(*user)
	}

	for i, summary := range m.Summary {
		resp.Summary[i] = ProblemStatSummaryDTO(*summary)
	}

	return resp
}

func stateP(s *models.State) *int32 {
	if s == nil {
		return nil
	}
	return (*int32)(s)
}
