// Package create_insurance_claim menangani command untuk membuat Claim asuransi baru.
package create_insurance_claim

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_insurance_claim"

// Command berisi data yang dibutuhkan untuk membuat Claim baru.
type Command struct {
	PolicyID      uuid.UUID
	ClaimType     insurance.ClaimType
	ClaimAmount   int64
	IncidentDate  time.Time
	SubmittedBy   uuid.UUID
	Description   string
	DocumentIDs   []uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_insurance_claim.
type Handler struct {
	repo insurance.ClaimWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ClaimWriteRepository) *Handler {
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

	if cmd.ClaimType == "" {
		return insurance.ErrInvalidClaimType
	}

	id := uuid.New()
	now := time.Now().UTC()

	docIDs := cmd.DocumentIDs
	if docIDs == nil {
		docIDs = []uuid.UUID{}
	}

	entity := &insurance.Claim{
		ID:            id,
		PolicyID:      cmd.PolicyID,
		ClaimType:     cmd.ClaimType,
		Status:        insurance.ClaimStatusSubmitted,
		ClaimAmount:   cmd.ClaimAmount,
		IncidentDate:  cmd.IncidentDate,
		SubmittedAt:   now,
		SubmittedBy:   cmd.SubmittedBy,
		Description:   cmd.Description,
		DocumentIDs:   docIDs,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save insurance claim: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
