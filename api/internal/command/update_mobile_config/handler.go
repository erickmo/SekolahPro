// Package update_mobile_config menangani command untuk mengupdate MobileAppConfig yang sudah ada.
package update_mobile_config

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/mobile_config"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_mobile_config"

// SemverRegex memvalidasi format versi semver (x.y.z).
var semverRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// Command berisi data yang dibutuhkan untuk mengupdate MobileAppConfig.
type Command struct {
	ID              uuid.UUID       `json:"id"`
	AppVariant      string          `json:"app_variant"`
	Platform        string          `json:"platform"`
	MinVersion      string          `json:"min_version"`
	CurrentVersion  string          `json:"current_version"`
	ForceUpdate     bool            `json:"force_update"`
	MaintenanceMode bool            `json:"maintenance_mode"`
	FeatureFlags    json.RawMessage `json:"feature_flags"`
	APIBaseURL      string          `json:"api_base_url"`
	ThemeConfig     json.RawMessage `json:"theme_config"`
	OfflineConfig   json.RawMessage `json:"offline_config"`
	IsActive        bool            `json:"is_active"`
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_mobile_config.
type Handler struct {
	readRepo  mobile_config.ReadRepository
	writeRepo mobile_config.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr mobile_config.ReadRepository, wr mobile_config.WriteRepository, eb eventbus.EventBus) *Handler {
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

	appVariant := mobile_config.AppVariant(cmd.AppVariant)
	if !mobile_config.ValidAppVariants[appVariant] {
		return mobile_config.ErrInvalidAppVariant
	}

	platform := mobile_config.Platform(cmd.Platform)
	if !mobile_config.ValidPlatforms[platform] {
		return mobile_config.ErrInvalidPlatform
	}

	if cmd.MinVersion != "" && !semverRegex.MatchString(cmd.MinVersion) {
		return mobile_config.ErrInvalidVersion
	}
	if cmd.CurrentVersion == "" || !semverRegex.MatchString(cmd.CurrentVersion) {
		return mobile_config.ErrInvalidVersion
	}

	if cmd.APIBaseURL == "" {
		return mobile_config.ErrEmptyAPIBaseURL
	}
	if len(cmd.APIBaseURL) > 255 {
		return mobile_config.ErrAPIBaseURLTooLong
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get mobile_config for update: %w", err)
	}

	var featureFlags mobile_config.FeatureFlags
	if cmd.FeatureFlags != nil {
		if err := featureFlags.Scan(cmd.FeatureFlags); err != nil {
			return fmt.Errorf("invalid feature_flags: %w", err)
		}
	} else {
		featureFlags = entity.FeatureFlags
	}

	var themeConfig mobile_config.ThemeConfig
	if cmd.ThemeConfig != nil {
		if err := themeConfig.Scan(cmd.ThemeConfig); err != nil {
			return fmt.Errorf("invalid theme_config: %w", err)
		}
	} else {
		themeConfig = entity.ThemeConfig
	}

	var offlineConfig mobile_config.OfflineConfig
	if cmd.OfflineConfig != nil {
		if err := offlineConfig.Scan(cmd.OfflineConfig); err != nil {
			return fmt.Errorf("invalid offline_config: %w", err)
		}
	} else {
		offlineConfig = entity.OfflineConfig
	}

	entity.AppVariant = appVariant
	entity.Platform = platform
	entity.MinVersion = cmd.MinVersion
	entity.CurrentVersion = cmd.CurrentVersion
	entity.ForceUpdate = cmd.ForceUpdate
	entity.MaintenanceMode = cmd.MaintenanceMode
	entity.FeatureFlags = featureFlags
	entity.APIBaseURL = cmd.APIBaseURL
	entity.ThemeConfig = themeConfig
	entity.OfflineConfig = offlineConfig
	entity.IsActive = cmd.IsActive
	entity.UpdatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update mobile_config: %w", err)
	}

	if err := h.eventBus.Publish(ctx, mobile_config.MobileConfigUpdatedEvent{
		TenantID:   s.TenantID,
		CompanyID:  s.CompanyID,
		ID:         entity.ID,
		AppVariant: entity.AppVariant,
		Platform:   entity.Platform,
	}); err != nil {
		return fmt.Errorf("publish mobile_config.updated: %w", err)
	}

	return nil
}
