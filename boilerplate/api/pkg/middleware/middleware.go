package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// Tracer mengembalikan middleware yang membuat OTel span untuk setiap request.
func Tracer(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path)
			defer span.End()

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.user_agent", r.UserAgent()),
			)

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r.WithContext(ctx))

			span.SetAttributes(attribute.Int("http.status_code", ww.Status()))
		})
	}
}

// Logger mengembalikan middleware yang mencatat setiap request menggunakan zerolog.
func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Dur("duration", time.Since(start)).
				Str("remote_addr", r.RemoteAddr).
				Msg("request")
		})
	}
}

// Recoverer mengembalikan middleware yang menangkap panic, mencatat stack trace,
// dan merespons dengan JSON 500.
func Recoverer(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					logger.Error().
						Interface("panic", rec).
						Str("path", r.URL.Path).
						Bytes("stack", stack).
						Msg("panic recovered")
					w.Header().Set("Content-Type", "application/json")
					http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RealIP menetapkan r.RemoteAddr ke IP client yang sebenarnya saat API di belakang reverse proxy.
func RealIP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := r.Header.Get("X-Real-IP"); ip != "" {
				r = r.Clone(r.Context())
				r.RemoteAddr = ip
			} else if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				if idx := strings.IndexByte(forwarded, ','); idx != -1 {
					forwarded = forwarded[:idx]
				}
				ip = strings.TrimSpace(forwarded)
				if ip != "" {
					r = r.Clone(r.Context())
					r.RemoteAddr = ip
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
