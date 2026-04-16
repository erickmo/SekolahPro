// Package kpi mendefinisikan domain untuk KPI / Health Indicators.
//
// Mengelola definisi KPI, pengukuran berkala, dan skenario stress test
// untuk memantau kesehatan koperasi sekolah.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package kpi

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── KPIDefinition ─────────────────────────────────────────────────────────────

// KPICategory mendefinisikan kategori KPI.
type KPICategory string

const (
	KPICategoryFinancial    KPICategory = "financial"
	KPICategoryOperational  KPICategory = "operational"
	KPICategoryMembership   KPICategory = "membership"
	KPICategoryCompliance   KPICategory = "compliance"
	KPICategoryServiceQuality KPICategory = "service_quality"
)

// KPIFrequency mendefinisikan frekuensi pengukuran.
type KPIFrequency string

const (
	KPIFrequencyDaily   KPIFrequency = "daily"
	KPIFrequencyWeekly  KPIFrequency = "weekly"
	KPIFrequencyMonthly KPIFrequency = "monthly"
	KPIFrequencyQuarterly KPIFrequency = "quarterly"
	KPIFrequencyYearly  KPIFrequency = "yearly"
)

// KPIDefinition adalah entity untuk definisi indikator kinerja.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type KPIDefinition struct {
	ID           uuid.UUID   `db:"id"`
	Name         string      `db:"name"`
	Code         string      `db:"code"`
	Category     KPICategory `db:"category"`
	Frequency    KPIFrequency `db:"frequency"`
	Unit         string      `db:"unit"`
	TargetValue  float64     `db:"target_value"`
	WarningThreshold float64 `db:"warning_threshold"`
	CriticalThreshold float64 `db:"critical_threshold"`
	IsActive     bool        `db:"is_active"`
	Description  string      `db:"description"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
	DeletedAt    *time.Time  `db:"deleted_at"`
}

// KPIDefinitionFilter berisi parameter filter opsional untuk list KPIDefinition.
type KPIDefinitionFilter struct {
	Category *KPICategory
	Frequency *KPIFrequency
	IsActive *bool
}

// KPIDefinitionWriteRepository mendefinisikan operasi write untuk KPIDefinition.
type KPIDefinitionWriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *KPIDefinition) error
	Update(ctx context.Context, s scope.Scope, e *KPIDefinition) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// KPIDefinitionReadRepository mendefinisikan operasi read untuk KPIDefinition.
type KPIDefinitionReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*KPIDefinition, error)
	List(ctx context.Context, s scope.Scope, filter KPIDefinitionFilter, limit, offset int, sortBy, order string) ([]*KPIDefinition, int, error)
}

// ── KPIMeasurement ────────────────────────────────────────────────────────────

// MeasurementStatus mendefinisikan status pengukuran.
type MeasurementStatus string

const (
	MeasurementStatusNormal   MeasurementStatus = "normal"
	MeasurementStatusWarning  MeasurementStatus = "warning"
	MeasurementStatusCritical MeasurementStatus = "critical"
)

// KPIMeasurement adalah entity untuk hasil pengukuran KPI.
type KPIMeasurement struct {
	ID             uuid.UUID        `db:"id"`
	DefinitionID   uuid.UUID        `db:"definition_id"`
	MeasuredValue  float64          `db:"measured_value"`
	Status         MeasurementStatus `db:"status"`
	MeasuredAt     time.Time        `db:"measured_at"`
	MeasuredBy     uuid.UUID        `db:"measured_by"`
	Notes          string           `db:"notes"`
	PeriodStart    time.Time        `db:"period_start"`
	PeriodEnd      time.Time        `db:"period_end"`
	CreatedAt      time.Time        `db:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at"`
	DeletedAt      *time.Time       `db:"deleted_at"`
}

// KPIMeasurementFilter berisi parameter filter opsional untuk list KPIMeasurement.
type KPIMeasurementFilter struct {
	DefinitionID *uuid.UUID
	Status       *MeasurementStatus
	MeasuredAtFrom *time.Time
	MeasuredAtTo   *time.Time
}

// KPIMeasurementWriteRepository mendefinisikan operasi write untuk KPIMeasurement.
type KPIMeasurementWriteRepository interface {
	SaveMeasurement(ctx context.Context, s scope.Scope, e *KPIMeasurement) error
	UpdateMeasurement(ctx context.Context, s scope.Scope, e *KPIMeasurement) error
	DeleteMeasurement(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// KPIMeasurementReadRepository mendefinisikan operasi read untuk KPIMeasurement.
type KPIMeasurementReadRepository interface {
	GetMeasurementByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*KPIMeasurement, error)
	ListMeasurements(ctx context.Context, s scope.Scope, filter KPIMeasurementFilter, limit, offset int, sortBy, order string) ([]*KPIMeasurement, int, error)
}

// ── StressTestScenario ────────────────────────────────────────────────────────

// ScenarioStatus mendefinisikan status skenario stress test.
type ScenarioStatus string

const (
	ScenarioStatusDraft     ScenarioStatus = "draft"
	ScenarioStatusRunning   ScenarioStatus = "running"
	ScenarioStatusCompleted ScenarioStatus = "completed"
	ScenarioStatusCancelled ScenarioStatus = "cancelled"
)

// StressTestScenario adalah entity untuk skenario uji stres koperasi.
type StressTestScenario struct {
	ID              uuid.UUID     `db:"id"`
	Name            string        `db:"name"`
	Description     string        `db:"description"`
	ScenarioStatus  ScenarioStatus `db:"scenario_status"`
	Parameters      string        `db:"parameters"`
	Results         string        `db:"results"`
	SimulatedAt     *time.Time    `db:"simulated_at"`
	SimulatedBy     *uuid.UUID    `db:"simulated_by"`
	CompletedAt     *time.Time    `db:"completed_at"`
	CreatedAt       time.Time     `db:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"`
	DeletedAt       *time.Time    `db:"deleted_at"`
}

// StressTestFilter berisi parameter filter opsional untuk list StressTestScenario.
type StressTestFilter struct {
	ScenarioStatus *ScenarioStatus
}

// StressTestWriteRepository mendefinisikan operasi write untuk StressTestScenario.
type StressTestWriteRepository interface {
	SaveScenario(ctx context.Context, s scope.Scope, e *StressTestScenario) error
	UpdateScenario(ctx context.Context, s scope.Scope, e *StressTestScenario) error
	DeleteScenario(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// StressTestReadRepository mendefinisikan operasi read untuk StressTestScenario.
type StressTestReadRepository interface {
	GetScenarioByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*StressTestScenario, error)
	ListScenarios(ctx context.Context, s scope.Scope, filter StressTestFilter, limit, offset int, sortBy, order string) ([]*StressTestScenario, int, error)
}
