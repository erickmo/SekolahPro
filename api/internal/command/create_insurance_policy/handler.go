// Package create_insurance_policy menangani command untuk membuat Policy asuransi baru.
package create_insurance_policy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_insurance_policy"

// Command berisi data yang dibutuhkan untuk membuat Policy baru.
type Command struct {
	ProductID             uuid.UUID
	NasabahID             uuid.UUID
	PolicyNo              string
	StartDate             time.Time
	EndDate               time.Time
	PremiumAmount         int64
	CoverageAmount        int64
	BeneficiaryName       string
	BeneficiaryRelation   string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_insurance_policy.
type Handler struct {
	repo insurance.PolicyWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.PolicyWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle memvalidasi command, membuat entity, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.PolicyNo == "" {
		return insurance.ErrPolicyNoEmpty
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &insurance.Policy{
		ID:                  id,
		ProductID:           cmd.ProductID,
		NasabahID:           cmd.NasabahID,
		PolicyNo:            cmd.PolicyNo,
		Status:              insurance.PolicyStatusPending,
		StartDate:           cmd.StartDate,
		EndDate:             cmd.EndDate,
		PremiumAmount:       cmd.PremiumAmount,
		CoverageAmount:      cmd.CoverageAmount,
		BeneficiaryName:     cmd.BeneficiaryName,
		BeneficiaryRelation: cmd.BeneficiaryRelation,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save insurance policy: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
