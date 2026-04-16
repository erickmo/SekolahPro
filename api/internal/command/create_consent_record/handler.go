// Package create_consent_record menangani command untuk membuat consent record baru.
package create_consent_record

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/consent"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_consent_record"

// Command berisi data yang dibutuhkan untuk membuat consent record baru.
type Command struct {
	NasabahID          uuid.UUID
	ConsentType        string
	ConsentPurpose     string
	LegalBasis         string
	ConsentText        string
	ConsentVersion     string
	ConsentGiven       bool
	ConsentMethod      string
	ParentID           *uuid.UUID
	ParentRelationship *string
	ParentConsentGiven *bool
	IPAddress          *string
	UserAgent          *string
	WitnessID          *uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_consent_record.
type Handler struct {
	repo     consent.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo consent.WriteRepository, eb eventbus.EventBus) *Handler {
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
	if cmd.ConsentType == "" {
		return fmt.Errorf("consent_type tidak boleh kosong")
	}
	if cmd.ConsentPurpose == "" {
		return fmt.Errorf("consent_purpose tidak boleh kosong")
	}
	if cmd.LegalBasis == "" {
		return fmt.Errorf("legal_basis tidak boleh kosong")
	}
	if cmd.ConsentText == "" {
		return fmt.Errorf("consent_text tidak boleh kosong")
	}
	if cmd.ConsentMethod == "" {
		return fmt.Errorf("consent_method tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &consent.ConsentRecord{
		ID:                 id,
		NasabahID:          cmd.NasabahID,
		ConsentType:        cmd.ConsentType,
		ConsentPurpose:     cmd.ConsentPurpose,
		LegalBasis:         cmd.LegalBasis,
		ConsentText:        cmd.ConsentText,
		ConsentVersion:     cmd.ConsentVersion,
		ConsentGiven:       cmd.ConsentGiven,
		ConsentMethod:      cmd.ConsentMethod,
		Withdrawn:          false,
		ParentID:           cmd.ParentID,
		ParentRelationship: cmd.ParentRelationship,
		ParentConsentGiven: cmd.ParentConsentGiven,
		GivenAt:            now,
		IPAddress:          cmd.IPAddress,
		UserAgent:          cmd.UserAgent,
		WitnessID:          cmd.WitnessID,
		CreatedAt:          now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save consent record: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, consent.ConsentCreatedEvent{
		TenantID:     s.TenantID,
		CompanyID:    s.CompanyID,
		ID:           id,
		NasabahID:    cmd.NasabahID,
		ConsentType:  cmd.ConsentType,
		ConsentGiven: cmd.ConsentGiven,
	}); err != nil {
		return fmt.Errorf("publish consent_record.created: %w", err)
	}

	return nil
}
