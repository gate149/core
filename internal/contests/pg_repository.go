package contests

import (
	"context"
	"database/sql"
	"errors"
	"sort"

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

//go:embed sql/create_contest.sql
var CreateContestQuery string

func (r *Repository) CreateContest(ctx context.Context, contestCreation models.ContestCreation) (uuid.UUID, error) {
	const op = "Repository.CreateContest"

	rows, err := r.db.QueryxContext(ctx, CreateContestQuery, contestCreation.Title)
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

//go:embed sql/get_contest.sql
var GetContestQuery string

func (r *Repository) GetContest(ctx context.Context, id uuid.UUID) (*models.Contest, error) {
	const op = "Repository.GetContest"

	var contest models.Contest
	err := r.db.GetContext(ctx, &contest, GetContestQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	return &contest, nil
}

//go:embed sql/update_contest.sql
var UpdateContestQuery string

func (r *Repository) UpdateContest(ctx context.Context, id uuid.UUID, contestUpdate models.ContestUpdate) error {
	const op = "Repository.UpdateContest"

	_, err := r.db.ExecContext(ctx, UpdateContestQuery, id,
		contestUpdate.Title,
		contestUpdate.IsPrivate,
		contestUpdate.MonitorEnabled,
	)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/delete_contest.sql
var DeleteContestQuery string

func (r *Repository) DeleteContest(ctx context.Context, id uuid.UUID) error {
	const op = "Repository.DeleteContest"

	_, err := r.db.ExecContext(ctx, DeleteContestQuery, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/list_contests.sql
var ListContestsQuery string

//go:embed sql/count_contests.sql
var CountContestsQuery string

func (r *Repository) ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error) {
	const op = "Repository.ListContests"

	contests := make([]*models.Contest, 0)
	err := r.db.SelectContext(ctx, &contests, ListContestsQuery,
		filter.OwnerId,
		filter.Search,
		filter.Descending,
		filter.PageSize,
		filter.Offset(),
	)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = r.db.GetContext(ctx, &count, CountContestsQuery,
		filter.OwnerId,
		filter.Search,
	)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.ContestsList{
		Contests: contests,
		Pagination: models.Pagination{
			Total: models.Total(count, int32(filter.PageSize)),
			Page:  int32(filter.Page),
		},
	}, nil
}

//go:embed sql/create_contest_problem.sql
var CreateContestProblemQuery string

func (r *Repository) CreateContestProblem(ctx context.Context, contestId, problemId uuid.UUID) error {
	const op = "Repository.CreateContestProblem"

	_, err := r.db.ExecContext(ctx, CreateContestProblemQuery, problemId, contestId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/delete_contest_problem.sql
var DeleteContestProblemQuery string

func (r *Repository) DeleteContestProblem(ctx context.Context, contestId, problemId uuid.UUID) error {
	const op = "Repository.DeleteContestProblem"

	_, err := r.db.ExecContext(ctx, DeleteContestProblemQuery, contestId, problemId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/get_contest_problem.sql
var GetContestProblemQuery string

func (r *Repository) GetContestProblem(ctx context.Context, contestId, problemId uuid.UUID) (*models.ContestProblem, error) {
	const op = "Repository.GetContestProblem"

	var contestProblem models.ContestProblem
	err := r.db.GetContext(ctx, &contestProblem, GetContestProblemQuery, contestId, problemId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &contestProblem, nil
}

//go:embed sql/get_contest_problems.sql
var GetContestProblemsQuery string

func (r *Repository) GetContestProblems(ctx context.Context, contestId uuid.UUID) ([]*models.ContestProblemsListItem, error) {
	const op = "Repository.GetContestProblems"

	contestProblems := make([]*models.ContestProblemsListItem, 0)
	err := r.db.SelectContext(ctx, &contestProblems, GetContestProblemsQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return contestProblems, nil
}

//go:embed sql/create_participant.sql
var CreateParticipantQuery string

func (r *Repository) CreateParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) error {
	const op = "Repository.CreateParticipant"

	_, err := r.db.ExecContext(ctx, CreateParticipantQuery, userId, contestId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/delete_participant.sql
var DeleteParticipantQuery string

func (r *Repository) DeleteParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) error {
	const op = "Repository.DeleteParticipant"

	_, err := r.db.ExecContext(ctx, DeleteParticipantQuery, userId, contestId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}
	return nil
}

var GetParticipantQuery string

func (r *Repository) IsParticipant(ctx context.Context, contestId, userId uuid.UUID) (bool, error) {
	const op = "Repository.IsParticipant"

	var id uuid.UUID
	err := r.db.GetContext(ctx, &id, GetParticipantQuery, userId, contestId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, pkg.HandlePgErr(err, op)
	}

	return true, nil
}

//go:embed sql/list_participants.sql
var ListParticipantsQuery string

//go:embed sql/count_participants.sql
var CountParticipantsQuery string

func (r *Repository) ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error) {
	const op = "Repository.ListParticipants"

	var participants []*models.User
	err := r.db.SelectContext(ctx, &participants,
		ListParticipantsQuery,
		filter.ContestId,
		filter.PageSize,
		filter.Offset(),
	)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = r.db.GetContext(ctx, &count, CountParticipantsQuery, filter.ContestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.UsersList{
		Users: participants,
		Pagination: models.Pagination{
			Total: models.Total(count, filter.PageSize),
			Page:  filter.Page,
		},
	}, nil
}

//go:embed sql/get_monitor_participants.sql
var GetMonitorParticipantsQuery string

//go:embed sql/get_monitor_statistics.sql
var GetMonitorStatisticsQuery string

//go:embed sql/get_monitor_main.sql
var GetMonitorMainQuery string

func (r *Repository) GetMonitor(ctx context.Context, contestId uuid.UUID) (*models.Monitor, error) {
	const op = "Repository.GetMonitor"

	participants := make([]*models.ParticipantsStat, 0)
	err := r.db.SelectContext(ctx, &participants, GetMonitorParticipantsQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	summary := make([]*models.ProblemStatSummary, 0)
	err = r.db.SelectContext(ctx, &summary, GetMonitorStatisticsQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	m := make(map[uuid.UUID][]*models.ProblemAttempts)

	rows, err := r.db.QueryxContext(ctx, GetMonitorMainQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	defer rows.Close()

	for rows.Next() {
		var att models.ProblemAttempts
		err = rows.StructScan(&att)
		if err != nil {
			return nil, pkg.HandlePgErr(err, op)
		}

		if m[att.UserId] == nil {
			m[att.UserId] = make([]*models.ProblemAttempts, 0)
		}

		m[att.UserId] = append(m[att.UserId], &att)
	}

	for _, v := range participants {
		v.Attempts = m[v.UserId]
	}

	sort.Slice(participants, func(i, j int) bool {
		if participants[i].Solved != participants[j].Solved {
			return participants[i].Solved > participants[j].Solved
		}

		return participants[i].Penalty < participants[j].Penalty
	})

	monitor := &models.Monitor{
		Participants: participants,
		Summary:      summary,
	}

	return monitor, nil
}
