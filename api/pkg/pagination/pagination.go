package pagination

import (
	"net/http"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 500
)

// ListParams berisi parameter pagination dan sorting.
type ListParams struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
}

// ParseFromRequest membaca parameter limit, offset, sort, order dari HTTP request.
func ParseFromRequest(r *http.Request) ListParams {
	q := r.URL.Query()

	limit := defaultLimit
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
		if l > maxLimit {
			limit = maxLimit
		} else {
			limit = l
		}
	}

	offset := 0
	if o, err := strconv.Atoi(q.Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	sortBy := q.Get("sort")
	order := q.Get("order")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return ListParams{
		Limit:  limit,
		Offset: offset,
		SortBy: sortBy,
		Order:  order,
	}
}
