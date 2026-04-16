// Package mobile_config mendefinisikan domain configuration untuk mobile app.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package mobile_config

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// AppVariant mendefinisikan varian aplikasi mobile.
type AppVariant string

const (
	AppVariantStudent AppVariant = "student"
	AppVariantParent  AppVariant = "parent"
	AppVariantStaff   AppVariant = "staff"
	AppVariantFull    AppVariant = "full"
)

// Platform mendefinisikan platform target aplikasi.
type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
	PlatformBoth    Platform = "both"
)

// FeatureFlags adalah map dari nama feature ke status aktif/nonaktif.
type FeatureFlags map[string]bool

// ThemeConfig menyimpan konfigurasi tema aplikasi mobile.
type ThemeConfig map[string]any

// OfflineConfig menyimpan konfigurasi offline/sync untuk mobile app.
type OfflineConfig struct {
	SyncInterval    int      `json:"sync_interval"`
	MaxCacheSize    int      `json:"max_cache_size"`
	OfflineFeatures []string `json:"offline_features"`
}

// MobileAppConfig adalah entity utama untuk konfigurasi mobile app.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type MobileAppConfig struct {
	ID              uuid.UUID      `db:"id"`
	AppVariant      AppVariant     `db:"app_variant"`
	Platform        Platform       `db:"platform"`
	MinVersion      string         `db:"min_version"`
	CurrentVersion  string         `db:"current_version"`
	ForceUpdate     bool           `db:"force_update"`
	MaintenanceMode bool           `db:"maintenance_mode"`
	FeatureFlags    FeatureFlags   `db:"feature_flags"`
	APIBaseURL      string         `db:"api_base_url"`
	ThemeConfig     ThemeConfig    `db:"theme_config"`
	OfflineConfig   OfflineConfig  `db:"offline_config"`
	IsActive        bool           `db:"is_active"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	DeletedAt       *time.Time     `db:"deleted_at"`
}

// Scan implements sql.Scanner for FeatureFlags JSONB column.
func (f *FeatureFlags) Scan(val any) error {
	if val == nil {
		*f = make(FeatureFlags)
		return nil
	}
	var bytes []byte
	switch v := val.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		*f = make(FeatureFlags)
		return nil
	}
	return json.Unmarshal(bytes, f)
}

// Scan implements sql.Scanner for ThemeConfig JSONB column.
func (t *ThemeConfig) Scan(val any) error {
	if val == nil {
		*t = make(ThemeConfig)
		return nil
	}
	var bytes []byte
	switch v := val.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		*t = make(ThemeConfig)
		return nil
	}
	return json.Unmarshal(bytes, t)
}

// Scan implements sql.Scanner for OfflineConfig JSONB column.
func (o *OfflineConfig) Scan(val any) error {
	if val == nil {
		*o = OfflineConfig{}
		return nil
	}
	var bytes []byte
	switch v := val.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		*o = OfflineConfig{}
		return nil
	}
	return json.Unmarshal(bytes, o)
}

// Value implements driver.Valuer for OfflineConfig JSONB column.
func (o OfflineConfig) Value() (sql.Null[OfflineConfig], error) {
	return sql.Null[OfflineConfig]{}, nil
}

// ValidAppVariants adalah whitelist untuk AppVariant.
var ValidAppVariants = map[AppVariant]bool{
	AppVariantStudent: true,
	AppVariantParent:  true,
	AppVariantStaff:   true,
	AppVariantFull:    true,
}

// ValidPlatforms adalah whitelist untuk Platform.
var ValidPlatforms = map[Platform]bool{
	PlatformAndroid: true,
	PlatformIOS:     true,
	PlatformBoth:    true,
}

// WriteRepository mendefinisikan operasi write untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *MobileAppConfig) error
	Update(ctx context.Context, s scope.Scope, e *MobileAppConfig) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*MobileAppConfig, error)
	List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string, filters ListFilters) ([]*MobileAppConfig, int, error)
}

// ListFilters menyimpan filter opsional untuk list query.
type ListFilters struct {
	AppVariant AppVariant
	Platform   Platform
}
