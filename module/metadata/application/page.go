package application

import "github.com/niflaot/gamehub-go/pkg/pagination"

// unlimitedPage returns a broad page for internal owner metadata composition.
func unlimitedPage() pagination.Page {
	return pagination.Page{Limit: pagination.MaxLimit}
}
