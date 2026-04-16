// Package record_collector_performance menangani command untuk merekam performa collector.
package record_collector_performance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "record_collector_performance"

// Command berisi data untuk merekam atau mengupdate performa collector.
type Command struct {
	CollectorID          uuid.UUID
	PeriodMonth          string
	TotalCalls           int
	SuccessfulContacts   int
	TotalVisits          int
	SuccessfulVisits     int
	CasesHandled         int
	CasesResolved        int
	TotalAmountCollected int64
	PromiseToPayCount    int
	PromiseKeptCount     int
	ContactRatePct       float64
	ResolutionRatePct    float64
	CollectionRatePct    float64
	PromiseKeptRatePct   float64
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command record_collector_performance.
type Handler struct {
	repo collection.PerformanceWriteRepository
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo collection.PerformanceWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle memvalidasi command, membuat entity Performance, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.CollectorID == uuid.Nil {
		return fmt.Errorf("collector_id tidak boleh kosong")
	}
	if len(cmd.PeriodMonth) != 7 || cmd.PeriodMonth[4] != '-' {
		return collection.ErrInvalidPeriodMonth
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &collection.Performance{
		ID:                   id,
		CollectorID:          cmd.CollectorID,
		PeriodMonth:          cmd.PeriodMonth,
		TotalCalls:           cmd.TotalCalls,
		SuccessfulContacts:   cmd.SuccessfulContacts,
		TotalVisits:          cmd.TotalVisits,
		SuccessfulVisits:     cmd.SuccessfulVisits,
		CasesHandled:         cmd.CasesHandled,
		CasesResolved:        cmd.CasesResolved,
		TotalAmountCollected: cmd.TotalAmountCollected,
		PromiseToPayCount:    cmd.PromiseToPayCount,
		PromiseKeptCount:     cmd.PromiseKeptCount,
		ContactRatePct:       cmd.ContactRatePct,
		ResolutionRatePct:    cmd.ResolutionRatePct,
		CollectionRatePct:    cmd.CollectionRatePct,
		PromiseKeptRatePct:   cmd.PromiseKeptRatePct,
		CalculatedAt:         now,
		CreatedAt:            now,
	}

	if err := h.repo.SavePerformance(ctx, s, entity); err != nil {
		return fmt.Errorf("save collector performance: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
