// Package governance adalah domain CQRS untuk cooperative governance & internal controls (ADR-K025).
//
// Mengelola struktur organisasi koperasi: posisi pengurus/pengawas, authority matrix,
// internal audit, dan conflict of interest.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package governance

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// GovernancePosition adalah entity untuk posisi organisasi koperasi.
type GovernancePosition struct {
	ID                  uuid.UUID  `db:"id"                   json:"id"`
	NasabahID           uuid.UUID  `db:"nasabah_id"           json:"nasabah_id"`
	PositionType        string     `db:"position_type"        json:"position_type"`
	PositionLevel       string     `db:"position_level"       json:"position_level"`
	TermStart           time.Time  `db:"term_start"           json:"term_start"`
	TermEnd             time.Time  `db:"term_end"             json:"term_end"`
	TermNumber          int        `db:"term_number"          json:"term_number"`
	Status              string     `db:"status"               json:"status"`
	AppointedBy         string     `db:"appointed_by"         json:"appointed_by"`
	AppointmentDocID    *uuid.UUID `db:"appointment_doc_id"   json:"appointment_doc_id,omitempty"`
	MaxApprovalAmount   int64      `db:"max_approval_amount"  json:"max_approval_amount"`
	CanDisburse         bool       `db:"can_disburse"         json:"can_disburse"`
	CanReverse          bool       `db:"can_reverse"          json:"can_reverse"`
	CanWaivePenalty     bool       `db:"can_waive_penalty"    json:"can_waive_penalty"`
	CanWriteOff         bool       `db:"can_write_off"        json:"can_write_off"`
	CreatedAt           time.Time  `db:"created_at"           json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"           json:"updated_at"`
	CreatedBy           uuid.UUID  `db:"created_by"           json:"created_by"`
	UpdatedBy           uuid.UUID  `db:"updated_by"           json:"updated_by"`
	DeletedAt           *time.Time `db:"deleted_at"           json:"-"`
}

// Position type constants.
const (
	PositionKetua            = "ketua"
	PositionSekretaris       = "sekretaris"
	PositionBendahara        = "bendahara"
	PositionPengawasKetua    = "pengawas_ketua"
	PositionPengawasAnggota  = "pengawas_anggota"
	PositionManajer          = "manajer"
	PositionDPSKetua         = "dps_ketua"
	PositionDPSAnggota       = "dps_anggota"
	PositionTeller           = "teller"
	PositionSupervisor       = "supervisor"
	PositionAdmin            = "admin"
)

// Position level constants.
const (
	LevelStrategic  = "strategic"
	LevelManagement = "management"
	LevelOperational = "operational"
)

// Position status constants.
const (
	StatusActive   = "active"
	StatusResigned = "resigned"
	StatusRemoved  = "removed"
	StatusExpired  = "expired"
)

// Appointed by constants.
const (
	AppointedByRAT             = "rat"
	AppointedByBoardResolution = "board_resolution"
	AppointedByManager         = "manager"
)

// ListFilter menyimpan filter opsional untuk List query.
type ListFilter struct {
	PositionType *string
	Status       *string
	PositionLevel *string
}

// WriteRepository mendefinisikan operasi write untuk governance positions.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *GovernancePosition) error
	Update(ctx context.Context, s scope.Scope, e *GovernancePosition) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk governance positions.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*GovernancePosition, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*GovernancePosition, int, error)
}
