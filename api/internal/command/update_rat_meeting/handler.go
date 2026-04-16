// Package update_rat_meeting menangani command untuk mengupdate RAT meeting.
package update_rat_meeting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/rat"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_rat_meeting"

// Command berisi data yang dibutuhkan untuk mengupdate RAT meeting.
type Command struct {
	ID                  uuid.UUID
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
	ActualAttendees     int
	QuorumMet           bool
	AgendaDocID         *uuid.UUID
	MinutesDocID        *uuid.UUID
	FinancialReportID   *uuid.UUID
	ShuProposalID       *uuid.UUID
	SecretaryID         uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_rat_meeting.
type Handler struct {
	readRepo  rat.ReadRepository
	writeRepo rat.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr rat.ReadRepository, wr rat.WriteRepository, eb eventbus.EventBus) *Handler {
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
		return fmt.Errorf("get rat meeting for update: %w", err)
	}

	if entity.Status == rat.MeetingStatusCompleted || entity.Status == rat.MeetingStatusCancelled {
		return fmt.Errorf("tidak dapat mengupdate meeting dengan status %s", entity.Status)
	}

	if cmd.MeetingType != "" {
		entity.MeetingType = cmd.MeetingType
	}
	if cmd.MeetingNumber != "" {
		entity.MeetingNumber = cmd.MeetingNumber
	}
	if cmd.Title != "" {
		entity.Title = cmd.Title
	}
	if cmd.Description != "" {
		entity.Description = cmd.Description
	}
	if !cmd.MeetingDate.IsZero() {
		entity.MeetingDate = cmd.MeetingDate
	}
	if cmd.MeetingTime != "" {
		entity.MeetingTime = cmd.MeetingTime
	}
	if cmd.MeetingLocation != "" {
		entity.MeetingLocation = cmd.MeetingLocation
	}
	if cmd.MeetingMode != "" {
		entity.MeetingMode = cmd.MeetingMode
	}
	entity.OnlineLink = cmd.OnlineLink
	if cmd.TotalEligibleMembers > 0 {
		entity.TotalEligibleMembers = cmd.TotalEligibleMembers
	}
	if cmd.QuorumRequired > 0 {
		entity.QuorumRequired = cmd.QuorumRequired
	}
	entity.ActualAttendees = cmd.ActualAttendees
	entity.QuorumMet = cmd.QuorumMet
	entity.AgendaDocID = cmd.AgendaDocID
	entity.MinutesDocID = cmd.MinutesDocID
	entity.FinancialReportID = cmd.FinancialReportID
	entity.ShuProposalID = cmd.ShuProposalID
	if cmd.SecretaryID != uuid.Nil {
		entity.SecretaryID = cmd.SecretaryID
	}
	entity.UpdatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update rat meeting: %w", err)
	}

	if err := h.eventBus.Publish(ctx, rat.MeetingUpdatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
	}); err != nil {
		return fmt.Errorf("publish rat_meeting.updated: %w", err)
	}

	return nil
}
