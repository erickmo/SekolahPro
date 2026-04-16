// Package create_rat_meeting menangani command untuk membuat RAT meeting baru.
package create_rat_meeting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/rat"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_rat_meeting"

// Command berisi data yang dibutuhkan untuk membuat RAT meeting baru.
type Command struct {
	MeetingType         string
	MeetingNumber       string
	Title               string
	Description         string
	MeetingDate         time.Time
	MeetingTime         string
	MeetingLocation     string
	MeetingMode         string
	OnlineLink          *string
	TotalEligibleMembers int
	QuorumRequired      int
	AgendaDocID         *uuid.UUID
	MinutesDocID        *uuid.UUID
	FinancialReportID   *uuid.UUID
	ShuProposalID       *uuid.UUID
	ConvenedBy          uuid.UUID
	SecretaryID         uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_rat_meeting.
type Handler struct {
	repo     rat.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo rat.WriteRepository, eb eventbus.EventBus) *Handler {
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

	if cmd.MeetingType == "" {
		return fmt.Errorf("meeting_type tidak boleh kosong")
	}
	if cmd.Title == "" {
		return fmt.Errorf("title tidak boleh kosong")
	}
	if cmd.MeetingDate.IsZero() {
		return fmt.Errorf("meeting_date tidak boleh kosong")
	}
	if cmd.MeetingLocation == "" {
		return fmt.Errorf("meeting_location tidak boleh kosong")
	}
	if cmd.MeetingMode == "" {
		return fmt.Errorf("meeting_mode tidak boleh kosong")
	}
	if cmd.ConvenedBy == uuid.Nil {
		return fmt.Errorf("convened_by tidak boleh kosong")
	}
	if cmd.SecretaryID == uuid.Nil {
		return fmt.Errorf("secretary_id tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &rat.Meeting{
		ID:                  id,
		MeetingType:         cmd.MeetingType,
		MeetingNumber:       cmd.MeetingNumber,
		Title:               cmd.Title,
		Description:         cmd.Description,
		MeetingDate:         cmd.MeetingDate,
		MeetingTime:         cmd.MeetingTime,
		MeetingLocation:     cmd.MeetingLocation,
		MeetingMode:         cmd.MeetingMode,
		OnlineLink:          cmd.OnlineLink,
		TotalEligibleMembers: cmd.TotalEligibleMembers,
		QuorumRequired:      cmd.QuorumRequired,
		ActualAttendees:     0,
		QuorumMet:           false,
		Status:              rat.MeetingStatusDraft,
		AgendaDocID:         cmd.AgendaDocID,
		MinutesDocID:        cmd.MinutesDocID,
		FinancialReportID:   cmd.FinancialReportID,
		ShuProposalID:       cmd.ShuProposalID,
		ConvenedBy:          cmd.ConvenedBy,
		SecretaryID:         cmd.SecretaryID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save rat meeting: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, rat.MeetingCreatedEvent{
		TenantID:    s.TenantID,
		CompanyID:   s.CompanyID,
		ID:          id,
		MeetingType: cmd.MeetingType,
		Title:       cmd.Title,
		MeetingDate: cmd.MeetingDate.Format("2006-01-02"),
	}); err != nil {
		return fmt.Errorf("publish rat_meeting.created: %w", err)
	}

	return nil
}
