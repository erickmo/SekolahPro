// Package create_insurance_product menangani command untuk membuat Product asuransi baru.
package create_insurance_product

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_insurance_product"

// Command berisi data yang dibutuhkan untuk membuat Product baru.
type Command struct {
	Name              string
	Code              string
	ProductType       insurance.ProductType
	ProviderName      string
	PremiumAmount     int64
	CoverageAmount    int64
	PremiumFrequency  string
	TermMonths        int
	Description       string
	TermsConditions   string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_insurance_product.
type Handler struct {
	repo insurance.ProductWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ProductWriteRepository) *Handler {
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

	if cmd.Name == "" {
		return insurance.ErrNameEmpty
	}
	if cmd.Code == "" {
		return insurance.ErrCodeEmpty
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &insurance.Product{
		ID:               id,
		Name:             cmd.Name,
		Code:             cmd.Code,
		ProductType:      cmd.ProductType,
		Status:           insurance.ProductStatusDraft,
		ProviderName:     cmd.ProviderName,
		PremiumAmount:    cmd.PremiumAmount,
		CoverageAmount:   cmd.CoverageAmount,
		PremiumFrequency: cmd.PremiumFrequency,
		TermMonths:       cmd.TermMonths,
		Description:      cmd.Description,
		TermsConditions:  cmd.TermsConditions,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save insurance product: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
