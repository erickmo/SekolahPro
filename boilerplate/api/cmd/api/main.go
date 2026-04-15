// @title           Boilerplate API
// @version         1.0
// @description     Go Clean Architecture API — CQRS + Vernon Hybrid Pattern
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@yourorg.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Format: "Bearer {token}"

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	"github.com/yourorg/boilerplate/infrastructure/cache"
	"github.com/yourorg/boilerplate/infrastructure/config"
	"github.com/yourorg/boilerplate/infrastructure/database"
	"github.com/yourorg/boilerplate/infrastructure/telemetry"
	deliveryhttp "github.com/yourorg/boilerplate/internal/delivery/http"
	"github.com/yourorg/boilerplate/internal/eventhandler"

	createexample "github.com/yourorg/boilerplate/internal/command/create_example"
	deleteexample "github.com/yourorg/boilerplate/internal/command/delete_example"
	updateexample "github.com/yourorg/boilerplate/internal/command/update_example"
	getexamplebyid "github.com/yourorg/boilerplate/internal/query/get_example_by_id"
	listexamples "github.com/yourorg/boilerplate/internal/query/list_examples"

	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/middleware"
	"github.com/yourorg/boilerplate/pkg/querybus"
	"github.com/yourorg/boilerplate/pkg/scope"
	"github.com/yourorg/boilerplate/pkg/vernon"
	"github.com/yourorg/boilerplate/pkg/vernonsync"

	"github.com/yourorg/boilerplate/internal/domain/product"
	"github.com/yourorg/boilerplate/internal/domain/product_category"
)

func main() {
	app := fx.New(
		fx.Provide(
			// Infrastructure
			provideConfig,
			provideLogger,
			database.NewDB,
			provideRedis,
			provideCache,
			provideEventBus,
			provideTelemetry,

			// Buses (CQRS pattern — untuk domain sederhana)
			commandbus.New,
			querybus.New,

			// Auth
			provideJWT,

			// Scope resolver — dipilih berdasarkan TENANT_MODE di config
			provideScopeResolver,

			// HTTP Handlers (CQRS pattern)
			deliveryhttp.NewExampleHandler,

			// Vernon pattern — untuk domain dengan banyak relasi/JOIN
			provideVernonRegistry,
			provideVernonSyncEngine,

			// Router
			newRouter,
		),
		fx.Invoke(
			initTelemetry,
			registerCommandHandlers,
			registerQueryHandlers,
			registerEventHandlers,
			registerVernonDomains,
			startVernonSync,
			startServer,
		),
	)

	app.Run()
}

// ── Infrastructure providers ──────────────────────────────────────────────────

func provideConfig() (*config.Config, error) {
	return config.Load()
}

func provideLogger(cfg *config.Config) zerolog.Logger {
	if cfg.App.Env == "development" {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
			With().Timestamp().Logger()
	}
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func provideRedis(cfg *config.Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opt), nil
}

func provideCache(client *redis.Client, cfg *config.Config) cache.Cache {
	return cache.NewRedisCache(client, cfg.App.Name)
}

// provideEventBus memilih backend event bus berdasarkan konfigurasi.
// InMemory untuk dev/test, NATS JetStream untuk production.
func provideEventBus(cfg *config.Config, logger zerolog.Logger) (eventbus.EventBus, error) {
	if cfg.NATS.UseNATS {
		logger.Info().Str("nats_url", cfg.NATS.URL).Msg("menggunakan NATS JetStream event bus")
		return eventbus.NewNATSEventBus(eventbus.NATSConfig{
			URL:        cfg.NATS.URL,
			StreamName: cfg.NATS.StreamName,
		})
	}
	logger.Info().Msg("menggunakan InMemory event bus (dev mode)")
	return eventbus.NewInMemoryEventBus()
}

func provideTelemetry(cfg *config.Config) (*telemetry.Provider, error) {
	// Telemetry diinisialisasi di initTelemetry (invoke) agar FX lifecycle bisa di-register.
	// Provider di sini hanya menyimpan config reference sementara.
	_ = cfg
	return nil, nil //nolint:nilnil
}

// provideTelemetryProvider adalah provider yang dipakai jika OTel dibutuhkan oleh package lain.
// Saat ini main menginisialisasi OTel via global otel.SetTracerProvider di initTelemetry.
func provideJWT(cfg *config.Config) *jwtpkg.Service {
	return jwtpkg.NewService(
		cfg.JWT.Secret,
		cfg.JWT.ExpiryHours,
		cfg.JWT.Issuer,
		cfg.JWT.RefreshMultiplier,
	)
}

// ── Telemetry lifecycle ───────────────────────────────────────────────────────

// initTelemetry menginisialisasi OTel SDK dan mendaftarkan shutdown ke lifecycle FX.
func initTelemetry(lc fx.Lifecycle, cfg *config.Config, logger zerolog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p, err := telemetry.Init(ctx, cfg.App.Name, cfg.Otel.ExporterEndpoint)
			if err != nil {
				logger.Warn().Err(err).Msg("telemetry init gagal — traces tidak dikirim ke Jaeger")
				return nil // non-fatal: app tetap jalan tanpa telemetry
			}
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					return p.Shutdown(ctx)
				},
			})
			logger.Info().Str("endpoint", cfg.Otel.ExporterEndpoint).Msg("OTel telemetry aktif")
			return nil
		},
	})
}

// ── Scope resolver ────────────────────────────────────────────────────────────

func provideScopeResolver(cfg *config.Config, jwtSvc *jwtpkg.Service) (scope.Resolver, error) {
	switch cfg.Tenant.Mode {
	case config.TenantModeSingle:
		s, err := buildScopeFromConfig(cfg.Tenant)
		if err != nil {
			return nil, err
		}
		log.Info().
			Str("mode", cfg.Tenant.Mode).
			Str("tenant_id", s.TenantID.String()).
			Str("company_id", s.CompanyID.String()).
			Msg("scope resolver: ConfigScopeResolver (single-tenant)")
		return scope.NewConfigScopeResolver(s), nil

	case config.TenantModeMulti:
		log.Info().
			Str("mode", cfg.Tenant.Mode).
			Msg("scope resolver: JWTScopeResolver (multi-tenant)")
		_ = jwtSvc
		return scope.NewJWTScopeResolver(middleware.ContextKeyClaims), nil

	default:
		return nil, fmt.Errorf("TENANT_MODE tidak valid: %q", cfg.Tenant.Mode)
	}
}

func buildScopeFromConfig(t config.TenantConfig) (scope.Scope, error) {
	tenantID, err := uuid.Parse(t.TenantID)
	if err != nil {
		return scope.Scope{}, fmt.Errorf("TENANT_ID tidak valid: %w", err)
	}
	companyID, err := uuid.Parse(t.CompanyID)
	if err != nil {
		return scope.Scope{}, fmt.Errorf("COMPANY_ID tidak valid: %w", err)
	}
	s := scope.Scope{TenantID: tenantID, CompanyID: companyID}
	if t.BranchID != "" {
		id, err := uuid.Parse(t.BranchID)
		if err != nil {
			return scope.Scope{}, fmt.Errorf("BRANCH_ID tidak valid: %w", err)
		}
		s.BranchID = &id
	}
	if t.WarehouseID != "" {
		id, err := uuid.Parse(t.WarehouseID)
		if err != nil {
			return scope.Scope{}, fmt.Errorf("WAREHOUSE_ID tidak valid: %w", err)
		}
		s.WarehouseID = &id
	}
	return s, nil
}

// ── Vernon hybrid pattern providers ──────────────────────────────────────────

func provideVernonRegistry(logger zerolog.Logger) *vernon.Registry {
	return vernon.NewRegistry(logger)
}

func provideVernonSyncEngine(
	db *sqlx.DB,
	registry *vernon.Registry,
	eb eventbus.EventBus,
	logger zerolog.Logger,
) *vernonsync.SyncEngine {
	return vernonsync.NewSyncEngine(db, registry, eb, logger)
}

// registerVernonDomains mendaftarkan semua domain Vernon ke Registry.
//
// Urutan pendaftaran penting: domain sumber (product_categories) harus didaftarkan
// sebelum consumer (products) agar reverse dependency map terbangun dengan benar.
func registerVernonDomains(
	db *sqlx.DB,
	registry *vernon.Registry,
	eb eventbus.EventBus,
	logger zerolog.Logger,
) {
	registerVernonDomain(db, registry, eb, logger, &product_category.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &product.Descriptor{})
}

// registerVernonDomain adalah helper DRY untuk mendaftarkan satu domain Vernon.
func registerVernonDomain(
	db *sqlx.DB,
	registry *vernon.Registry,
	eb eventbus.EventBus,
	logger zerolog.Logger,
	desc vernon.DomainDescriptor,
) {
	repo := vernon.NewBaseRepository(db, desc.TableName())
	svc := vernon.NewBaseService(db, repo, desc, eb, logger)
	hdlr := vernon.NewBaseHandler(svc, logger)
	registry.Register(desc.TableName(), &vernon.RegisteredDomain{
		Descriptor: desc,
		Handler:    hdlr,
		Service:    svc,
	})
}

func startVernonSync(lc fx.Lifecycle, engine *vernonsync.SyncEngine) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return engine.Subscribe(ctx)
		},
	})
}

// ── CQRS handler registration ─────────────────────────────────────────────────

// registerCommandHandlers mendaftarkan semua command handler ke command bus.
func registerCommandHandlers(
	bus *commandbus.CommandBus,
	db *sqlx.DB,
	eb eventbus.EventBus,
	logger zerolog.Logger,
) {
	logger.Info().Msg("registering command handlers")
	repo := database.NewExampleRepository(db)
	commandbus.Register(bus, createexample.NewHandler(repo, eb))
	commandbus.Register(bus, updateexample.NewHandler(repo, repo, eb))
	commandbus.Register(bus, deleteexample.NewHandler(repo))
}

// registerQueryHandlers mendaftarkan semua query handler ke query bus.
func registerQueryHandlers(
	bus *querybus.QueryBus,
	db *sqlx.DB,
	logger zerolog.Logger,
) {
	logger.Info().Msg("registering query handlers")
	repo := database.NewExampleRepository(db)
	querybus.Register(bus, getexamplebyid.NewHandler(repo))
	querybus.Register(bus, listexamples.NewHandler(repo))
}

// registerEventHandlers mendaftarkan semua event handler dan mengelola lifecycle router.
func registerEventHandlers(
	lc fx.Lifecycle,
	eb eventbus.EventBus,
	logger zerolog.Logger,
) {
	logger.Info().Msg("registering event handlers")

	h := eventhandler.NewExampleEventHandler(logger)
	if err := h.RegisterHandlers(eb); err != nil {
		logger.Fatal().Err(err).Msg("gagal mendaftarkan event handlers")
	}

	runner, ok := eb.(eventbus.RouterRunner)
	if !ok {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := runner.StartRouter(context.Background()); err != nil {
					logger.Error().Err(err).Msg("event router error")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return runner.StopRouter(ctx)
		},
	})
}
