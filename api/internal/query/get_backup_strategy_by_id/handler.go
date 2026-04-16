// Package get_backup_strategy_by_id menangani query untuk mengambil BackupStrategy berdasarkan ID.
package get_backup_strategy_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/bcp"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_backup_strategy_by_id"

// Query berisi parameter untuk mengambil BackupStrategy berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                 string  `json:"id"`
	StrategyType       string  `json:"strategy_type"`
	Frequency          string  `json:"frequency"`
	RetentionDays      int     `json:"retention_days"`
	StorageLocation    string  `json:"storage_location"`
	EncryptionEnabled  bool    `json:"encryption_enabled"`
	IsActive           bool    `json:"is_active"`
	LastBackupAt       *string `json:"last_backup_at"`
	NextBackupAt       *string `json:"next_backup_at"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

// Handler menangani Query get_backup_strategy_by_id.
type Handler struct {
	repo bcp.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo bcp.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil BackupStrategy berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get backup strategy by id: %w", err)
	}

	result := &Result{
		ID:                entity.ID.String(),
		StrategyType:      string(entity.StrategyType),
		Frequency:         entity.Frequency,
		RetentionDays:     entity.RetentionDays,
		StorageLocation:   entity.StorageLocation,
		EncryptionEnabled: entity.EncryptionEnabled,
		IsActive:          entity.IsActive,
		CreatedAt:         entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if entity.LastBackupAt != nil {
		v := entity.LastBackupAt.Format("2006-01-02T15:04:05Z")
		result.LastBackupAt = &v
	}
	if entity.NextBackupAt != nil {
		v := entity.NextBackupAt.Format("2006-01-02T15:04:05Z")
		result.NextBackupAt = &v
	}

	return result, nil
}
