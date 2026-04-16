// Package create_dissolution menangani command untuk membuat proses pembubaran koperasi baru.
package create_dissolution

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_dissolution"

// Command berisi data yang dibutuhkan untuk membuat proses pembubaran baru.
type Command struct {
	DissolutionType dissolution.DissolutionType
	RatMeetingID    *uuid.UUID
	Reason          string
	EffectiveDate   time.Time
	LiquidatorIDs   []uuid.UUID
	SupervisorID    *uuid.UUID
	ClaimDeadline   *time.Time
	InitiatedBy     uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_dissolution.
type Handler struct {
	repo     dissolution.WriteRepository
	readRepo dissolution.ReadRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo dissolution.WriteRepository, readRepo dissolution.ReadRepository, eb eventbus.EventBus) *Handler {
	return &Handler{repo: repo, readRepo: readRepo, eventBus: eb}
}

// Handle memvalidasi command, membuat entity, menyimpan ke repository,
// dan mempublikasikan domain event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.DissolutionType == "" {
		return dissolution.ErrDissolutionTypeRequired
	}
	if !dissolution.ValidDissolutionTypes[cmd.DissolutionType] {
		return dissolution.ErrInvalidDissolutionType
	}
	if cmd.Reason == "" {
		return dissolution.ErrReasonRequired
	}
	if cmd.EffectiveDate.IsZero() {
		return dissolution.ErrEffectiveDateRequired
	}
	if cmd.InitiatedBy == uuid.Nil {
		return dissolution.ErrInitiatedByRequired
	}
	if cmd.DissolutionType == dissolution.DissolutionTypeVoluntary && cmd.RatMeetingID == nil {
		return dissolution.ErrRatMeetingRequired
	}

	id := uuid.New()
	now := time.Now().UTC()

	liquidatorIDs := cmd.LiquidatorIDs
	if liquidatorIDs == nil {
		liquidatorIDs = []uuid.UUID{}
	}

	entity := &dissolution.DissolutionProcess{
		ID:              id,
		DissolutionType: cmd.DissolutionType,
		RatMeetingID:    cmd.RatMeetingID,
		Reason:          cmd.Reason,
		EffectiveDate:   cmd.EffectiveDate,
		LiquidatorIDs:   liquidatorIDs,
		SupervisorID:    cmd.SupervisorID,
		Stage:           dissolution.StageAnnounced,
		ClaimDeadline:   cmd.ClaimDeadline,
		Status:          dissolution.StatusInitiated,
		InitiatedAt:     now,
		InitiatedBy:     cmd.InitiatedBy,
		CreatedAt:       now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save dissolution: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, dissolution.DissolutionInitiatedEvent{
		TenantID:        s.TenantID,
		CompanyID:       s.CompanyID,
		ID:              id,
		DissolutionType: cmd.DissolutionType,
		Stage:           dissolution.StageAnnounced,
		Status:          dissolution.StatusInitiated,
	}); err != nil {
		return fmt.Errorf("publish dissolution.initiated: %w", err)
	}

	return nil
}
