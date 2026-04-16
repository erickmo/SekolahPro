// Package bcp mendefinisikan domain untuk Business Continuity Planning.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package bcp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// StrategyType mendefinisikan jenis backup strategy.
type StrategyType string

const (
	StrategyTypeFull         StrategyType = "full"
	StrategyTypeIncremental  StrategyType = "incremental"
	StrategyTypeDifferential StrategyType = "differential"
	StrategyTypeSnapshot     StrategyType = "snapshot"
)

// BackupStrategy adalah entity utama untuk domain BCP.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type BackupStrategy struct {
	ID               uuid.UUID    `db:"id"`
	StrategyType     StrategyType `db:"strategy_type"`
	Frequency        string       `db:"frequency"`
	RetentionDays    int          `db:"retention_days"`
	StorageLocation  string       `db:"storage_location"`
	EncryptionEnabled bool        `db:"encryption_enabled"`
	IsActive         bool         `db:"is_active"`
	LastBackupAt     *time.Time   `db:"last_backup_at"`
	NextBackupAt     *time.Time   `db:"next_backup_at"`
	CreatedAt        time.Time    `db:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at"`
	DeletedAt        *time.Time   `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain BCP.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *BackupStrategy) error
	Update(ctx context.Context, s scope.Scope, e *BackupStrategy) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain BCP.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*BackupStrategy, error)
	List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*BackupStrategy, int, error)
}
