// Package resolve_incident menangani command untuk menyelesaikan IncidentRecord.
package resolve_incident

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/incident"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "resolve_incident"

// Command berisi data yang dibutuhkan untuk menyelesaikan IncidentRecord.
type Command struct {
	ID             uuid.UUID
	RootCause      string
	Resolution     string
	ResolutionTime int
	ResolvedBy     uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command resolve_incident.
type Handler struct {
	readRepo  incident.ReadRepository
	writeRepo incident.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr incident.ReadRepository, wr incident.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengambil entity yang ada, mengupdate status, dan mempublish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get incident for resolve: %w", err)
	}

	if entity.Status == incident.StatusResolved || entity.Status == incident.StatusClosed {
		return incident.ErrAlreadyResolved
	}

	now := time.Now().UTC()
	resolutionTime := cmd.ResolutionTime
	rootCause := cmd.RootCause
	resolution := cmd.Resolution
	resolvedBy := cmd.ResolvedBy

	entity.Status = incident.StatusResolved
	entity.RootCause = &rootCause
	entity.Resolution = &resolution
	entity.ResolutionTime = &resolutionTime
	entity.ResolvedBy = &resolvedBy
	entity.ResolvedAt = &now
	entity.UpdatedAt = now

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("resolve incident: %w", err)
	}

	if err := h.eventBus.Publish(ctx, incident.IncidentResolvedEvent{
		TenantID:       s.TenantID,
		CompanyID:      s.CompanyID,
		ID:             entity.ID,
		Severity:       string(entity.Severity),
		ResolutionTime: resolutionTime,
	}); err != nil {
		return fmt.Errorf("publish incident.resolved: %w", err)
	}

	return nil
}
