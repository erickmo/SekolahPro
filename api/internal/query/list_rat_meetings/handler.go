// Package list_rat_meetings menangani query untuk mengambil daftar RAT Meeting.
package list_rat_meetings

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/rat"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_rat_meetings"

// Query berisi parameter untuk mengambil daftar RAT Meeting.
type Query struct {
	Params pagination.ListParams
	Filter rat.ListFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu RAT Meeting dalam hasil list.
type Item struct {
	ID              string `json:"id"`
	MeetingType     string `json:"meeting_type"`
	MeetingNumber   string `json:"meeting_number"`
	Title           string `json:"title"`
	MeetingDate     string `json:"meeting_date"`
	MeetingLocation string `json:"meeting_location"`
	MeetingMode     string `json:"meeting_mode"`
	Status          string `json:"status"`
	QuorumMet       bool   `json:"quorum_met"`
	CreatedAt       string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_rat_meetings.
type Handler struct {
	repo rat.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo rat.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar RAT Meeting dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, qry.Filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list rat meetings: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:              e.ID.String(),
			MeetingType:     e.MeetingType,
			MeetingNumber:   e.MeetingNumber,
			Title:           e.Title,
			MeetingDate:     e.MeetingDate.Format("2006-01-02T15:04:05Z"),
			MeetingLocation: e.MeetingLocation,
			MeetingMode:     e.MeetingMode,
			Status:          e.Status,
			QuorumMet:       e.QuorumMet,
			CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
