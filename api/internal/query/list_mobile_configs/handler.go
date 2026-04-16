// Package list_mobile_configs menangani query untuk mengambil daftar MobileAppConfig.
package list_mobile_configs

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/mobile_config"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_mobile_configs"

// Query berisi parameter untuk mengambil daftar MobileAppConfig.
type Query struct {
	Params  pagination.ListParams
	Filters mobile_config.ListFilters
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu MobileAppConfig dalam hasil list.
type Item struct {
	ID              uuid.UUID                  `json:"id"`
	AppVariant      mobile_config.AppVariant   `json:"app_variant"`
	Platform        mobile_config.Platform     `json:"platform"`
	MinVersion      string                     `json:"min_version"`
	CurrentVersion  string                     `json:"current_version"`
	ForceUpdate     bool                       `json:"force_update"`
	MaintenanceMode bool                       `json:"maintenance_mode"`
	FeatureFlags    mobile_config.FeatureFlags  `json:"feature_flags"`
	APIBaseURL      string                     `json:"api_base_url"`
	ThemeConfig     mobile_config.ThemeConfig   `json:"theme_config"`
	OfflineConfig   mobile_config.OfflineConfig `json:"offline_config"`
	IsActive        bool                       `json:"is_active"`
	CreatedAt       string                     `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_mobile_configs.
type Handler struct {
	repo mobile_config.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo mobile_config.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar MobileAppConfig dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order, qry.Filters)
	if err != nil {
		return nil, fmt.Errorf("list mobile_configs: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:              e.ID,
			AppVariant:      e.AppVariant,
			Platform:        e.Platform,
			MinVersion:      e.MinVersion,
			CurrentVersion:  e.CurrentVersion,
			ForceUpdate:     e.ForceUpdate,
			MaintenanceMode: e.MaintenanceMode,
			FeatureFlags:    e.FeatureFlags,
			APIBaseURL:      e.APIBaseURL,
			ThemeConfig:     e.ThemeConfig,
			OfflineConfig:   e.OfflineConfig,
			IsActive:        e.IsActive,
			CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
