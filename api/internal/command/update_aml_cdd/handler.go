// Package update_aml_cdd menangani command untuk mengupdate CDD record.
package update_aml_cdd

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_cdd"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_aml_cdd"

// Command berisi data yang dibutuhkan untuk mengupdate CDD record.
type Command struct {
	ID                     uuid.UUID
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
	Status                 string
	RestrictionReason      *string
	ExitReason             *string
	UpdatedBy              uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_aml_cdd.
type Handler struct {
	readRepo  aml_cdd.ReadRepository
	writeRepo aml_cdd.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr aml_cdd.ReadRepository, wr aml_cdd.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengambil entity yang ada, mengupdate, dan mempublish event.
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
		return fmt.Errorf("get aml cdd for update: %w", err)
	}

	if cmd.CDDPurpose != "" {
		entity.CDDPurpose = cmd.CDDPurpose
	}
	entity.SourceOfFunds = cmd.SourceOfFunds
	entity.SourceOfWealth = cmd.SourceOfWealth
	entity.IsPEP = cmd.IsPEP
	if cmd.PEPType != "" {
		entity.PEPType = cmd.PEPType
	}
	entity.PEPPosition = cmd.PEPPosition
	entity.PEPCountry = cmd.PEPCountry
	if cmd.PEPRelationship != "" {
		entity.PEPRelationship = cmd.PEPRelationship
	}
	entity.PEPScreeningDate = cmd.PEPScreeningDate
	entity.PEPScreeningSource = cmd.PEPScreeningSource
	entity.BeneficialOwnerName = cmd.BeneficialOwnerName
	entity.BeneficialOwnerIDNo = cmd.BeneficialOwnerIDNo
	entity.BeneficialOwnershipPct = cmd.BeneficialOwnershipPct
	entity.BOVerified = cmd.BOVerified
	if cmd.Status != "" {
		entity.Status = cmd.Status
	}
	entity.RestrictionReason = cmd.RestrictionReason
	entity.ExitReason = cmd.ExitReason
	entity.UpdatedAt = time.Now().UTC()
	if cmd.UpdatedBy != uuid.Nil {
		entity.UpdatedBy = cmd.UpdatedBy
	}

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update aml cdd: %w", err)
	}

	if err := h.eventBus.Publish(ctx, aml_cdd.CDDUpdatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
	}); err != nil {
		return fmt.Errorf("publish aml_cdd.updated: %w", err)
	}

	return nil
}
