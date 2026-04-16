// Package create_aml_alert menangani command untuk membuat transaction alert baru.
package create_aml_alert

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_alert"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_aml_alert"

// Command berisi data yang dibutuhkan untuk membuat transaction alert baru.
type Command struct {
	NasabahID          uuid.UUID
	TransaksiID        *uuid.UUID
	RuleID             uuid.UUID
	RuleType           string
	AlertType          string
	Severity           string
	Description        string
	TransactionDetails json.RawMessage
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_aml_alert.
type Handler struct {
	repo     aml_alert.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo aml_alert.WriteRepository, eb eventbus.EventBus) *Handler {
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

	if cmd.NasabahID == uuid.Nil {
		return fmt.Errorf("nasabah_id tidak boleh kosong")
	}
	if cmd.RuleID == uuid.Nil {
		return fmt.Errorf("rule_id tidak boleh kosong")
	}
	if cmd.AlertType == "" {
		return fmt.Errorf("alert_type tidak boleh kosong")
	}
	if cmd.Severity == "" {
		return fmt.Errorf("severity tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &aml_alert.TransactionAlert{
		ID:                 id,
		NasabahID:          cmd.NasabahID,
		TransaksiID:        cmd.TransaksiID,
		RuleID:             cmd.RuleID,
		RuleType:           cmd.RuleType,
		AlertType:          cmd.AlertType,
		Severity:           cmd.Severity,
		Description:        cmd.Description,
		TransactionDetails: cmd.TransactionDetails,
		LTKMFiled:         false,
		Status:             aml_alert.AlertStatusNew,
		DetectedAt:         now,
		CreatedAt:          now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save aml alert: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, aml_alert.AlertCreatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        id,
		NasabahID: cmd.NasabahID,
		RuleID:    cmd.RuleID,
		AlertType: cmd.AlertType,
		Severity:  cmd.Severity,
	}); err != nil {
		return fmt.Errorf("publish aml_alert.created: %w", err)
	}

	return nil
}
