// Package update_aml_cdd_risk menangani command untuk mengupdate risk assessment CDD.
package update_aml_cdd_risk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_cdd"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_aml_cdd_risk"

// Command berisi data yang dibutuhkan untuk mengupdate risk assessment CDD.
type Command struct {
	ID            uuid.UUID
	RiskLevel     string
	RiskScore     int
	RiskCategory  json.RawMessage
	CDDLevel      string
	NextReviewDate *time.Time
	UpdatedBy     uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_aml_cdd_risk.
type Handler struct {
	readRepo  aml_cdd.ReadRepository
	writeRepo aml_cdd.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr aml_cdd.ReadRepository, wr aml_cdd.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi dan mengupdate risk assessment CDD.
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
		return fmt.Errorf("get aml cdd for risk update: %w", err)
	}

	oldRisk := entity.RiskLevel

	if cmd.RiskLevel != "" {
		entity.RiskLevel = cmd.RiskLevel
	}
	entity.RiskScore = cmd.RiskScore
	if cmd.RiskCategory != nil {
		entity.RiskCategory = cmd.RiskCategory
	}
	if cmd.CDDLevel != "" {
		entity.CDDLevel = cmd.CDDLevel
	}
	entity.NextReviewDate = cmd.NextReviewDate
	entity.ReviewCount++
	now := time.Now().UTC()
	entity.LastReviewedAt = &now
	if cmd.UpdatedBy != uuid.Nil {
		entity.LastReviewedBy = &cmd.UpdatedBy
		entity.UpdatedBy = cmd.UpdatedBy
	}
	entity.UpdatedAt = now

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update aml cdd risk: %w", err)
	}

	if err := h.eventBus.Publish(ctx, aml_cdd.CDDRiskUpdatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		OldRisk:   oldRisk,
		NewRisk:   entity.RiskLevel,
	}); err != nil {
		return fmt.Errorf("publish aml_cdd.risk_updated: %w", err)
	}

	return nil
}
