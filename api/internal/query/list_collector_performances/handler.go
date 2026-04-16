// Package list_collector_performances menangani query untuk mengambil daftar Performance.
package list_collector_performances

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_collector_performances"

// Query berisi parameter untuk mengambil daftar Performance.
type Query struct {
	CollectorID *uuid.UUID
	PeriodMonth string
	Limit       int
	Offset      int
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu Performance dalam hasil list.
type Item struct {
	ID                   string  `json:"id"`
	CollectorID          string  `json:"collector_id"`
	PeriodMonth          string  `json:"period_month"`
	TotalCalls           int     `json:"total_calls"`
	SuccessfulContacts   int     `json:"successful_contacts"`
	TotalVisits          int     `json:"total_visits"`
	SuccessfulVisits     int     `json:"successful_visits"`
	CasesHandled         int     `json:"cases_handled"`
	CasesResolved        int     `json:"cases_resolved"`
	TotalAmountCollected int64   `json:"total_amount_collected"`
	PromiseToPayCount    int     `json:"promise_to_pay_count"`
	PromiseKeptCount     int     `json:"promise_kept_count"`
	ContactRatePct       float64 `json:"contact_rate_pct"`
	ResolutionRatePct    float64 `json:"resolution_rate_pct"`
	CollectionRatePct    float64 `json:"collection_rate_pct"`
	PromiseKeptRatePct   float64 `json:"promise_kept_rate_pct"`
	CalculatedAt         string  `json:"calculated_at"`
	CreatedAt            string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_collector_performances.
type Handler struct {
	repo collection.PerformanceReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo collection.PerformanceReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar Performance dengan filter dan pagination.
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

	entities, total, err := h.repo.ListPerformances(ctx, s, qry.CollectorID, qry.PeriodMonth, limit, qry.Offset)
	if err != nil {
		return nil, fmt.Errorf("list collector performances: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, p := range entities {
		items = append(items, &Item{
			ID:                   p.ID.String(),
			CollectorID:          p.CollectorID.String(),
			PeriodMonth:          p.PeriodMonth,
			TotalCalls:           p.TotalCalls,
			SuccessfulContacts:   p.SuccessfulContacts,
			TotalVisits:          p.TotalVisits,
			SuccessfulVisits:     p.SuccessfulVisits,
			CasesHandled:         p.CasesHandled,
			CasesResolved:        p.CasesResolved,
			TotalAmountCollected: p.TotalAmountCollected,
			PromiseToPayCount:    p.PromiseToPayCount,
			PromiseKeptCount:     p.PromiseKeptCount,
			ContactRatePct:       p.ContactRatePct,
			ResolutionRatePct:    p.ResolutionRatePct,
			CollectionRatePct:    p.CollectionRatePct,
			PromiseKeptRatePct:   p.PromiseKeptRatePct,
			CalculatedAt:         p.CalculatedAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:            p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
