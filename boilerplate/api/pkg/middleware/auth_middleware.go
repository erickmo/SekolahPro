// Package middleware menyediakan HTTP middleware untuk API.
package middleware

import (
	"context"
	"net/http"
	"strings"

	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
)

// contextKey adalah tipe untuk context keys agar tidak collision.
type contextKey string

const (
	// ContextKeyClaims adalah key untuk menyimpan JWT claims di context.
	ContextKeyClaims contextKey = "jwt_claims"
	// ContextKeyCompanyID adalah key untuk menyimpan Company ID di context.
	ContextKeyCompanyID contextKey = "company_id"
)

// RequireAuth adalah middleware yang memvalidasi JWT Bearer token.
// Claims akan diinjeksikan ke request context jika token valid.
func RequireAuth(jwtService *jwtpkg.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.ValidateClaims(parts[1])
			if err != nil {
				http.Error(w, `{"error":"token invalid or expired"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
			if claims.CompanyID != nil {
				ctx = context.WithValue(ctx, ContextKeyCompanyID, *claims.CompanyID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole adalah middleware yang memastikan user memiliki salah satu role yang diizinkan.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ContextKeyClaims).(*jwtpkg.Claims)
			if !ok || claims == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if _, permitted := allowed[claims.Role]; !permitted {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromContext mengambil JWT claims dari context.
func ClaimsFromContext(ctx context.Context) (*jwtpkg.Claims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*jwtpkg.Claims)
	return claims, ok
}

// CompanyIDFromContext mengambil Company ID dari context.
func CompanyIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextKeyCompanyID).(string)
	return v, ok && v != ""
}
