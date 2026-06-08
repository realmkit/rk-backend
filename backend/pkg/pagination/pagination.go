package pagination

import (
	"errors"
	"fmt"
	"strings"
)

// DefaultLimit is the limit used when requests omit a limit.
const DefaultLimit = 50

// MaxLimit is the maximum accepted page size.
const MaxLimit = 100

// ErrInvalidLimit reports that a pagination limit is invalid.
var ErrInvalidLimit = errors.New("invalid pagination limit")

// Request contains incoming pagination options.
type Request struct {
	// Limit is the requested page size.
	Limit int

	// Cursor is the opaque cursor from a previous page.
	Cursor string
}

// Page contains normalized pagination options.
type Page struct {
	// Limit is the normalized page size.
	Limit int

	// Cursor is the normalized cursor.
	Cursor string
}

// New returns normalized pagination options.
func New(request Request) (Page, error) {
	if request.Limit < 0 {
		return Page{}, fmt.Errorf("%w: %d", ErrInvalidLimit, request.Limit)
	}

	limit := request.Limit
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Page{
		Limit:  limit,
		Cursor: strings.TrimSpace(request.Cursor),
	}, nil
}
