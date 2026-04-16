// Package create_collection_activity menangani command untuk membuat Activity baru.
package create_collection_activity

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_collection_activity"

// Command berisi data yang dibutuhkan untuk membuat Activity baru.
type Command struct {
	CaseID           uuid.UUID
	ActivityType     collection.ActivityType
	ActivityDate     time.Time
	PerformedBy      uuid.UUID
	ContactResult    collection.ContactResult
	Notes            string
	NasabahResponse  *collection.NasabahResponse
	PromiseAmount    *int64
	PromiseDate      *time.Time
	DocumentIDs      []uuid.UUID
	FollowupRequired bool
	FollowupDate     *time.Time
	FollowupType     *string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_collection_activity.
type Handler struct {
	repo collection.ActivityWriteRepository
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo collection.ActivityWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle memvalidasi command, membuat entity Activity, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.CaseID == uuid.Nil {
		return fmt.Errorf("case_id tidak boleh kosong")
	}
	if cmd.PerformedBy == uuid.Nil {
		return fmt.Errorf("performed_by tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &collection.Activity{
		ID:               id,
		CaseID:           cmd.CaseID,
		ActivityType:     cmd.ActivityType,
		ActivityDate:     cmd.ActivityDate,
		PerformedBy:      cmd.PerformedBy,
		ContactResult:    cmd.ContactResult,
		Notes:            cmd.Notes,
		NasabahResponse:  cmd.NasabahResponse,
		PromiseAmount:    cmd.PromiseAmount,
		PromiseDate:      cmd.PromiseDate,
		DocumentIDs:      cmd.DocumentIDs,
		FollowupRequired: cmd.FollowupRequired,
		FollowupDate:     cmd.FollowupDate,
		FollowupType:     cmd.FollowupType,
		CreatedAt:        now,
	}

	if err := h.repo.SaveActivity(ctx, s, entity); err != nil {
		return fmt.Errorf("save collection activity: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
