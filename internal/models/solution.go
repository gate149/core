package models

import (
	"time"

	"github.com/gate149/core/pkg"
	"github.com/google/uuid"
)

type LanguageName int32

const (
	Golang LanguageName = 10
	Cpp    LanguageName = 20
	Python LanguageName = 30
)

func (n LanguageName) Valid() error {
	const op = "LanguageName.Valid"

	switch n {
	case Golang, Cpp, Python:
		return nil
	default:
		return pkg.Wrap(pkg.ErrBadInput, nil, op, "invalid language")
	}
}

type State int32

const (
	Saved State = 1 // saved to db

	GotCE State = 101 // compilation error
	GotTL State = 102 // time limit exceeded
	GotML State = 103 // memory limit exceeded
	GotRE State = 104 // runtime error
	GotPE State = 105 // presentation error
	GotWA State = 106 // wrong answer

	Accepted State = 200 // accepted
)

type Solution struct {
	Id uuid.UUID `db:"id"`

	UserId   uuid.UUID `db:"user_id"`
	Username string    `db:"username"`

	Solution string `db:"solution"`

	State      State        `db:"state"`
	Score      int32        `db:"score"`
	Penalty    int32        `db:"penalty"`
	TimeStat   int32        `db:"time_stat"`
	MemoryStat int32        `db:"memory_stat"`
	Language   LanguageName `db:"language"`

	ProblemId    uuid.UUID `db:"problem_id"`
	ProblemTitle string    `db:"problem_title"`

	Position int32 `db:"position"`

	ContestId    uuid.UUID `db:"contest_id"`
	ContestTitle string    `db:"contest_title"`

	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

type SolutionUpdate struct {
	State      State
	Score      int32
	TimeStat   int32
	MemoryStat int32
}

type SolutionCreation struct {
	Solution  string
	ProblemId uuid.UUID
	ContestId uuid.UUID
	UserId    uuid.UUID
	Language  LanguageName
	Penalty   int32
}

type SolutionsListItem struct {
	Id uuid.UUID `db:"id"`

	UserId   uuid.UUID `db:"user_id"`
	Username string    `db:"username"`

	State      State        `db:"state"`
	Score      int32        `db:"score"`
	Penalty    int32        `db:"penalty"`
	TimeStat   int32        `db:"time_stat"`
	MemoryStat int32        `db:"memory_stat"`
	Language   LanguageName `db:"language"`

	ProblemId    uuid.UUID `db:"problem_id"`
	ProblemTitle string    `db:"problem_title"`

	Position int32 `db:"position"`

	ContestId    uuid.UUID `db:"contest_id"`
	ContestTitle string    `db:"contest_title"`

	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

type SolutionsList struct {
	Solutions  []*SolutionsListItem
	Pagination Pagination
}

type SolutionsFilter struct {
	Page      int32
	PageSize  int32
	ContestId *uuid.UUID
	UserId    *uuid.UUID
	ProblemId *uuid.UUID
	Language  *LanguageName
	State     *State
	Order     *int32
}

func (f SolutionsFilter) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}
