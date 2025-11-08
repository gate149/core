package problems

import (
	"context"

	"github.com/gate149/core/pkg"
	"github.com/google/uuid"

	"github.com/gate149/core/internal/models"
	"github.com/jmoiron/sqlx"

	_ "embed"
)

type Repository struct {
	_db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		_db: db,
	}
}

func (r *Repository) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := r._db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *Repository) DB() Querier {
	return r._db
}

//go:embed sql/create_problem.sql
var CreateProblemQuery string

func (r *Repository) CreateProblem(ctx context.Context, q Querier, title string) (uuid.UUID, error) {
	const op = "Repository.CreateProblem"

	rows, err := q.QueryxContext(ctx, CreateProblemQuery, title)
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

//go:embed sql/get_problem_by_id.sql
var GetProblemByIdQuery string

func (r *Repository) GetProblemById(ctx context.Context, q Querier, id uuid.UUID) (*models.Problem, error) {
	const op = "Repository.ReadProblemById"

	var problem models.Problem
	err := q.GetContext(ctx, &problem, GetProblemByIdQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &problem, nil
}

//go:embed sql/delete_problem.sql
var DeleteProblemQuery string

func (r *Repository) DeleteProblem(ctx context.Context, q Querier, id uuid.UUID) error {
	const op = "Repository.DeleteProblem"

	_, err := q.ExecContext(ctx, DeleteProblemQuery, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

//go:embed sql/list_problems.sql
var ListProblemsQuery string

//go:embed sql/count_problems.sql
var CountProblemsQuery string

func (r *Repository) ListProblems(ctx context.Context, q Querier, filter models.ProblemsFilter) (*models.ProblemsList, error) {
	const op = "ContestRepository.ListProblems"

	var title *string
	if filter.Title != nil && *filter.Title != "" {
		title = filter.Title
	}

	var order int = 0
	if filter.Order != nil {
		order = int(*filter.Order)
	}

	// Get count
	var count int32
	err := q.GetContext(ctx, &count, CountProblemsQuery, filter.OwnerId, title)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	// Get problems list
	list := make([]*models.ProblemsListItem, 0)
	err = q.SelectContext(ctx, &list, ListProblemsQuery, filter.OwnerId, title, order, filter.PageSize, filter.Offset())
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.ProblemsList{
		Problems: list,
		Pagination: models.Pagination{
			Total: models.Total(count, filter.PageSize),
			Page:  filter.Page,
		},
	}, nil
}

//go:embed sql/update_problem.sql
var UpdateProblemQuery string

func (r *Repository) UpdateProblem(ctx context.Context, q Querier, id uuid.UUID, problem *models.ProblemUpdate) error {
	const op = "Repository.UpdateProblem"

	_, err := q.ExecContext(ctx, UpdateProblemQuery,
		id,
		problem.Title,
		problem.TimeLimit,
		problem.MemoryLimit,
		problem.IsPrivate,
		problem.Legend,
		problem.InputFormat,
		problem.OutputFormat,
		problem.Notes,
		problem.Scoring,
		problem.LegendHtml,
		problem.InputFormatHtml,
		problem.OutputFormatHtml,
		problem.NotesHtml,
		problem.ScoringHtml,
		problem.Meta,
		problem.Samples,
	)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}
