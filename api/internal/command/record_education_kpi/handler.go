// Package record_education_kpi menangani command untuk mencatat EducationKPI baru.
package record_education_kpi

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "record_education_kpi"

// Command berisi data yang dibutuhkan untuk mencatat EducationKPI.
type Command struct {
	MetricType  enrollment.MetricType
	PeriodMonth string
	Value       float64
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command record_education_kpi.
type Handler struct {
	repo enrollment.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo enrollment.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle memvalidasi command, membuat entity KPI, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}
	if !isValidMetricType(cmd.MetricType) {
		return enrollment.ErrInvalidMetricType
	}
	if cmd.PeriodMonth == "" {
		return fmt.Errorf("period_month tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	kpi := &enrollment.EducationKPI{
		ID:          id,
		MetricType:  cmd.MetricType,
		PeriodMonth: cmd.PeriodMonth,
		Value:       cmd.Value,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.repo.SaveKPI(ctx, s, kpi); err != nil {
		return fmt.Errorf("save education kpi: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	return nil
}

func isValidMetricType(t enrollment.MetricType) bool {
	switch t {
	case enrollment.MetricCompletionRate, enrollment.MetricAvgScore,
		enrollment.MetricEnrollmentCount, enrollment.MetricPassRate:
		return true
	default:
		return false
	}
}
