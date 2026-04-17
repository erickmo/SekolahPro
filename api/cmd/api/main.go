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
	loginhdlr "github.com/yourorg/boilerplate/internal/command/login"
	createuser "github.com/yourorg/boilerplate/internal/command/register_user"
	changepwd "github.com/yourorg/boilerplate/internal/command/change_password"
	createrole "github.com/yourorg/boilerplate/internal/command/create_role"
	assignrole "github.com/yourorg/boilerplate/internal/command/assign_role"
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

	"github.com/yourorg/boilerplate/internal/domain/academic_year"
	"github.com/yourorg/boilerplate/internal/domain/class_room"
	"github.com/yourorg/boilerplate/internal/domain/nasabah"
	"github.com/yourorg/boilerplate/internal/domain/product"
	"github.com/yourorg/boilerplate/internal/domain/product_category"
	"github.com/yourorg/boilerplate/internal/domain/produk_akad"
	"github.com/yourorg/boilerplate/internal/domain/rekening"
	"github.com/yourorg/boilerplate/internal/domain/student"
	"github.com/yourorg/boilerplate/internal/domain/student_address"
	"github.com/yourorg/boilerplate/internal/domain/student_admission"
	"github.com/yourorg/boilerplate/internal/domain/student_class_placement"
	"github.com/yourorg/boilerplate/internal/domain/student_document"
	"github.com/yourorg/boilerplate/internal/domain/student_guardian"
	"github.com/yourorg/boilerplate/internal/domain/simpanan_pokok_wajib"
	"github.com/yourorg/boilerplate/internal/domain/tabungan"
	"github.com/yourorg/boilerplate/internal/domain/deposito"
	"github.com/yourorg/boilerplate/internal/domain/teacher"

	// Sprint 4 — Kurikulum + Pembiayaan
	"github.com/yourorg/boilerplate/internal/domain/curriculum"
	"github.com/yourorg/boilerplate/internal/domain/subject"
	"github.com/yourorg/boilerplate/internal/domain/academic_calendar"
	"github.com/yourorg/boilerplate/internal/domain/teaching_schedule"
	"github.com/yourorg/boilerplate/internal/domain/lesson_plan"
	"github.com/yourorg/boilerplate/internal/domain/teaching_journal"
	"github.com/yourorg/boilerplate/internal/domain/pinjaman"
	"github.com/yourorg/boilerplate/internal/domain/angsuran"
	"github.com/yourorg/boilerplate/internal/domain/denda"
	"github.com/yourorg/boilerplate/internal/domain/jaminan"

	// Sprint 5 — Kehadiran & Nilai + Transaksi
	"github.com/yourorg/boilerplate/internal/domain/daily_attendance"
	"github.com/yourorg/boilerplate/internal/domain/academic_record"
	"github.com/yourorg/boilerplate/internal/domain/subject_grade"
	"github.com/yourorg/boilerplate/internal/domain/exam_assessment"
	"github.com/yourorg/boilerplate/internal/domain/health_record"
	"github.com/yourorg/boilerplate/internal/domain/discipline"
	"github.com/yourorg/boilerplate/internal/domain/achievement"
	"github.com/yourorg/boilerplate/internal/domain/extracurricular"
	"github.com/yourorg/boilerplate/internal/domain/transaksi"
	"github.com/yourorg/boilerplate/internal/domain/teller_session"
	"github.com/yourorg/boilerplate/internal/domain/money_denomination"
	"github.com/yourorg/boilerplate/internal/domain/kas"
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

			// Auth & User/Role handlers (Sprint 1)
			provideAuthHandler,
			provideUserHandler,
			provideRoleHandler,

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
	// Contoh domain (bisa dihapus di production)
	registerVernonDomain(db, registry, eb, logger, &product_category.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &product.Descriptor{})

	// Foundation domains (ADR-010, 012, 011) — urutan: sumber dulu, consumer terakhir
	registerVernonDomain(db, registry, eb, logger, &academic_year.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &teacher.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &class_room.Descriptor{}) // autoloads academic_year + teacher

	// Koperasi core domains (ADR-K003, K001, K002) — Sprint 2
	registerVernonDomain(db, registry, eb, logger, &nasabah.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &produk_akad.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &rekening.Descriptor{}) // autoloads nasabah + produk_akad

	// Student domains (ADR-S001, S007, S003, S010, S014, S016) — Sprint 3
	registerVernonDomain(db, registry, eb, logger, &student.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &student_address.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &student_guardian.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &student_document.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &student_class_placement.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &student_admission.Descriptor{})

	// Simpanan domains (ADR-K004, K005, K006) — Sprint 3
	registerVernonDomain(db, registry, eb, logger, &simpanan_pokok_wajib.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &tabungan.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &deposito.Descriptor{})

	// Kurikulum & Akademik domains (ADR-S019, S020, S023, S021, S024, S025) — Sprint 4
	registerVernonDomain(db, registry, eb, logger, &curriculum.CurriculaDescriptor{})
	registerVernonDomain(db, registry, eb, logger, &curriculum.LearningOutcomesDescriptor{})
	registerVernonDomain(db, registry, eb, logger, &subject.SubjectsDescriptor{})          // autoloads curriculum
	registerVernonDomain(db, registry, eb, logger, &subject.SubjectConfigurationsDescriptor{})
	registerVernonDomain(db, registry, eb, logger, &academic_calendar.AcademicCalendarEventsDescriptor{}) // autoloads academic_year
	registerVernonDomain(db, registry, eb, logger, &teaching_schedule.TimeSlotDescriptor{})
	registerVernonDomain(db, registry, eb, logger, &teaching_schedule.ScheduleEntryDescriptor{}) // autoloads subject + teacher + class_room
	registerVernonDomain(db, registry, eb, logger, &lesson_plan.LessonPlanDescriptor{})       // autoloads subject + teacher
	registerVernonDomain(db, registry, eb, logger, &lesson_plan.LessonPlanAttachmentDescriptor{})
	registerVernonDomain(db, registry, eb, logger, &teaching_journal.TeachingJournalDescriptor{})  // autoloads schedule + teacher
	registerVernonDomain(db, registry, eb, logger, &teaching_journal.JournalSessionAttendanceDescriptor{})

	// Pembiayaan domains (ADR-K007, K008, K009, K010) — Sprint 4
	registerVernonDomain(db, registry, eb, logger, &pinjaman.Descriptor{})          // autoloads nasabah + rekening + produk_akad
	registerVernonDomain(db, registry, eb, logger, &angsuran.Descriptor{})          // autoloads pinjaman
	registerVernonDomain(db, registry, eb, logger, &denda.Descriptor{})             // autoloads pinjaman + angsuran
	registerVernonDomain(db, registry, eb, logger, &jaminan.Descriptor{})           // autoloads pinjaman

	// Kehadiran & Penilaian domains (ADR-S008, S004, S011, S022, S005, S012, S013, S015) — Sprint 5
	registerVernonDomain(db, registry, eb, logger, &daily_attendance.Descriptor{})  // autoloads student + class_room
	registerVernonDomain(db, registry, eb, logger, &academic_record.Descriptor{})   // autoloads student + academic_year
	registerVernonDomain(db, registry, eb, logger, &subject_grade.Descriptor{})     // autoloads student + subject + teacher
	registerVernonDomain(db, registry, eb, logger, &exam_assessment.Descriptor{})   // autoloads subject + teacher + class_room
	registerVernonDomain(db, registry, eb, logger, &health_record.Descriptor{})     // autoloads student
	registerVernonDomain(db, registry, eb, logger, &discipline.Descriptor{})        // autoloads student + teacher
	registerVernonDomain(db, registry, eb, logger, &achievement.Descriptor{})       // autoloads student
	registerVernonDomain(db, registry, eb, logger, &extracurricular.Descriptor{})   // autoloads student + teacher

	// Transaksi & Operasional domains (ADR-K011, K012, K013, K014) — Sprint 5
	registerVernonDomain(db, registry, eb, logger, &transaksi.Descriptor{})         // autoloads rekening
	registerVernonDomain(db, registry, eb, logger, &teller_session.Descriptor{})
	registerVernonDomain(db, registry, eb, logger, &money_denomination.Descriptor{}) // autoloads teller_session
	registerVernonDomain(db, registry, eb, logger, &kas.Descriptor{})                // autoloads teller_session
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

	// Example domain
	repo := database.NewExampleRepository(db)
	commandbus.Register(bus, createexample.NewHandler(repo, eb))
	commandbus.Register(bus, updateexample.NewHandler(repo, repo, eb))
	commandbus.Register(bus, deleteexample.NewHandler(repo))

	// Sprint 1: Auth & User domain
	userWriteRepo := database.NewUserRepository(db)
	userReadRepo := database.NewUserRepository(db)
	roleReadRepo := database.NewRoleRepository(db)
	userRoleWriteRepo := database.NewRoleRepository(db)
	userRoleReadRepo := database.NewRoleRepository(db)

	commandbus.Register(bus, createuser.NewHandler(userWriteRepo, userReadRepo, eb))
	commandbus.Register(bus, createrole.NewHandler(database.NewRoleRepository(db), roleReadRepo, eb))
	commandbus.Register(bus, assignrole.NewHandler(userRoleWriteRepo, userRoleReadRepo, eb))
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

// ── Sprint 1 handler providers ───────────────────────────────────────────────

func provideAuthHandler(
	db *sqlx.DB,
	jwtSvc *jwtpkg.Service,
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
) *deliveryhttp.AuthHandler {
	userReadRepo := database.NewUserRepository(db)
	userWriteRepo := database.NewUserRepository(db)
	loginHdlr := loginhdlr.NewHandler(userReadRepo, userWriteRepo, jwtSvc)
	changePwdHdlr := changepwd.NewHandler(userReadRepo, userWriteRepo)
	return deliveryhttp.NewAuthHandler(loginHdlr, changePwdHdlr, userReadRepo, jwtSvc, cb, qb)
}

func provideUserHandler(
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
	db *sqlx.DB,
) *deliveryhttp.UserHandler {
	userReadRepo := database.NewUserRepository(db)
	userWriteRepo := database.NewUserRepository(db)
	return deliveryhttp.NewUserHandler(userReadRepo, userWriteRepo, cb, qb)
}

func provideRoleHandler(
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
	db *sqlx.DB,
) *deliveryhttp.RoleHandler {
	roleReadRepo := database.NewRoleRepository(db)
	userRoleReadRepo := database.NewRoleRepository(db)
	return deliveryhttp.NewRoleHandler(roleReadRepo, userRoleReadRepo, cb, qb)
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
