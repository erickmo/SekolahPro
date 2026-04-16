// Package create_aml_cdd menangani command untuk membuat CDD record baru.
package create_aml_cdd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_cdd"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_aml_cdd"

// Command berisi data yang dibutuhkan untuk membuat CDD record baru.
type Command struct {
	NasabahID              uuid.UUID
	RiskLevel              string
	RiskScore              int
	RiskCategory           json.RawMessage
	CDDLevel               string
	CDDPurpose             string
	SourceOfFunds          *string
	SourceOfWealth         *string
	IsPEP                  bool
	PEPType                string
	PEPPosition            *string
	PEPCountry             *string
	PEPRelationship        string
	PEPScreeningDate       *time.Time
	PEPScreeningSource     *string
	BeneficialOwnerName    *string
	BeneficialOwnerIDNo    *string
	BeneficialOwnershipPct *float64
	BOVerified             bool
	NextReviewDate         *time.Time
	CreatedBy              uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_aml_cdd.
type Handler struct {
	repo     aml_cdd.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo aml_cdd.WriteRepository, eb eventbus.EventBus) *Handler {
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
	if cmd.RiskLevel == "" {
		return fmt.Errorf("risk_level tidak boleh kosong")
	}
	if cmd.CDDLevel == "" {
		return fmt.Errorf("cdd_level tidak boleh kosong")
	}
	if cmd.CreatedBy == uuid.Nil {
		return fmt.Errorf("created_by tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &aml_cdd.CustomerDueDiligence{
		ID:                     id,
		NasabahID:              cmd.NasabahID,
		RiskLevel:              cmd.RiskLevel,
		RiskScore:              cmd.RiskScore,
		RiskCategory:           cmd.RiskCategory,
		CDDLevel:               cmd.CDDLevel,
		CDDPurpose:             cmd.CDDPurpose,
		SourceOfFunds:          cmd.SourceOfFunds,
		SourceOfWealth:         cmd.SourceOfWealth,
		IsPEP:                  cmd.IsPEP,
		PEPType:                cmd.PEPType,
		PEPPosition:            cmd.PEPPosition,
		PEPCountry:             cmd.PEPCountry,
		PEPRelationship:        cmd.PEPRelationship,
		PEPScreeningDate:       cmd.PEPScreeningDate,
		PEPScreeningSource:     cmd.PEPScreeningSource,
		BeneficialOwnerName:    cmd.BeneficialOwnerName,
		BeneficialOwnerIDNo:    cmd.BeneficialOwnerIDNo,
		BeneficialOwnershipPct: cmd.BeneficialOwnershipPct,
		BOVerified:             cmd.BOVerified,
		NextReviewDate:         cmd.NextReviewDate,
		ReviewCount:            0,
		Status:                 aml_cdd.CDDStatusPending,
		CreatedAt:              now,
		UpdatedAt:              now,
		CreatedBy:              cmd.CreatedBy,
		UpdatedBy:              cmd.CreatedBy,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save aml cdd: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, aml_cdd.CDDCreatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        id,
		NasabahID: cmd.NasabahID,
		RiskLevel: cmd.RiskLevel,
		CDDLevel:  cmd.CDDLevel,
	}); err != nil {
		return fmt.Errorf("publish aml_cdd.created: %w", err)
	}

	return nil
}
