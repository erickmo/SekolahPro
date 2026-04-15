package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	NATS       NATSConfig
	Otel       OtelConfig
	JWT        JWTConfig
	HTTPServer HTTPServerConfig
	Tenant     TenantConfig
}

// TenantMode constants — dipakai oleh pkg/tenant dan infrastructure.
const (
	TenantModeSingle = "single" // Satu org per deployment; TenantID dari config.
	TenantModeMulti  = "multi"  // Banyak org; TenantID dari JWT claims.
)

// TenantConfig menyimpan konfigurasi tenancy dan hierarki organisasi.
//
// Hierarki 4 level: Tenant (L1) → Company (L2) → Branch (L3, opsional) → Warehouse (L4, opsional)
//
// Mode "single": semua level yang dipakai harus dikonfigurasi di sini.
// Mode "multi":  semua level dibaca dari JWT claims per request.
type TenantConfig struct {
	// Mode menentukan apakah deployment ini single-tenant atau multi-tenant.
	// Valid: "single" | "multi"
	Mode string `mapstructure:"TENANT_MODE"`
	// TenantID adalah UUID top-level namespace. Wajib jika Mode = "single".
	TenantID string `mapstructure:"TENANT_ID"`
	// CompanyID adalah UUID legal entity. Wajib jika Mode = "single".
	CompanyID string `mapstructure:"COMPANY_ID"`
	// BranchID adalah UUID cabang operasional. Opsional meski Mode = "single".
	BranchID string `mapstructure:"BRANCH_ID"`
	// WarehouseID adalah UUID lokasi fisik penyimpanan. Opsional.
	// Lebih baik kirim via request body daripada di token/config.
	WarehouseID string `mapstructure:"WAREHOUSE_ID"`
}

type AppConfig struct {
	Name     string `mapstructure:"APP_NAME"`
	Env      string `mapstructure:"APP_ENV"`
	HTTPPort string `mapstructure:"HTTP_PORT"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"DATABASE_URL"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

type RedisConfig struct {
	URL        string `mapstructure:"REDIS_URL"`
	TTLSeconds int    `mapstructure:"REDIS_TTL_SECONDS"`
}

type NATSConfig struct {
	URL        string `mapstructure:"NATS_URL"`
	StreamName string `mapstructure:"NATS_STREAM_NAME"`
	UseNATS    bool   `mapstructure:"USE_NATS"`
}

type OtelConfig struct {
	ExporterEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	PrometheusPort   string `mapstructure:"PROMETHEUS_PORT"`
}

type JWTConfig struct {
	Secret            string `mapstructure:"JWT_SECRET"`
	ExpiryHours       int    `mapstructure:"JWT_EXPIRY_HOURS"`
	Issuer            string `mapstructure:"JWT_ISSUER"`
	RefreshMultiplier int    `mapstructure:"JWT_REFRESH_MULTIPLIER"`
}

type HTTPServerConfig struct {
	ReadTimeoutSeconds       int `mapstructure:"HTTP_READ_TIMEOUT_SECONDS"`
	WriteTimeoutSeconds      int `mapstructure:"HTTP_WRITE_TIMEOUT_SECONDS"`
	IdleTimeoutSeconds       int `mapstructure:"HTTP_IDLE_TIMEOUT_SECONDS"`
	MiddlewareTimeoutSeconds int `mapstructure:"HTTP_MIDDLEWARE_TIMEOUT_SECONDS"`
	MaxPageSize              int `mapstructure:"HTTP_MAX_PAGE_SIZE"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")

	viper.SetDefault("APP_NAME", "boilerplate")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")
	viper.SetDefault("REDIS_TTL_SECONDS", 300)
	viper.SetDefault("NATS_STREAM_NAME", "APP_EVENTS")
	viper.SetDefault("USE_NATS", false)
	viper.SetDefault("PROMETHEUS_PORT", "9090")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("JWT_ISSUER", "boilerplate-api")
	viper.SetDefault("JWT_REFRESH_MULTIPLIER", 7)
	viper.SetDefault("HTTP_READ_TIMEOUT_SECONDS", 15)
	viper.SetDefault("HTTP_WRITE_TIMEOUT_SECONDS", 15)
	viper.SetDefault("HTTP_IDLE_TIMEOUT_SECONDS", 60)
	viper.SetDefault("HTTP_MIDDLEWARE_TIMEOUT_SECONDS", 30)
	viper.SetDefault("HTTP_MAX_PAGE_SIZE", 1000)
	viper.SetDefault("TENANT_MODE", TenantModeSingle)

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return nil, fmt.Errorf("gagal membaca config file: %w", err)
		}
	}

	cfg := &Config{}
	cfg.App = AppConfig{
		Name:     viper.GetString("APP_NAME"),
		Env:      viper.GetString("APP_ENV"),
		HTTPPort: viper.GetString("HTTP_PORT"),
		LogLevel: viper.GetString("LOG_LEVEL"),
	}
	cfg.Database = DatabaseConfig{
		URL:             viper.GetString("DATABASE_URL"),
		MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
	}
	cfg.Redis = RedisConfig{
		URL:        viper.GetString("REDIS_URL"),
		TTLSeconds: viper.GetInt("REDIS_TTL_SECONDS"),
	}
	cfg.NATS = NATSConfig{
		URL:        viper.GetString("NATS_URL"),
		StreamName: viper.GetString("NATS_STREAM_NAME"),
		UseNATS:    viper.GetBool("USE_NATS"),
	}
	cfg.Otel = OtelConfig{
		ExporterEndpoint: viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
		PrometheusPort:   viper.GetString("PROMETHEUS_PORT"),
	}
	cfg.JWT = JWTConfig{
		Secret:            viper.GetString("JWT_SECRET"),
		ExpiryHours:       viper.GetInt("JWT_EXPIRY_HOURS"),
		Issuer:            viper.GetString("JWT_ISSUER"),
		RefreshMultiplier: viper.GetInt("JWT_REFRESH_MULTIPLIER"),
	}
	cfg.HTTPServer = HTTPServerConfig{
		ReadTimeoutSeconds:       viper.GetInt("HTTP_READ_TIMEOUT_SECONDS"),
		WriteTimeoutSeconds:      viper.GetInt("HTTP_WRITE_TIMEOUT_SECONDS"),
		IdleTimeoutSeconds:       viper.GetInt("HTTP_IDLE_TIMEOUT_SECONDS"),
		MiddlewareTimeoutSeconds: viper.GetInt("HTTP_MIDDLEWARE_TIMEOUT_SECONDS"),
		MaxPageSize:              viper.GetInt("HTTP_MAX_PAGE_SIZE"),
	}
	cfg.Tenant = TenantConfig{
		Mode:        viper.GetString("TENANT_MODE"),
		TenantID:    viper.GetString("TENANT_ID"),
		CompanyID:   viper.GetString("COMPANY_ID"),
		BranchID:    viper.GetString("BRANCH_ID"),
		WarehouseID: viper.GetString("WAREHOUSE_ID"),
	}

	if err := validateTenantConfig(cfg.Tenant); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateTenantConfig memvalidasi konfigurasi tenancy saat startup.
func validateTenantConfig(t TenantConfig) error {
	switch t.Mode {
	case TenantModeSingle:
		if strings.TrimSpace(t.TenantID) == "" {
			return errors.New("TENANT_ID wajib diisi saat TENANT_MODE=single")
		}
		if _, err := uuidParse(t.TenantID); err != nil {
			return fmt.Errorf("TENANT_ID harus UUID yang valid: %w", err)
		}
		if strings.TrimSpace(t.CompanyID) == "" {
			return errors.New("COMPANY_ID wajib diisi saat TENANT_MODE=single")
		}
		if _, err := uuidParse(t.CompanyID); err != nil {
			return fmt.Errorf("COMPANY_ID harus UUID yang valid: %w", err)
		}
		if t.BranchID != "" {
			if _, err := uuidParse(t.BranchID); err != nil {
				return fmt.Errorf("BRANCH_ID harus UUID yang valid: %w", err)
			}
		}
		if t.WarehouseID != "" {
			if _, err := uuidParse(t.WarehouseID); err != nil {
				return fmt.Errorf("WAREHOUSE_ID harus UUID yang valid: %w", err)
			}
		}
	case TenantModeMulti:
		// Semua level datang dari JWT per request, tidak perlu di config
	default:
		return fmt.Errorf("TENANT_MODE tidak valid: %q (valid: single|multi)", t.Mode)
	}
	return nil
}

// uuidParse adalah thin wrapper agar package config tidak import github.com/google/uuid.
// Cukup validasi format UUID v4 tanpa dependency tambahan.
func uuidParse(s string) (string, error) {
	if len(s) != 36 {
		return "", fmt.Errorf("panjang UUID tidak valid")
	}
	// Validasi format 8-4-4-4-12
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return "", fmt.Errorf("karakter UUID tidak valid di posisi %d", i)
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return "", fmt.Errorf("karakter UUID tidak valid di posisi %d", i)
			}
		}
	}
	return s, nil
}
