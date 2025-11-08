package pkg

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	NoPermission       = errors.New("no permission")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnhandled       = errors.New("unhandled")
	ErrNotFound        = errors.New("not found")
	ErrBadInput        = errors.New("bad input")
	ErrInternal        = errors.New("internal")
)

type CustomError struct {
	Basic   error  // Basic error (ErrInternal, ErrBadInput etc.) is used to map to HTTP status
	Cause   error  // Extra error (cause) is used for logging
	Op      string // Operation during which the error occurred
	Message string // Message
}

func (e *CustomError) Error() string {
	var parts []string
	if e.Basic != nil {
		parts = append(parts, e.Basic.Error())
	}
	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}
	if e.Op != "" && e.Message != "" {
		parts = append(parts, fmt.Sprintf("during %s: %s", e.Op, e.Message))
	}
	return strings.Join(parts, ": ")
}

func (e *CustomError) Unwrap() []error {
	return []error{e.Basic, e.Cause}
}

func Wrap(basic error, err error, op string, msg string) error {
	return &CustomError{
		Basic:   basic,
		Cause:   err,
		Op:      op,
		Message: msg,
	}
}

func ToREST(err error) int {
	switch {
	case errors.Is(err, ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, ErrBadInput):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInternal):
		return http.StatusInternalServerError
	case errors.Is(err, NoPermission):
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}

func HandlePgErr(err error, op string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return Wrap(ErrBadInput, err, op, "integrity constraint violation")
		}

		return Wrap(ErrUnhandled, err, op, "unexpected postgres error")
	}

	if errors.Is(err, sql.ErrNoRows) {
		return Wrap(ErrNotFound, err, op, "no rows found")
	}

	return Wrap(ErrUnhandled, err, op, "unexpected error")
}
