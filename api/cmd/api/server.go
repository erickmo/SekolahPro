package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.uber.org/fx"

	"github.com/yourorg/boilerplate/infrastructure/config"
	deliveryhttp "github.com/yourorg/boilerplate/internal/delivery/http"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/middleware"
	"github.com/yourorg/boilerplate/pkg/scope"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

func newRouter(
	cfg *config.Config,
	logger zerolog.Logger,
	jwtSvc *jwtpkg.Service,
	scopeResolver scope.Resolver,
	exampleHandler *deliveryhttp.ExampleHandler,
	vernonRegistry *vernon.Registry,
) chi.Router {
	r := chi.NewRouter()

	// ── Global Middleware ──────────────────────────────────────────────────────
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Tracer(cfg.App.Name))
	r.Use(middleware.Recoverer(logger))
	r.Use(chimiddleware.Timeout(time.Duration(cfg.HTTPServer.MiddlewareTimeoutSeconds) * time.Second))

	// ── Public Routes (tidak butuh auth) ──────────────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","tenant_mode":"` + cfg.Tenant.Mode + `"}`))
	})

	// /metrics — Prometheus scrape endpoint.
	// OTel Prometheus exporter mendaftarkan metrics ke default registry secara otomatis.
	r.Handle("/metrics", promhttp.Handler())

	// /swagger/ — Swagger UI (hanya aktif jika docs/ sudah digenerate via `make swagger`)
	r.Handle("/swagger/*", http.StripPrefix("/swagger", http.FileServer(http.Dir("docs"))))

	// ── Protected Routes (butuh auth + scope organisasi) ──────────────────────
	// Urutan: RequireAuth → ResolveScope (fail-open) → RequireScope (fail-closed) → handler
	//
	// ResolveScope: inject Scope ke context bila berhasil; lanjutkan walau gagal.
	// RequireScope: reject jika level yang diminta tidak tersedia di context.
	//
	// Mode "single": ResolveScope selalu inject Scope dari config.
	// Mode "multi":  ResolveScope baca semua level dari JWT claims.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtSvc))
		r.Use(scope.ResolveScope(scopeResolver))
		r.Use(scope.RequireScope(scope.LevelTenant, scope.LevelCompany))

		exampleHandler.RegisterRoutes(r)

			// Vernon domains — setiap domain terdaftar di-mount otomatis.
			// Route: GET|POST /{domain}/, GET|PUT|PATCH|DELETE /{domain}/{id}
			vernonRegistry.MountRoutes(r)
	})

	return r
}

func startServer(lc fx.Lifecycle, r chi.Router, cfg *config.Config, logger zerolog.Logger) {
	srv := &http.Server{
		Addr:         ":" + cfg.App.HTTPPort,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.HTTPServer.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTPServer.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.HTTPServer.IdleTimeoutSeconds) * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info().
				Str("port", cfg.App.HTTPPort).
				Str("tenant_mode", cfg.Tenant.Mode).
				Msg("starting HTTP server")
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal().Err(err).Msg("HTTP server error")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info().Msg("shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
