// Package record_kpi_measurement menangani command untuk merekam KPIMeasurement.
package record_kpi_measurement

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "record_kpi_measurement"

// Command berisi data untuk merekam KPIMeasurement baru.
type Command struct {
	DefinitionID uuid.UUID
	MeasuredValue float64
	MeasuredBy   uuid.UUID
	Notes        string
	PeriodStart  time.Time
	PeriodEnd    time.Time
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command record_kpi_measurement.
type Handler struct {
	repo kpi.KPIMeasurementWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIMeasurementWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle membuat entity KPIMeasurement dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	id := uuid.New()
	now := time.Now().UTC()

	status := kpi.MeasurementStatusNormal
	// Status ditentukan berdasarkan logika sederhana di application layer.

	entity := &kpi.KPIMeasurement{
		ID:            id,
		DefinitionID:  cmd.DefinitionID,
		MeasuredValue: cmd.MeasuredValue,
		Status:        status,
		MeasuredAt:    now,
		MeasuredBy:    cmd.MeasuredBy,
		Notes:         cmd.Notes,
		PeriodStart:   cmd.PeriodStart,
		PeriodEnd:     cmd.PeriodEnd,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.repo.SaveMeasurement(ctx, s, entity); err != nil {
		return fmt.Errorf("save kpi measurement: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
