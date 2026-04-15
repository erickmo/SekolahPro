// Package tenant menyediakan abstraksi untuk multi-tenant dan single-tenant.
//
// Desain:
//   - Domain layer (internal/domain/) TIDAK boleh import package ini.
//   - Application layer (internal/command/, internal/query/) membaca TenantID
//     dari context lalu meneruskannya sebagai parameter eksplisit ke repository.
//   - Repository menerima tenantID sebagai parameter, bukan dari context.
package tenant

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// Mode adalah konstanta untuk mode tenancy yang didukung.
const (
	ModeSingle = "single" // Satu organisasi per deployment. TenantID dari config.
	ModeMulti  = "multi"  // Banyak organisasi. TenantID dari JWT claims.
)

// ErrMissing dikembalikan saat tenant ID tidak ditemukan di context.
var ErrMissing = errors.New("tenant ID tidak ditemukan — pastikan middleware ResolveTenant dipasang")

// contextKey adalah tipe private untuk menghindari collision di context.
type contextKey struct{}

// WithTenantID menyimpan tenant ID ke dalam context.
func WithTenantID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// TenantIDFromContext mengambil tenant ID dari context.
// Mengembalikan (uuid.Nil, false) jika tidak ada.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(contextKey{}).(uuid.UUID)
	if !ok || v == uuid.Nil {
		return uuid.Nil, false
	}
	return v, true
}

// MustTenantIDFromContext mengambil tenant ID dari context.
// Panic jika tidak ada — gunakan hanya di tempat yang dijamin sudah ada middleware RequireTenant.
func MustTenantIDFromContext(ctx context.Context) uuid.UUID {
	v, ok := TenantIDFromContext(ctx)
	if !ok {
		panic("tenant ID missing from context — RequireTenant middleware belum dipasang")
	}
	return v
}

// ─── Resolver Interface ────────────────────────────────────────────────────────

// Resolver mendefinisikan cara mengambil tenant ID dari HTTP request.
// Implementasi dipilih berdasarkan TENANT_MODE di config (via Uber FX).
type Resolver interface {
	// Resolve mengambil tenant ID dari request. Mengembalikan error jika tidak bisa.
	Resolve(r *http.Request) (uuid.UUID, error)
}

// ─── ConfigResolver (untuk mode "single") ─────────────────────────────────────

// ConfigResolver menggunakan tenant ID fixed dari konfigurasi aplikasi.
// Dipakai pada deployment single-tenant.
type ConfigResolver struct {
	tenantID uuid.UUID
}

// NewConfigResolver membuat ConfigResolver dengan tenant ID yang sudah di-parse.
func NewConfigResolver(tenantID uuid.UUID) *ConfigResolver {
	return &ConfigResolver{tenantID: tenantID}
}

// Resolve selalu mengembalikan tenant ID yang sama (dari config).
func (r *ConfigResolver) Resolve(_ *http.Request) (uuid.UUID, error) {
	return r.tenantID, nil
}

// ─── JWTResolver (untuk mode "multi") ─────────────────────────────────────────

// claimsReader mendefinisikan minimal interface untuk membaca TenantID dari JWT claims.
// Ini menghindari circular import dengan pkg/jwt.
type claimsReader interface {
	GetTenantID() string
}

// JWTResolver mengambil tenant ID dari JWT claims yang sudah ada di context.
// Harus dipasang SETELAH middleware RequireAuth.
type JWTResolver struct {
	claimsKey any // context key yang dipakai middleware auth
}

// NewJWTResolver membuat JWTResolver.
// claimsKey adalah context key yang sama dengan yang dipakai RequireAuth middleware.
func NewJWTResolver(claimsKey any) *JWTResolver {
	return &JWTResolver{claimsKey: claimsKey}
}

// Resolve mengambil tenant ID dari JWT claims di context.
func (r *JWTResolver) Resolve(req *http.Request) (uuid.UUID, error) {
	v := req.Context().Value(r.claimsKey)
	if v == nil {
		return uuid.Nil, fmt.Errorf("JWT claims tidak ditemukan — RequireAuth belum dipasang")
	}

	cr, ok := v.(claimsReader)
	if !ok {
		return uuid.Nil, fmt.Errorf("JWT claims tidak mengimplementasi claimsReader")
	}

	tenantID := cr.GetTenantID()
	if tenantID == "" {
		return uuid.Nil, fmt.Errorf("tenant_id kosong di JWT claims")
	}

	id, err := uuid.Parse(tenantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("tenant_id di JWT bukan UUID yang valid: %w", err)
	}

	return id, nil
}

// ─── Middlewares ───────────────────────────────────────────────────────────────

// ResolveTenant adalah middleware yang mencoba mengambil tenant ID dari request
// dan menyimpannya ke context. Bersifat fail-open: jika resolver gagal,
// request tetap dilanjutkan tanpa tenant di context.
// Gunakan RequireTenant setelahnya untuk memaksa keberadaan tenant.
func ResolveTenant(resolver Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := resolver.Resolve(r)
			if err == nil && id != uuid.Nil {
				r = r.WithContext(WithTenantID(r.Context(), id))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireTenant adalah middleware yang menolak request jika tidak ada tenant ID di context.
// Harus dipasang setelah ResolveTenant dan RequireAuth.
func RequireTenant() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := TenantIDFromContext(r.Context()); !ok {
				http.Error(w, `{"error":"tenant tidak teridentifikasi"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
