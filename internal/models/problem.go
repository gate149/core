package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Meta struct {
	Count int      `json:"count"`
	Names []string `json:"names"` // e.g "01", "02", "03"
}

func (m *Meta) Scan(src interface{}) error {
	if src == nil {
		*m = Meta{}
		return nil
	}

	// Expect src to be []byte (JSONB data)
	data, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte for JSONB, got %T", src)
	}

	// Unmarshal JSON into Meta
	return json.Unmarshal(data, m)
}

type Sample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Samples []Sample

func (s *Samples) Scan(src interface{}) error {
	if src == nil {
		*s = nil
		return nil
	}

	// Expect src to be []byte (JSONB data)
	data, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte for JSONB, got %T", src)
	}

	// Unmarshal JSON into []Sample
	return json.Unmarshal(data, s)
}

type Problem struct {
	Id          uuid.UUID `db:"id"`
	Title       string    `db:"title"`
	TimeLimit   int32     `db:"time_limit"`
	MemoryLimit int32     `db:"memory_limit"`
	IsPrivate   bool      `db:"is_private"`

	Legend       string `db:"legend"`
	InputFormat  string `db:"input_format"`
	OutputFormat string `db:"output_format"`
	Notes        string `db:"notes"`
	Scoring      string `db:"scoring"`

	LegendHtml       string `db:"legend_html"`
	InputFormatHtml  string `db:"input_format_html"`
	OutputFormatHtml string `db:"output_format_html"`
	NotesHtml        string `db:"notes_html"`
	ScoringHtml      string `db:"scoring_html"`

	OwnerId *uuid.UUID `db:"owner_id"` // Problem creator/owner

	Meta    Meta    `db:"meta"`    // JSONB field
	Samples Samples `db:"samples"` // JSONB field

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ProblemsListItem struct {
	Id          uuid.UUID `db:"id"`
	Title       string    `db:"title"`
	MemoryLimit int32     `db:"memory_limit"`
	TimeLimit   int32     `db:"time_limit"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ProblemsList struct {
	Problems   []*ProblemsListItem `json:"problems"`
	Pagination Pagination          `json:"pagination"`
}

type ProblemsFilter struct {
	Page     int32
	PageSize int32
	OwnerId  *uuid.UUID // Filter by owner to get user's private problems
	Title    *string    // Legacy filter for database trigram search
	Search   *string    // Typesense full-text search
	Order    *int32
}

func (f ProblemsFilter) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

type ProblemUpdate struct {
	Title       *string `db:"title"`
	MemoryLimit *int32  `db:"memory_limit"`
	TimeLimit   *int32  `db:"time_limit"`
	IsPrivate   *bool   `db:"is_private"`

	Legend       *string `db:"legend"`
	InputFormat  *string `db:"input_format"`
	OutputFormat *string `db:"output_format"`
	Notes        *string `db:"notes"`
	Scoring      *string `db:"scoring"`

	LegendHtml       *string `db:"legend_html"`
	InputFormatHtml  *string `db:"input_format_html"`
	OutputFormatHtml *string `db:"output_format_html"`
	NotesHtml        *string `db:"notes_html"`
	ScoringHtml      *string `db:"scoring_html"`

	Meta    *Meta     `db:"meta"`    // JSONB field
	Samples *[]Sample `db:"samples"` // JSONB field
}

type ProblemStatement struct {
	Legend       string `db:"legend"`
	InputFormat  string `db:"input_format"`
	OutputFormat string `db:"output_format"`
	Notes        string `db:"notes"`
	Scoring      string `db:"scoring"`
}

type Html5ProblemStatement struct {
	LegendHtml       string `db:"legend_html"`
	InputFormatHtml  string `db:"input_format_html"`
	OutputFormatHtml string `db:"output_format_html"`
	NotesHtml        string `db:"notes_html"`
	ScoringHtml      string `db:"scoring_html"`
}
