package solutions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gate149/core/internal/models"
	"github.com/google/uuid"
)

type Publisher interface {
	Publish(subject string, data []byte) error
}

type Repo interface {
	GetSolution(ctx context.Context, id uuid.UUID) (*models.Solution, error)
	CreateSolution(ctx context.Context, creation *models.SolutionCreation) (uuid.UUID, error)
	UpdateSolution(ctx context.Context, id uuid.UUID, update *models.SolutionUpdate) error
	ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error)
}

type ProblemsUC interface {
	GetProblemById(ctx context.Context, id uuid.UUID) (*models.Problem, error)
	DownloadTestsArchive(ctx context.Context, id uuid.UUID) (string, error)
	UnarchiveTestsArchive(ctx context.Context, zipPath, destDirPath string) (string, error)
}

type UseCase struct {
	solutionsRepo Repo
	problemsUC    ProblemsUC
	pub           Publisher
}

func NewUseCase(
	solutionsRepo Repo,
	problemsUC ProblemsUC,
	pub Publisher,
) *UseCase {
	return &UseCase{
		solutionsRepo: solutionsRepo,
		problemsUC:    problemsUC,
		pub:           pub,
	}
}

func (uc *UseCase) GetSolution(ctx context.Context, id uuid.UUID) (*models.Solution, error) {
	return uc.solutionsRepo.GetSolution(ctx, id)
}

func (uc *UseCase) CreateSolution(ctx context.Context, creation *models.SolutionCreation) (uuid.UUID, error) {
	solutionId, err := uc.solutionsRepo.CreateSolution(ctx, creation)
	if err != nil {
		return uuid.Nil, err
	}

	go uc.performTesting(context.Background(), solutionId, creation)

	return solutionId, nil
}

func (uc *UseCase) UpdateSolution(ctx context.Context, id uuid.UUID, update *models.SolutionUpdate) error {
	return uc.solutionsRepo.UpdateSolution(ctx, id, update)
}

func (uc *UseCase) ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error) {
	return uc.solutionsRepo.ListSolutions(ctx, filter)
}

func (uc *UseCase) publish(contestId int32, msg *Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return uc.pub.Publish(fmt.Sprintf("contest-%d-solutions", contestId), b)
}

func (uc *UseCase) performTesting(ctx context.Context, solutionId uuid.UUID, creation *models.SolutionCreation) {
	// FIXME:

	problem, err := uc.problemsUC.GetProblemById(ctx, creation.ProblemId)
	if err != nil {
		return
	}

	if problem.Meta.Count == 0 {
		uc.solutionsRepo.UpdateSolution(ctx, solutionId, &models.SolutionUpdate{
			State:      models.Accepted,
			Score:      100,
			TimeStat:   0,
			MemoryStat: 0,
		})
		return
	}

	archivePath, err := uc.problemsUC.DownloadTestsArchive(ctx, creation.ProblemId)
	if err != nil {
		return
	}

	defer os.Remove(archivePath)

	_, err = uc.problemsUC.UnarchiveTestsArchive(ctx, archivePath, solutionId.String())
	if err != nil {
		return
	}

	return
}

const (
	MessageTypeCreate = "CREATE"
	MessageTypeUpdate = "UPDATE"
	MessageTypeDelete = "DELETE"
)

type SolutionsListItem struct {
	Id int32 `json:"id"`

	UserId   int32  `json:"user_id"`
	Username string `json:"username"`

	State      models.State        `json:"state"`
	Score      int32               `json:"score"`
	Penalty    int32               `json:"penalty"`
	TimeStat   int32               `json:"time_stat"`
	MemoryStat int32               `json:"memory_stat"`
	Language   models.LanguageName `json:"language"`

	ProblemId    int32  `json:"problem_id"`
	ProblemTitle string `json:"problem_title"`

	Position int32 `json:"position"`

	ContestId    int32  `json:"contest_id"`
	ContestTitle string `json:"contest_title"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	MessageType string            `json:"message_type"`
	Message     *string           `json:"message,omitempty"`
	Solution    SolutionsListItem `json:"solution"`
}
