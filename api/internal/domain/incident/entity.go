// Package incident mendefinisikan domain untuk Incident Management (BCP/DRP).
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package incident

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// Severity mendefinisikan level severity incident.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Status mendefinisikan status incident.
type Status string

const (
	StatusOpen         Status = "open"
	StatusInvestigating Status = "investigating"
	StatusResolved     Status = "resolved"
	StatusClosed       Status = "closed"
)

// IncidentRecord adalah entity utama untuk domain Incident.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type IncidentRecord struct {
	ID               uuid.UUID  `db:"id"`
	Severity         Severity   `db:"severity"`
	Status           Status     `db:"status"`
	IncidentType     string     `db:"incident_type"`
	Title            string     `db:"title"`
	Description      string     `db:"description"`
	AffectedSystems  []string   `db:"affected_systems"`
	RootCause        *string    `db:"root_cause"`
	Resolution       *string    `db:"resolution"`
	ResolutionTime   *int       `db:"resolution_time"`
	ResolvedBy       *uuid.UUID `db:"resolved_by"`
	ResolvedAt       *time.Time `db:"resolved_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain Incident.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *IncidentRecord) error
	Update(ctx context.Context, s scope.Scope, e *IncidentRecord) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain Incident.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*IncidentRecord, error)
	List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*IncidentRecord, int, error)
}
