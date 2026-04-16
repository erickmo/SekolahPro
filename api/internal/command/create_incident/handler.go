// Package create_incident menangani command untuk membuat IncidentRecord baru.
package create_incident

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/incident"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_incident"

// Command berisi data yang dibutuhkan untuk membuat IncidentRecord baru.
type Command struct {
	Severity        string
	IncidentType    string
	Title           string
	Description     string
	AffectedSystems []string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_incident.
type Handler struct {
	repo     incident.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo incident.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{repo: repo, eventBus: eb}
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

	if err := validateSeverity(cmd.Severity); err != nil {
		return err
	}
	if cmd.Title == "" {
		return incident.ErrTitleEmpty
	}
	if len(cmd.Title) > 255 {
		return incident.ErrTitleTooLong
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &incident.IncidentRecord{
		ID:              id,
		Severity:        incident.Severity(cmd.Severity),
		Status:          incident.StatusOpen,
		IncidentType:    cmd.IncidentType,
		Title:           cmd.Title,
		Description:     cmd.Description,
		AffectedSystems: cmd.AffectedSystems,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save incident: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, incident.IncidentCreatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        id,
		Severity:  cmd.Severity,
		Title:     cmd.Title,
	}); err != nil {
		return fmt.Errorf("publish incident.created: %w", err)
	}

	return nil
}

// validSeverities berisi daftar severity yang diterima.
var validSeverities = map[string]bool{
	"critical": true, "high": true, "medium": true, "low": true,
}

func validateSeverity(s string) error {
	if s == "" {
		return incident.ErrSeverityEmpty
	}
	if !validSeverities[s] {
		return incident.ErrInvalidSeverity
	}
	return nil
}
