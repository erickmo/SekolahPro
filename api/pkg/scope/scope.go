// Package scope menyediakan tipe dan helper untuk hierarki organisasi 4 level:
//
//	Tenant (wajib) → Company (wajib) → Branch (opsional) → Warehouse (opsional)
//
// Aturan layer:
//   - Domain layer (internal/domain/) BOLEH import package ini karena Scope adalah
//     pure value object tanpa framework dependency.
//   - Application layer (internal/command/, internal/query/) membaca Scope dari context
//     lalu meneruskannya sebagai parameter eksplisit ke repository.
//   - Repository menerima Scope sebagai parameter — TIDAK membaca dari context.
package scope

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// Level merepresentasikan level hierarki organisasi yang dibutuhkan.
type Level int

const (
	LevelTenant    Level = iota + 1 // Wajib — top-level namespace isolasi data
	LevelCompany                    // Wajib untuk operasi bisnis — legal entity
	LevelBranch                     // Opsional — unit operasional/cabang
	LevelWarehouse                  // Opsional — lokasi fisik penyimpanan
)

// Scope menyimpan konteks organisasi untuk sebuah request.
// Ini adalah pure value object — tidak ada logic framework, aman diimport oleh domain layer.
type Scope struct {
	TenantID    uuid.UUID  // Selalu ada — L1
	CompanyID   uuid.UUID  // Ada setelah select-scope — L2
	BranchID    *uuid.UUID // Opsional — L3; nil jika tidak relevan
	WarehouseID *uuid.UUID // Opsional — L4; prefer kirim via request body, bukan di token
}

var (
	ErrMissingTenant    = errors.New("tenant tidak teridentifikasi")
	ErrMissingCompany   = errors.New("company tidak teridentifikasi — pilih perusahaan terlebih dahulu")
	ErrMissingBranch    = errors.New("branch tidak teridentifikasi")
	ErrMissingWarehouse = errors.New("warehouse tidak teridentifikasi")
)

// IsValid memeriksa bahwa TenantID dan CompanyID sudah di-set (minimum untuk operasi bisnis).
func (s Scope) IsValid() error {
	if s.TenantID == uuid.Nil {
		return ErrMissingTenant
	}
	if s.CompanyID == uuid.Nil {
		return ErrMissingCompany
	}
	return nil
}

// HasBranch mengembalikan true jika BranchID ter-set.
func (s Scope) HasBranch() bool {
	return s.BranchID != nil && *s.BranchID != uuid.Nil
}

// HasWarehouse mengembalikan true jika WarehouseID ter-set.
func (s Scope) HasWarehouse() bool {
	return s.WarehouseID != nil && *s.WarehouseID != uuid.Nil
}

// ── Context Helpers ────────────────────────────────────────────────────────────

// contextKey adalah tipe private untuk menghindari collision di context.
type contextKey struct{}

// WithScope menyimpan Scope ke dalam context.
func WithScope(ctx context.Context, s Scope) context.Context {
	return context.WithValue(ctx, contextKey{}, s)
}

// ScopeFromContext mengambil Scope dari context.
// Mengembalikan (Scope{}, false) jika tidak ada atau TenantID kosong.
func ScopeFromContext(ctx context.Context) (Scope, bool) {
	s, ok := ctx.Value(contextKey{}).(Scope)
	if !ok || s.TenantID == uuid.Nil {
		return Scope{}, false
	}
	return s, true
}

// MustScopeFromContext mengambil Scope dari context.
// Panic jika tidak ada — gunakan hanya di tempat yang dijamin sudah ada middleware RequireScope.
func MustScopeFromContext(ctx context.Context) Scope {
	s, ok := ScopeFromContext(ctx)
	if !ok {
		panic("scope organisasi missing dari context — ResolveScope middleware belum dipasang")
	}
	return s
}

// ── Resolver Interface ─────────────────────────────────────────────────────────

// Resolver mendefinisikan cara mengambil Scope dari HTTP request.
// Implementasi dipilih berdasarkan TENANT_MODE di config (via Uber FX).
type Resolver interface {
	Resolve(r *http.Request) (Scope, error)
}

// ── ConfigScopeResolver (mode "single") ───────────────────────────────────────

// ConfigScopeResolver menggunakan Scope fixed dari konfigurasi aplikasi.
// Dipakai pada deployment single-tenant di mana semua level sudah diketahui saat startup.
type ConfigScopeResolver struct {
	fixed Scope
}

// NewConfigScopeResolver membuat ConfigScopeResolver dengan Scope yang sudah di-build dari config.
func NewConfigScopeResolver(fixed Scope) *ConfigScopeResolver {
	return &ConfigScopeResolver{fixed: fixed}
}

// Resolve selalu mengembalikan Scope yang sama (dari config).
func (r *ConfigScopeResolver) Resolve(_ *http.Request) (Scope, error) {
	return r.fixed, nil
}

// ── JWTScopeResolver (mode "multi") ───────────────────────────────────────────

// scopeClaimsReader mendefinisikan interface minimal untuk membaca semua level scope
// dari JWT claims. Diimplementasi oleh pkg/jwt.Claims untuk menghindari circular import.
type scopeClaimsReader interface {
	GetTenantID() string
	GetCompanyID() *string
	GetBranchID() *string
	GetWarehouseID() *string
}

// JWTScopeResolver membaca Scope dari JWT claims yang sudah ada di context.
// Harus dipasang SETELAH middleware RequireAuth yang menyimpan claims ke context.
type JWTScopeResolver struct {
	claimsKey any // context key identik dengan yang dipakai RequireAuth middleware
}

// NewJWTScopeResolver membuat JWTScopeResolver.
// claimsKey harus sama persis (tipe dan nilai) dengan key yang dipakai RequireAuth.
func NewJWTScopeResolver(claimsKey any) *JWTScopeResolver {
	return &JWTScopeResolver{claimsKey: claimsKey}
}

// Resolve mengambil Scope dari JWT claims di context.
func (r *JWTScopeResolver) Resolve(req *http.Request) (Scope, error) {
	v := req.Context().Value(r.claimsKey)
	if v == nil {
		return Scope{}, fmt.Errorf("JWT claims tidak ditemukan — RequireAuth belum dipasang")
	}

	cr, ok := v.(scopeClaimsReader)
	if !ok {
		return Scope{}, fmt.Errorf("JWT claims tidak mengimplementasi scopeClaimsReader")
	}

	tidStr := cr.GetTenantID()
	if tidStr == "" {
		return Scope{}, fmt.Errorf("tenant_id kosong di JWT claims")
	}
	tenantID, err := uuid.Parse(tidStr)
	if err != nil {
		return Scope{}, fmt.Errorf("tenant_id di JWT bukan UUID valid: %w", err)
	}

	s := Scope{TenantID: tenantID}

	if cid := cr.GetCompanyID(); cid != nil && *cid != "" {
		id, err := uuid.Parse(*cid)
		if err != nil {
			return Scope{}, fmt.Errorf("company_id di JWT bukan UUID valid: %w", err)
		}
		s.CompanyID = id
	}

	if bid := cr.GetBranchID(); bid != nil && *bid != "" {
		id, err := uuid.Parse(*bid)
		if err != nil {
			return Scope{}, fmt.Errorf("branch_id di JWT bukan UUID valid: %w", err)
		}
		s.BranchID = &id
	}

	if wid := cr.GetWarehouseID(); wid != nil && *wid != "" {
		id, err := uuid.Parse(*wid)
		if err != nil {
			return Scope{}, fmt.Errorf("warehouse_id di JWT bukan UUID valid: %w", err)
		}
		s.WarehouseID = &id
	}

	return s, nil
}

// ── Middlewares ────────────────────────────────────────────────────────────────

// ResolveScope adalah middleware yang mencoba mengambil Scope dari request
// dan menyimpannya ke context. Bersifat fail-open: jika resolver gagal,
// request tetap dilanjutkan tanpa scope di context.
// Selalu pasang RequireScope setelahnya untuk enforce level yang dibutuhkan.
func ResolveScope(resolver Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, err := resolver.Resolve(r)
			if err == nil && s.TenantID != uuid.Nil {
				r = r.WithContext(WithScope(r.Context(), s))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireScope adalah middleware yang menolak request jika level scope yang dibutuhkan
// tidak tersedia di context. Harus dipasang setelah ResolveScope.
//
// Contoh penggunaan:
//
//	r.Use(scope.RequireScope(scope.LevelTenant, scope.LevelCompany))
func RequireScope(levels ...Level) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, ok := ScopeFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"scope organisasi tidak teridentifikasi"}`, http.StatusForbidden)
				return
			}

			for _, lvl := range levels {
				var errMsg string
				switch lvl {
				case LevelTenant:
					if s.TenantID == uuid.Nil {
						errMsg = ErrMissingTenant.Error()
					}
				case LevelCompany:
					if s.CompanyID == uuid.Nil {
						errMsg = ErrMissingCompany.Error()
					}
				case LevelBranch:
					if !s.HasBranch() {
						errMsg = ErrMissingBranch.Error()
					}
				case LevelWarehouse:
					if !s.HasWarehouse() {
						errMsg = ErrMissingWarehouse.Error()
					}
				}
				if errMsg != "" {
					http.Error(w, `{"error":"`+errMsg+`"}`, http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
