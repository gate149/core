package contests

import (
	"context"

	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/pkg"
	"github.com/google/uuid"
)

type ContestRepo interface {
	CreateContest(ctx context.Context, creation models.ContestCreation) (uuid.UUID, error)
	GetContest(ctx context.Context, id uuid.UUID) (*models.Contest, error)
	ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error)
	UpdateContest(ctx context.Context, id uuid.UUID, contestUpdate models.ContestUpdate) error
	DeleteContest(ctx context.Context, id uuid.UUID) error

	CreateContestProblem(ctx context.Context, contestId uuid.UUID, problemId uuid.UUID) error
	GetContestProblem(ctx context.Context, contestId uuid.UUID, problemId uuid.UUID) (*models.ContestProblem, error)
	GetContestProblems(ctx context.Context, contestId uuid.UUID) ([]*models.ContestProblemsListItem, error)
	DeleteContestProblem(ctx context.Context, contestId uuid.UUID, problemId uuid.UUID) error

	CreateParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) error
	IsParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) (bool, error)
	DeleteParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) error
	ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error)

	GetMonitor(ctx context.Context, contestId uuid.UUID) (*models.Monitor, error)
}

type UseCase struct {
	contestRepo ContestRepo
}

func NewContestUseCase(
	contestRepo ContestRepo,
) *UseCase {
	return &UseCase{
		contestRepo: contestRepo,
	}
}

func (uc *UseCase) CreateContest(
	ctx context.Context,
	creation models.ContestCreation,
) (uuid.UUID, error) {
	const op = "UseCase.CreateContest"

	id, err := uc.contestRepo.CreateContest(ctx, creation)
	if err != nil {
		return uuid.Nil, pkg.Wrap(nil, err, op, "can't create contest")
	}

	return id, nil
}

func (uc *UseCase) GetContest(ctx context.Context, id uuid.UUID) (*models.Contest, error) {
	return uc.contestRepo.GetContest(ctx, id)
}

func (uc *UseCase) ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error) {
	const op = "UseCase.ListContests"

	contestsList, err := uc.contestRepo.ListContests(ctx, filter)
	if err != nil {
		return nil, pkg.Wrap(nil, err, op, "can't list contests from database")
	}
	return contestsList, nil
}

func (uc *UseCase) UpdateContest(ctx context.Context, id uuid.UUID, contestUpdate models.ContestUpdate) error {
	const op = "UseCase.UpdateContest"

	err := uc.contestRepo.UpdateContest(ctx, id, contestUpdate)
	if err != nil {
		return pkg.Wrap(nil, err, op, "can't update contest")
	}

	return nil
}

func (uc *UseCase) DeleteContest(ctx context.Context, id uuid.UUID) error {
	return uc.contestRepo.DeleteContest(ctx, id)
}

func (uc *UseCase) CreateContestProblem(ctx context.Context, contestId uuid.UUID, problemId uuid.UUID) error {
	return uc.contestRepo.CreateContestProblem(ctx, contestId, problemId)
}

func (uc *UseCase) GetContestProblem(ctx context.Context, contestId uuid.UUID, problemId uuid.UUID) (*models.ContestProblem, error) {
	return uc.contestRepo.GetContestProblem(ctx, contestId, problemId)
}

func (uc *UseCase) GetContestProblems(ctx context.Context, contestId uuid.UUID) ([]*models.ContestProblemsListItem, error) {
	return uc.contestRepo.GetContestProblems(ctx, contestId)
}

func (uc *UseCase) DeleteContestProblem(ctx context.Context, contestId uuid.UUID, problemId uuid.UUID) error {
	return uc.contestRepo.DeleteContestProblem(ctx, contestId, problemId)
}

func (uc *UseCase) CreateParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) error {
	return uc.contestRepo.CreateParticipant(ctx, contestId, userId)
}

func (uc *UseCase) IsParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) (bool, error) {
	return uc.contestRepo.IsParticipant(ctx, contestId, userId)
}

func (uc *UseCase) DeleteParticipant(ctx context.Context, contestId uuid.UUID, userId uuid.UUID) error {
	return uc.contestRepo.DeleteParticipant(ctx, contestId, userId)
}

func (uc *UseCase) ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error) {
	return uc.contestRepo.ListParticipants(ctx, filter)
}

func (uc *UseCase) GetMonitor(ctx context.Context, contestId uuid.UUID) (*models.Monitor, error) {
	return uc.contestRepo.GetMonitor(ctx, contestId)
}
