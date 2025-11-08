package models

import (
	"time"

	"github.com/google/uuid"
)

type Contest struct {
	Id             uuid.UUID `db:"id"`
	Title          string    `db:"title"`
	IsPrivate      bool      `db:"is_private"`
	MonitorEnabled bool      `db:"monitor_enabled"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type ContestCreation struct {
	Title string `json:"title"`
}

type ContestsList struct {
	Contests   []*Contest
	Pagination Pagination
}

type ContestsFilter struct {
	Page       int64
	PageSize   int64
	OwnerId    *uuid.UUID
	Descending bool
	Search     *string
}

func (f ContestsFilter) Offset() int64 {
	return (f.Page - 1) * f.PageSize
}

type ContestUpdate struct {
	Title          *string `json:"title"`
	IsPrivate      *bool   `json:"is_private"`
	MonitorEnabled *bool   `json:"monitor_enabled"`
}

type Monitor struct {
	Participants []*ParticipantsStat
	Summary      []*ProblemStatSummary
}

type ProblemAttempts struct {
	UserId    uuid.UUID `db:"user_id"`
	ProblemId uuid.UUID `db:"problem_id"`
	Position  int32     `db:"position"`
	FAttempts int32     `db:"f_atts"`
	State     *State    `db:"state"`
}

type ParticipantsStat struct {
	UserId   uuid.UUID `db:"user_id"`
	Username string    `db:"username"`
	Solved   int32     `db:"solved_problems"`
	Penalty  int32     `db:"penalty"`
	Attempts []*ProblemAttempts
}

type ProblemStatSummary struct {
	ProblemId   uuid.UUID `db:"problem_id"`
	Position    int32     `db:"position"`
	SAttempts   int32     `db:"s_atts"`
	UnsAttempts int32     `db:"uns_atts"`
	TAttempts   int32     `db:"t_atts"`
}

type ContestProblemsListItem struct {
	ProblemId   uuid.UUID `db:"problem_id"`
	Position    int32     `db:"position"`
	Title       string    `db:"title"`
	TimeLimit   int32     `db:"time_limit"`
	MemoryLimit int32     `db:"memory_limit"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ContestProblem struct {
	ProblemId   uuid.UUID `db:"problem_id"`
	Title       string    `db:"title"`
	TimeLimit   int32     `db:"time_limit"`
	MemoryLimit int32     `db:"memory_limit"`

	Position int32 `db:"position"`

	LegendHtml       string `db:"legend_html"`
	InputFormatHtml  string `db:"input_format_html"`
	OutputFormatHtml string `db:"output_format_html"`
	NotesHtml        string `db:"notes_html"`
	ScoringHtml      string `db:"scoring_html"`

	Meta    Meta    `db:"meta"`    // JSONB field
	Samples Samples `db:"samples"` // JSONB field

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ParticipantsFilter struct {
	Page      int32
	PageSize  int32
	ContestId uuid.UUID
}

func (f ParticipantsFilter) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}
