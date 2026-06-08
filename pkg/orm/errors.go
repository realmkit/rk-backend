package orm

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrNotFound reports that a requested persistence record does not exist.
var ErrNotFound = errors.New("record not found")

// ErrConflict reports that a persistence operation violated uniqueness.
var ErrConflict = errors.New("record conflict")

// ErrUnavailable reports that a persistence dependency is unavailable.
var ErrUnavailable = errors.New("persistence unavailable")

// PostgresUniqueViolationCode is the PostgreSQL unique violation SQLSTATE.
const PostgresUniqueViolationCode = "23505"

// TranslateError maps common GORM and PostgreSQL errors to GameHub errors.
func TranslateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Join(ErrNotFound, err)
	}
	if isUniqueViolation(err) {
		return errors.Join(ErrConflict, err)
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return errors.Join(ErrUnavailable, err)
	}
	return err
}

// isUniqueViolation reports whether err contains a PostgreSQL unique violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == PostgresUniqueViolationCode
}
