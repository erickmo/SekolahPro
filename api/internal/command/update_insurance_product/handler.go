// Package update_insurance_product menangani command untuk mengubah Product asuransi.
package update_insurance_product

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_insurance_product"

// Command berisi data untuk mengubah Product asuransi.
type Command struct {
	ID                uuid.UUID
	Name              string
	Code              string
	ProductType       insurance.ProductType
	Status            insurance.ProductStatus
	ProviderName      string
	PremiumAmount     int64
	CoverageAmount    int64
	PremiumFrequency  string
	TermMonths        int
	Description       string
	TermsConditions   string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_insurance_product.
type Handler struct {
	repo insurance.ProductWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ProductWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah field, dan menyimpan kembali.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(insurance.ProductReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement ProductReadRepository")
	}

	entity, err := readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get insurance product: %w", err)
	}

	entity.Name = cmd.Name
	entity.Code = cmd.Code
	entity.ProductType = cmd.ProductType
	entity.Status = cmd.Status
	entity.ProviderName = cmd.ProviderName
	entity.PremiumAmount = cmd.PremiumAmount
	entity.CoverageAmount = cmd.CoverageAmount
	entity.PremiumFrequency = cmd.PremiumFrequency
	entity.TermMonths = cmd.TermMonths
	entity.Description = cmd.Description
	entity.TermsConditions = cmd.TermsConditions
	entity.UpdatedAt = time.Now().UTC()

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update insurance product: %w", err)
	}
	return nil
}
