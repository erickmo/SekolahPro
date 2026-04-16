// Package list_backup_strategies menangani query untuk mengambil daftar BackupStrategy.
package list_backup_strategies

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/bcp"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_backup_strategies"

// Query berisi parameter untuk mengambil daftar BackupStrategy.
type Query struct {
	Params pagination.ListParams
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu BackupStrategy dalam hasil list.
type Item struct {
	ID                string  `json:"id"`
	StrategyType      string  `json:"strategy_type"`
	Frequency         string  `json:"frequency"`
	RetentionDays     int     `json:"retention_days"`
	StorageLocation   string  `json:"storage_location"`
	EncryptionEnabled bool    `json:"encryption_enabled"`
	IsActive          bool    `json:"is_active"`
	LastBackupAt      *string `json:"last_backup_at"`
	NextBackupAt      *string `json:"next_backup_at"`
	CreatedAt         string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_backup_strategies.
type Handler struct {
	repo bcp.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo bcp.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar BackupStrategy dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list backup strategies: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		item := &Item{
			ID:                e.ID.String(),
			StrategyType:      string(e.StrategyType),
			Frequency:         e.Frequency,
			RetentionDays:     e.RetentionDays,
			StorageLocation:   e.StorageLocation,
			EncryptionEnabled: e.EncryptionEnabled,
			IsActive:          e.IsActive,
			CreatedAt:         e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.LastBackupAt != nil {
			v := e.LastBackupAt.Format("2006-01-02T15:04:05Z")
			item.LastBackupAt = &v
		}
		if e.NextBackupAt != nil {
			v := e.NextBackupAt.Format("2006-01-02T15:04:05Z")
			item.NextBackupAt = &v
		}
		items = append(items, item)
	}

	return &Result{Data: items, Total: total}, nil
}
