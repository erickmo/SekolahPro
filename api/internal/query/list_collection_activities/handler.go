// Package list_collection_activities menangani query untuk mengambil daftar Activity.
package list_collection_activities

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_collection_activities"

// Query berisi parameter untuk mengambil daftar Activity.
type Query struct {
	CaseID       uuid.UUID
	ActivityType *collection.ActivityType
	Limit        int
	Offset       int
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu Activity dalam hasil list.
type Item struct {
	ID               string  `json:"id"`
	CaseID           string  `json:"case_id"`
	ActivityType     string  `json:"activity_type"`
	ActivityDate     string  `json:"activity_date"`
	PerformedBy      string  `json:"performed_by"`
	ContactResult    string  `json:"contact_result"`
	Notes            string  `json:"notes"`
	NasabahResponse  *string `json:"nasabah_response"`
	PromiseAmount    *int64  `json:"promise_amount"`
	PromiseDate      *string `json:"promise_date"`
	FollowupRequired bool    `json:"followup_required"`
	FollowupDate     *string `json:"followup_date"`
	FollowupType     *string `json:"followup_type"`
	CreatedAt        string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_collection_activities.
type Handler struct {
	repo collection.ActivityReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo collection.ActivityReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar Activity berdasarkan case_id dengan filter.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	limit := qry.Limit
	if limit <= 0 {
		limit = 20
	}

	entities, total, err := h.repo.ListActivities(ctx, s, qry.CaseID, qry.ActivityType, limit, qry.Offset)
	if err != nil {
		return nil, fmt.Errorf("list collection activities: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, a := range entities {
		item := &Item{
			ID:               a.ID.String(),
			CaseID:           a.CaseID.String(),
			ActivityType:     string(a.ActivityType),
			ActivityDate:     a.ActivityDate.Format("2006-01-02"),
			PerformedBy:      a.PerformedBy.String(),
			ContactResult:    string(a.ContactResult),
			Notes:            a.Notes,
			FollowupRequired: a.FollowupRequired,
			CreatedAt:        a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if a.NasabahResponse != nil {
			s := string(*a.NasabahResponse)
			item.NasabahResponse = &s
		}
		if a.PromiseAmount != nil {
			item.PromiseAmount = a.PromiseAmount
		}
		if a.PromiseDate != nil {
			s := a.PromiseDate.Format("2006-01-02")
			item.PromiseDate = &s
		}
		if a.FollowupDate != nil {
			s := a.FollowupDate.Format("2006-01-02")
			item.FollowupDate = &s
		}
		if a.FollowupType != nil {
			item.FollowupType = a.FollowupType
		}
		items = append(items, item)
	}

	return &Result{Data: items, Total: total}, nil
}
