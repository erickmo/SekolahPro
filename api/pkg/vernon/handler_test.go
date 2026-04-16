package vernon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// ──────────────────────────────────────────────────────────────────────────────
// Test fixtures and helpers
// ──────────────────────────────────────────────────────────────────────────────

// fixedTime is a deterministic timestamp used across all tests.
var fixedTime = time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)

// validTenantID and validCompanyID are consistent UUIDs used in test scope.
var (
	validTenantID  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	validCompanyID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

// validScope returns a Scope with valid tenant and company IDs.
func validScope() scope.Scope {
	return scope.Scope{
		TenantID:  validTenantID,
		CompanyID: validCompanyID,
	}
}

// makeSampleDomain creates a BaseDomain with predictable values for assertions.
func makeSampleDomain() *BaseDomain {
	return &BaseDomain{
		ID:          uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
		TenantID:    validTenantID,
		CompanyID:   validCompanyID,
		Rels:        map[string]RelDef{},
		Data:        map[string]any{"name": "Test Item", "status": "active"},
		SyncStatus:  SyncStatusSynced,
		SyncVersion: 1,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
		DeletedAt:   nil,
	}
}

// stubDescriptor is a minimal DomainDescriptor for testing.
type stubDescriptor struct {
	tableName string
	rels      map[string]RelDef
}

func (d *stubDescriptor) TableName() string                { return d.tableName }
func (d *stubDescriptor) DefaultRels() map[string]RelDef  { return d.rels }
func (d *stubDescriptor) Validate(_ map[string]any) error { return nil }

// mockService captures handler calls and returns predetermined results.
type mockService struct {
	listFn    func(ctx context.Context, s scope.Scope, p FindParams) ([]BaseDomain, int64, error)
	getByIDFn func(ctx context.Context, s scope.Scope, id uuid.UUID) (*BaseDomain, error)
	createFn  func(ctx context.Context, s scope.Scope, data map[string]any) (*BaseDomain, error)
	updateFn  func(ctx context.Context, s scope.Scope, id uuid.UUID, data map[string]any) (*BaseDomain, error)
	patchFn   func(ctx context.Context, s scope.Scope, id uuid.UUID, partial map[string]any) (*BaseDomain, error)
	deleteFn  func(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// The mockService does not embed BaseService; instead we build a test-only
// handler that delegates to these functions directly, bypassing the real
// service layer entirely. This avoids the need for a database.

// testHandler replicates BaseHandler routing logic but uses mockService.
// It is intentionally minimal — it mirrors the real handler's error handling,
// response formatting, and middleware behavior so tests exercise the actual
// production code paths for parseUUID, scopeOrError, respondJSON, etc.
type testHandler struct {
	svc *mockService
	log zerolog.Logger
}

func newTestHandler(svc *mockService) *testHandler {
	return &testHandler{svc: svc, log: zerolog.Nop()}
}

// RegisterRoutes mirrors BaseHandler.RegisterRoutes exactly.
func (h *testHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.getByID)
	r.Put("/{id}", h.update)
	r.Patch("/{id}", h.patch)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/_sync", h.forceSync)
}

// list mirrors BaseHandler.list — calls mock listFn.
func (h *testHandler) list(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	params := ParseQueryParams(r, maxPageSize)
	items, total, err := h.svc.listFn(r.Context(), s, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	data := make([]map[string]any, len(items))
	for i := range items {
		data[i] = FormatItem(&items[i])
	}
	respondList(w, data, total, params)
}

// create mirrors BaseHandler.create.
func (h *testHandler) create(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "request body bukan JSON valid")
		return
	}
	entity, err := h.svc.createFn(r.Context(), s, body)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"data": FormatItem(entity)})
}

// getByID mirrors BaseHandler.getByID.
func (h *testHandler) getByID(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	entity, err := h.svc.getByIDFn(r.Context(), s, id)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data": FormatItem(entity),
		"meta": map[string]any{"_sync_status": entity.SyncStatus, "_sync_version": entity.SyncVersion},
	})
}

// update mirrors BaseHandler.update.
func (h *testHandler) update(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "request body bukan JSON valid")
		return
	}
	entity, err := h.svc.updateFn(r.Context(), s, id, body)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": FormatItem(entity)})
}

// patch mirrors BaseHandler.patch.
func (h *testHandler) patch(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "request body bukan JSON valid")
		return
	}
	entity, err := h.svc.patchFn(r.Context(), s, id, body)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": FormatItem(entity)})
}

// delete mirrors BaseHandler.delete.
func (h *testHandler) delete(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.deleteFn(r.Context(), s, id); err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// forceSync mirrors BaseHandler.forceSync (simplified — no real sync).
func (h *testHandler) forceSync(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	entity, err := h.svc.getByIDFn(r.Context(), s, id)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]any{
		"message": "sync dijadwalkan",
		"meta":    map[string]any{"_sync_status": entity.SyncStatus, "_sync_version": entity.SyncVersion},
	})
}

// setupRouter creates a chi.Mux with the testHandler routes registered.
func setupRouter(th *testHandler) *chi.Mux {
	r := chi.NewRouter()
	th.RegisterRoutes(r)
	return r
}

// doRequest executes an HTTP request against the router and returns the response.
func doRequest(router *chi.Mux, method, path string, body string, withScope bool) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if withScope {
		req = req.WithContext(scope.WithScope(req.Context(), validScope()))
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// decodeResponse is a test helper that decodes JSON response body into a map.
func decodeResponse(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to decode response JSON: %v\nbody: %s", err, string(body))
	}
	return result
}

// ──────────────────────────────────────────────────────────────────────────────
// 1. Query Parser Tests (pure functions, no DB)
// ──────────────────────────────────────────────────────────────────────────────

func TestParseQueryParams(t *testing.T) {
	tests := []struct {
		name       string
		rawQuery   string
		maxLimit   int
		wantPage   int
		wantLimit  int
		wantSort   string
		wantFilter map[string]string
	}{
		{
			name:       "default values when no query params",
			rawQuery:   "",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "custom _page and _limit",
			rawQuery:   "_page=3&_limit=10",
			maxLimit:   100,
			wantPage:   3,
			wantLimit:  10,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "max limit enforcement caps to maxLimit",
			rawQuery:   "_limit=500",
			maxLimit:   50,
			wantPage:   1,
			wantLimit:  50,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "max limit equal to requested is allowed",
			rawQuery:   "_limit=100",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  100,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "negative page ignored falls back to default 1",
			rawQuery:   "_page=-1",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "zero page ignored falls back to default 1",
			rawQuery:   "_page=0",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "negative limit ignored falls back to default",
			rawQuery:   "_limit=-5",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "zero limit ignored falls back to default",
			rawQuery:   "_limit=0",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "custom _sort value",
			rawQuery:   "_sort=-created_at",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "-created_at",
			wantFilter: map[string]string{},
		},
		{
			name:       "filters extracted as non-system params",
			rawQuery:   "status=active&priority=high",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{"status": "active", "priority": "high"},
		},
		{
			name:       "dot notation filter preserved as-is",
			rawQuery:   "customer.name=Maju+Jaya",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{"customer.name": "Maju Jaya"},
		},
		{
			name:       "non-numeric page ignored",
			rawQuery:   "_page=abc",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "non-numeric limit ignored",
			rawQuery:   "_limit=xyz",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  defaultLimit,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
		{
			name:       "mixed system and filter params together",
			rawQuery:   "_page=2&_limit=5&_sort=name&status=active&category=general",
			maxLimit:   100,
			wantPage:   2,
			wantLimit:  5,
			wantSort:   "name",
			wantFilter: map[string]string{"status": "active", "category": "general"},
		},
		{
			name:       "limit of 1 is allowed as minimum valid",
			rawQuery:   "_limit=1",
			maxLimit:   100,
			wantPage:   1,
			wantLimit:  1,
			wantSort:   "",
			wantFilter: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.rawQuery, nil)
			got := ParseQueryParams(req, tt.maxLimit)

			if got.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", got.Page, tt.wantPage)
			}
			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", got.Limit, tt.wantLimit)
			}
			if got.Sort != tt.wantSort {
				t.Errorf("Sort = %q, want %q", got.Sort, tt.wantSort)
			}
			if len(got.Filters) != len(tt.wantFilter) {
				t.Errorf("Filters count = %d, want %d", len(got.Filters), len(tt.wantFilter))
			}
			for k, wantV := range tt.wantFilter {
				gotV, ok := got.Filters[k]
				if !ok {
					t.Errorf("missing filter key %q", k)
				} else if gotV != wantV {
					t.Errorf("Filter[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// BuildWhereClause Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestBuildWhereClause(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		companyID     string
		filters       map[string]string
		wantContains  []string // substrings that must appear in WHERE
		wantArgCount  int      // expected number of args
	}{
		{
			name:         "without filters returns base conditions only",
			tenantID:     validTenantID.String(),
			companyID:    validCompanyID.String(),
			filters:      nil,
			wantContains: []string{"deleted_at IS NULL", "tenant_id = $1", "company_id = $2"},
			wantArgCount: 2,
		},
		{
			name:         "with one filter adds one condition and arg",
			tenantID:     validTenantID.String(),
			companyID:    validCompanyID.String(),
			filters:      map[string]string{"status": "active"},
			wantContains: []string{"ILIKE $3", "deleted_at IS NULL", "tenant_id = $1"},
			wantArgCount: 3,
		},
		{
			name:      "with multiple filters adds multiple conditions",
			tenantID:  validTenantID.String(),
			companyID: validCompanyID.String(),
			filters:   map[string]string{"status": "active", "category": "premium"},
			wantContains: []string{
				"ILIKE $3",
				"ILIKE $4",
				"deleted_at IS NULL",
			},
			wantArgCount: 4,
		},
		{
			name:      "dot notation filter produces jsonb path",
			tenantID:  validTenantID.String(),
			companyID: validCompanyID.String(),
			filters:   map[string]string{"customer.name": "Maju"},
			wantContains: []string{
				"_data->'customer'->>'name'",
				"ILIKE $3",
			},
			wantArgCount: 3,
		},
		{
			name:         "empty filter map same as nil",
			tenantID:     validTenantID.String(),
			companyID:    validCompanyID.String(),
			filters:      map[string]string{},
			wantContains: []string{"deleted_at IS NULL", "tenant_id = $1", "company_id = $2"},
			wantArgCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			where, args := BuildWhereClause(tt.tenantID, tt.companyID, tt.filters)

			for _, substr := range tt.wantContains {
				if !strings.Contains(where, substr) {
					t.Errorf("WHERE clause %q missing expected substring %q", where, substr)
				}
			}
			if len(args) != tt.wantArgCount {
				t.Errorf("args count = %d, want %d (WHERE: %s)", len(args), tt.wantArgCount, where)
			}

			// Verify scope args are always the first two.
			if len(args) >= 2 {
				if args[0] != tt.tenantID {
					t.Errorf("args[0] = %v, want tenantID %s", args[0], tt.tenantID)
				}
				if args[1] != tt.companyID {
					t.Errorf("args[1] = %v, want companyID %s", args[1], tt.companyID)
				}
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// BuildOrderClause Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestBuildOrderClause(t *testing.T) {
	tests := []struct {
		name string
		sort string
		want string
	}{
		{
			name: "empty sort defaults to created_at DESC",
			sort: "",
			want: "created_at DESC",
		},
		{
			name: "DESC prefix with dash",
			sort: "-created_at",
			want: "created_at DESC",
		},
		{
			name: "ASC without prefix on JSONB field",
			sort: "name",
			want: "_data->>'name' ASC",
		},
		{
			name: "ASC on top-level field id",
			sort: "id",
			want: "id ASC",
		},
		{
			name: "ASC on top-level field created_at",
			sort: "created_at",
			want: "created_at ASC",
		},
		{
			name: "ASC on top-level field updated_at",
			sort: "updated_at",
			want: "updated_at ASC",
		},
		{
			name: "ASC on top-level field _sync_status",
			sort: "_sync_status",
			want: "_sync_status ASC",
		},
		{
			name: "ASC on top-level field _sync_version",
			sort: "_sync_version",
			want: "_sync_version ASC",
		},
		{
			name: "DESC on top-level field id",
			sort: "-id",
			want: "id DESC",
		},
		{
			name: "DESC on JSONB field",
			sort: "-priority",
			want: "_data->>'priority' DESC",
		},
		{
			name: "SQL injection with special chars falls back to default",
			sort: "name; DROP TABLE users;--",
			want: "created_at DESC",
		},
		{
			name: "SQL injection with single quote falls back",
			sort: "name' OR '1'='1",
			want: "created_at DESC",
		},
		{
			name: "SQL injection with spaces falls back",
			sort: "name ASC",
			want: "created_at DESC",
		},
		{
			name: "underscore field is valid JSONB",
			sort: "customer_name",
			want: "_data->>'customer_name' ASC",
		},
		{
			name: "numeric in field name is valid",
			sort: "field1",
			want: "_data->>'field1' ASC",
		},
		{
			name: "top-level field tenant_id",
			sort: "tenant_id",
			want: "tenant_id ASC",
		},
		{
			name: "top-level field company_id",
			sort: "company_id",
			want: "company_id ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildOrderClause(tt.sort)
			if got != tt.want {
				t.Errorf("BuildOrderClause(%q) = %q, want %q", tt.sort, got, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// toJSONBPath Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestToJSONBPath(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "simple field uses text operator",
			field: "status",
			want:  "_data->>'status'",
		},
		{
			name:  "dot notation one level deep",
			field: "customer.name",
			want:  "_data->'customer'->>'name'",
		},
		{
			name:  "dot notation two levels deep",
			field: "order.customer.name",
			want:  "_data->'order'->'customer'->>'name'",
		},
		{
			name:  "dot notation three levels deep",
			field: "a.b.c.d",
			want:  "_data->'a'->'b'->'c'->>'d'",
		},
		{
			name:  "underscore in field name",
			field: "customer_name",
			want:  "_data->>'customer_name'",
		},
		{
			name:  "invalid chars with semicolon fallback to safe default",
			field: "name; DROP TABLE",
			want:  "_data->>'id'",
		},
		{
			name:  "single quote injection fallback",
			field: "name'",
			want:  "_data->>'id'",
		},
		{
			name:  "space in field fallback",
			field: "field name",
			want:  "_data->>'id'",
		},
		{
			name:  "hyphen in field fallback",
			field: "field-name",
			want:  "_data->>'id'",
		},
		{
			name:  "empty string fallback",
			field: "",
			want:  "_data->>'id'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toJSONBPath(tt.field)
			if got != tt.want {
				t.Errorf("toJSONBPath(%q) = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// isTopLevelField Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestIsTopLevelField(t *testing.T) {
	tests := []struct {
		field string
		want  bool
	}{
		{"id", true},
		{"tenant_id", true},
		{"company_id", true},
		{"created_at", true},
		{"updated_at", true},
		{"_sync_status", true},
		{"_sync_version", true},
		{"name", false},
		{"status", false},
		{"deleted_at", false},
		{"data", false},
		{"", false},
		{"rels", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := isTopLevelField(tt.field)
			if got != tt.want {
				t.Errorf("isTopLevelField(%q) = %v, want %v", tt.field, got, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// 2. Handler Tests with Mock Service (no real DB)
// ──────────────────────────────────────────────────────────────────────────────

// defaultMockService returns a mockService with happy-path stubs.
func defaultMockService() *mockService {
	entity := makeSampleDomain()
	return &mockService{
		listFn: func(_ context.Context, _ scope.Scope, _ FindParams) ([]BaseDomain, int64, error) {
			return []BaseDomain{*entity}, 1, nil
		},
		getByIDFn: func(_ context.Context, _ scope.Scope, _ uuid.UUID) (*BaseDomain, error) {
			return entity, nil
		},
		createFn: func(_ context.Context, _ scope.Scope, _ map[string]any) (*BaseDomain, error) {
			return entity, nil
		},
		updateFn: func(_ context.Context, _ scope.Scope, _ uuid.UUID, _ map[string]any) (*BaseDomain, error) {
			return entity, nil
		},
		patchFn: func(_ context.Context, _ scope.Scope, _ uuid.UUID, _ map[string]any) (*BaseDomain, error) {
			return entity, nil
		},
		deleteFn: func(_ context.Context, _ scope.Scope, _ uuid.UUID) error {
			return nil
		},
	}
}

func TestHandler_AllEndpoints_ReturnForbidden_WhenScopeMissing(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	validID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list without scope", http.MethodGet, "/", ""},
		{"create without scope", http.MethodPost, "/", `{"name":"test"}`},
		{"getByID without scope", http.MethodGet, "/" + validID, ""},
		{"update without scope", http.MethodPut, "/" + validID, `{"name":"test"}`},
		{"patch without scope", http.MethodPatch, "/" + validID, `{"name":"test"}`},
		{"delete without scope", http.MethodDelete, "/" + validID, ""},
		{"forceSync without scope", http.MethodPost, "/"+validID+"/_sync", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doRequest(router, tt.method, tt.path, tt.body, false)

			if w.Code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
			}

			resp := decodeResponse(t, w.Body.Bytes())
			errObj, ok := resp["error"].(map[string]any)
			if !ok {
				t.Fatal("response missing 'error' object")
			}
			if code, _ := errObj["code"].(string); code != "FORBIDDEN" {
				t.Errorf("error.code = %q, want %q", code, "FORBIDDEN")
			}
		})
	}
}

func TestHandler_Create_ReturnsBadRequest_WhenInvalidJSON(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodPost, "/", "not-json{{{", true)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "INVALID_JSON" {
		t.Errorf("error.code = %q, want %q", code, "INVALID_JSON")
	}
}

func TestHandler_Update_ReturnsBadRequest_WhenInvalidJSON(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodPut, "/"+id, "broken json!!!", true)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "INVALID_JSON" {
		t.Errorf("error.code = %q, want %q", code, "INVALID_JSON")
	}
}

func TestHandler_Patch_ReturnsBadRequest_WhenInvalidJSON(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodPatch, "/"+id, "not-json", true)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "INVALID_JSON" {
		t.Errorf("error.code = %q, want %q", code, "INVALID_JSON")
	}
}

func TestHandler_GetByID_ReturnsBadRequest_WhenInvalidUUID(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodGet, "/not-a-uuid", "", true)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "INVALID_ID" {
		t.Errorf("error.code = %q, want %q", code, "INVALID_ID")
	}
}

func TestHandler_Delete_ReturnsBadRequest_WhenInvalidUUID(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodDelete, "/12345", "", true)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "INVALID_ID" {
		t.Errorf("error.code = %q, want %q", code, "INVALID_ID")
	}
}

func TestHandler_ForceSync_ReturnsBadRequest_WhenInvalidUUID(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodPost, "/bad-id/_sync", "", true)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "INVALID_ID" {
		t.Errorf("error.code = %q, want %q", code, "INVALID_ID")
	}
}

func TestHandler_GetByID_ReturnsNotFound_WhenEntityMissing(t *testing.T) {
	svc := defaultMockService()
	svc.getByIDFn = func(_ context.Context, _ scope.Scope, _ uuid.UUID) (*BaseDomain, error) {
		return nil, ErrNotFound
	}
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodGet, "/"+id, "", true)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "NOT_FOUND" {
		t.Errorf("error.code = %q, want %q", code, "NOT_FOUND")
	}
}

func TestHandler_Delete_ReturnsNotFound_WhenEntityMissing(t *testing.T) {
	svc := defaultMockService()
	svc.deleteFn = func(_ context.Context, _ scope.Scope, _ uuid.UUID) error {
		return ErrNotFound
	}
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodDelete, "/"+id, "", true)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_Update_ReturnsNotFound_WhenEntityMissing(t *testing.T) {
	svc := defaultMockService()
	svc.updateFn = func(_ context.Context, _ scope.Scope, _ uuid.UUID, _ map[string]any) (*BaseDomain, error) {
		return nil, ErrNotFound
	}
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodPut, "/"+id, `{"name":"x"}`, true)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_Patch_ReturnsNotFound_WhenEntityMissing(t *testing.T) {
	svc := defaultMockService()
	svc.patchFn = func(_ context.Context, _ scope.Scope, _ uuid.UUID, _ map[string]any) (*BaseDomain, error) {
		return nil, ErrNotFound
	}
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodPatch, "/"+id, `{"name":"x"}`, true)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Response Format Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestHandler_List_ResponseFormat(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodGet, "/", "", true)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	resp := decodeResponse(t, w.Body.Bytes())

	// Must have "data" array.
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("response 'data' is not an array")
	}
	if len(data) != 1 {
		t.Fatalf("data length = %d, want 1", len(data))
	}

	// Must have "meta" with pagination.
	meta, ok := resp["meta"].(map[string]any)
	if !ok {
		t.Fatal("response 'meta' missing or not object")
	}
	for _, key := range []string{"total", "page", "per_page", "total_pages"} {
		if _, exists := meta[key]; !exists {
			t.Errorf("meta missing key %q", key)
		}
	}

	// Verify meta values.
	if total, _ := meta["total"].(float64); total != 1 {
		t.Errorf("meta.total = %v, want 1", total)
	}
	if page, _ := meta["page"].(float64); page != 1 {
		t.Errorf("meta.page = %v, want 1", page)
	}
	if perPage, _ := meta["per_page"].(float64); perPage != defaultLimit {
		t.Errorf("meta.per_page = %v, want %d", perPage, defaultLimit)
	}
}

func TestHandler_GetByID_ResponseFormat(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodGet, "/"+id, "", true)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	resp := decodeResponse(t, w.Body.Bytes())

	// Must have "data" object.
	if _, ok := resp["data"]; !ok {
		t.Fatal("response missing 'data' field")
	}

	// Must have "meta" with sync info.
	meta, ok := resp["meta"].(map[string]any)
	if !ok {
		t.Fatal("response 'meta' missing or not object")
	}
	if _, exists := meta["_sync_status"]; !exists {
		t.Error("meta missing '_sync_status'")
	}
	if _, exists := meta["_sync_version"]; !exists {
		t.Error("meta missing '_sync_version'")
	}
}

func TestHandler_Create_ResponseFormat(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodPost, "/", `{"name":"Test"}`, true)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	resp := decodeResponse(t, w.Body.Bytes())

	// Must have "data" but NOT "meta" (create endpoint does not return meta).
	if _, ok := resp["data"]; !ok {
		t.Fatal("response missing 'data' field")
	}
	if _, ok := resp["meta"]; ok {
		t.Error("response should NOT have 'meta' for create endpoint")
	}
}

func TestHandler_Delete_ReturnsNoContent_OnSuccess(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodDelete, "/"+id, "", true)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// 204 should have empty body.
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body for 204, got %d bytes", w.Body.Len())
	}
}

func TestHandler_ForceSync_ResponseFormat(t *testing.T) {
	svc := defaultMockService()
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodPost, "/"+id+"/_sync", "", true)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusAccepted)
	}

	resp := decodeResponse(t, w.Body.Bytes())

	// Must have "message" field.
	if msg, _ := resp["message"].(string); msg != "sync dijadwalkan" {
		t.Errorf("message = %q, want %q", msg, "sync dijadwalkan")
	}

	// Must have "meta" with sync info.
	if meta, ok := resp["meta"].(map[string]any); !ok {
		t.Fatal("response 'meta' missing")
	} else {
		if _, exists := meta["_sync_status"]; !exists {
			t.Error("meta missing '_sync_status'")
		}
		if _, exists := meta["_sync_version"]; !exists {
			t.Error("meta missing '_sync_version'")
		}
	}
}

func TestHandler_Create_ReturnsUnprocessableEntity_WhenValidationFails(t *testing.T) {
	svc := defaultMockService()
	svc.createFn = func(_ context.Context, _ scope.Scope, _ map[string]any) (*BaseDomain, error) {
		return nil, fmt.Errorf("name is required")
	}
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodPost, "/", `{"no_name":true}`, true)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "VALIDATION_ERROR" {
		t.Errorf("error.code = %q, want %q", code, "VALIDATION_ERROR")
	}
}

func TestHandler_List_ReturnsInternalServerError_WhenServiceFails(t *testing.T) {
	svc := defaultMockService()
	svc.listFn = func(_ context.Context, _ scope.Scope, _ FindParams) ([]BaseDomain, int64, error) {
		return nil, 0, fmt.Errorf("database connection lost")
	}
	th := newTestHandler(svc)
	router := setupRouter(th)

	w := doRequest(router, http.MethodGet, "/", "", true)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	resp := decodeResponse(t, w.Body.Bytes())
	errObj, _ := resp["error"].(map[string]any)
	if code, _ := errObj["code"].(string); code != "INTERNAL_ERROR" {
		t.Errorf("error.code = %q, want %q", code, "INTERNAL_ERROR")
	}
}

func TestHandler_ForceSync_ReturnsNotFound_WhenEntityMissing(t *testing.T) {
	svc := defaultMockService()
	svc.getByIDFn = func(_ context.Context, _ scope.Scope, _ uuid.UUID) (*BaseDomain, error) {
		return nil, ErrNotFound
	}
	th := newTestHandler(svc)
	router := setupRouter(th)

	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := doRequest(router, http.MethodPost, "/"+id+"/_sync", "", true)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// 3. Helper Function Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestFormatItem(t *testing.T) {
	t.Run("merges _data with top-level fields", func(t *testing.T) {
		entity := makeSampleDomain()
		result := FormatItem(entity)

		// Data field from _data.
		if name, _ := result["name"].(string); name != "Test Item" {
			t.Errorf("name = %q, want %q", name, "Test Item")
		}
		if status, _ := result["status"].(string); status != "active" {
			t.Errorf("status = %q, want %q", status, "active")
		}

		// Top-level fields.
		if _, ok := result["id"]; !ok {
			t.Error("missing 'id' field")
		}
		if _, ok := result["created_at"]; !ok {
			t.Error("missing 'created_at' field")
		}
		if _, ok := result["updated_at"]; !ok {
			t.Error("missing 'updated_at' field")
		}
		if _, ok := result["_sync_status"]; !ok {
			t.Error("missing '_sync_status' field")
		}
		if _, ok := result["_sync_version"]; !ok {
			t.Error("missing '_sync_version' field")
		}
	})

	t.Run("deleted_at omitted when nil", func(t *testing.T) {
		entity := makeSampleDomain()
		entity.DeletedAt = nil
		result := FormatItem(entity)

		if _, ok := result["deleted_at"]; ok {
			t.Error("deleted_at should be omitted when nil")
		}
	})

	t.Run("deleted_at present when set", func(t *testing.T) {
		entity := makeSampleDomain()
		deletedAt := fixedTime
		entity.DeletedAt = &deletedAt
		result := FormatItem(entity)

		if _, ok := result["deleted_at"]; !ok {
			t.Error("deleted_at should be present when set")
		}
	})

	t.Run("data fields override top-level when same key", func(t *testing.T) {
		entity := makeSampleDomain()
		// Put a value in _data that has the same key as a top-level field.
		entity.Data["id"] = "should-be-overridden"
		result := FormatItem(entity)

		// The top-level id should win because maps.Copy of Data happens first,
		// then top-level fields are set explicitly.
		idVal := result["id"]
		if idVal == "should-be-overridden" {
			t.Error("data field 'id' should be overridden by top-level field")
		}
	})
}

func TestExtractFields(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		fields []string
		want   map[string]any
	}{
		{
			name:   "empty fields list returns full data",
			data:   map[string]any{"a": 1, "b": 2},
			fields: nil,
			want:   map[string]any{"a": 1, "b": 2},
		},
		{
			name:   "empty fields slice returns full data",
			data:   map[string]any{"a": 1, "b": 2},
			fields: []string{},
			want:   map[string]any{"a": 1, "b": 2},
		},
		{
			name:   "subset of fields filters correctly",
			data:   map[string]any{"name": "Alice", "age": 30, "city": "Jakarta"},
			fields: []string{"name", "city"},
			want:   map[string]any{"name": "Alice", "city": "Jakarta"},
		},
		{
			name:   "missing fields are skipped silently",
			data:   map[string]any{"name": "Alice"},
			fields: []string{"name", "missing_key"},
			want:   map[string]any{"name": "Alice"},
		},
		{
			name:   "all fields missing returns empty map",
			data:   map[string]any{"name": "Alice"},
			fields: []string{"x", "y", "z"},
			want:   map[string]any{},
		},
		{
			name:   "exact match returns all",
			data:   map[string]any{"a": 1, "b": "two"},
			fields: []string{"a", "b"},
			want:   map[string]any{"a": 1, "b": "two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFields(tt.data, tt.fields)
			if len(got) != len(tt.want) {
				t.Errorf("result length = %d, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q", k)
				} else if gotV != wantV {
					t.Errorf("result[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// 4. Domain Type Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestBaseDomain_SetDataField_InitializesNilMap(t *testing.T) {
	b := &BaseDomain{}
	if b.Data != nil {
		t.Fatal("Data should start as nil")
	}
	b.SetDataField("key", "value")
	if b.Data == nil {
		t.Fatal("Data should be initialized after SetDataField")
	}
	if v, ok := b.Data["key"]; !ok || v != "value" {
		t.Errorf("Data[key] = %v, want 'value'", v)
	}
}

func TestBaseDomain_SetDataField_Overwrites(t *testing.T) {
	b := &BaseDomain{Data: map[string]any{"key": "old"}}
	b.SetDataField("key", "new")
	if b.Data["key"] != "new" {
		t.Errorf("Data[key] = %v, want 'new'", b.Data["key"])
	}
}

func TestBaseDomain_GetDataField(t *testing.T) {
	b := &BaseDomain{Data: map[string]any{"name": "Alice"}}

	t.Run("existing key returns value and true", func(t *testing.T) {
		v, ok := b.GetDataField("name")
		if !ok {
			t.Error("expected ok=true for existing key")
		}
		if v != "Alice" {
			t.Errorf("value = %v, want 'Alice'", v)
		}
	})

	t.Run("missing key returns nil and false", func(t *testing.T) {
		v, ok := b.GetDataField("nonexistent")
		if ok {
			t.Error("expected ok=false for missing key")
		}
		if v != nil {
			t.Errorf("value = %v, want nil", v)
		}
	})

	t.Run("nil data map returns nil and false", func(t *testing.T) {
		b := &BaseDomain{}
		v, ok := b.GetDataField("anything")
		if ok {
			t.Error("expected ok=false for nil data map")
		}
		if v != nil {
			t.Errorf("value = %v, want nil", v)
		}
	})
}

func TestSyncEvent_EventName(t *testing.T) {
	tests := []struct {
		domain string
		want   string
	}{
		{"customers", "sync.customers"},
		{"orders", "sync.orders"},
		{"", "sync."},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			e := SyncEvent{Domain: tt.domain}
			if got := e.EventName(); got != tt.want {
				t.Errorf("EventName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSyncEvent_AggregateID(t *testing.T) {
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	e := SyncEvent{EntityID: id}
	if got := e.AggregateID(); got != id {
		t.Errorf("AggregateID() = %q, want %q", got, id)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Error Sentinel Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestErrorSentinelValues(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		want  string
	}{
		{"ErrNotFound", ErrNotFound, "entity not found"},
		{"ErrVersionConflict", ErrVersionConflict, "sync version conflict"},
		{"ErrCycleDetected", ErrCycleDetected, "circular autoload dependency detected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error sentinel must not be nil")
			}
			if !strings.Contains(tt.err.Error(), tt.want) {
				t.Errorf("error message %q does not contain %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrorSentinels_AreDistinct(t *testing.T) {
	errors := []error{ErrNotFound, ErrVersionConflict, ErrCycleDetected}
	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[i] == errors[j] {
				t.Errorf("sentinel errors[%d] and errors[%d] are the same", i, j)
			}
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Action and Sync Status Constants Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestActionConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"ActionCreate", ActionCreate, "create"},
		{"ActionUpdate", ActionUpdate, "update"},
		{"ActionPatch", ActionPatch, "patch"},
		{"ActionDelete", ActionDelete, "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("constant = %q, want %q", tt.value, tt.want)
			}
		})
	}
}

func TestSyncStatusConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"SyncStatusSynced", SyncStatusSynced, "synced"},
		{"SyncStatusPending", SyncStatusPending, "pending"},
		{"SyncStatusError", SyncStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("constant = %q, want %q", tt.value, tt.want)
			}
		})
	}
}

func TestRelTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"RelBelongsTo", RelBelongsTo, "belongs_to"},
		{"RelHasOne", RelHasOne, "has_one"},
		{"RelHasMany", RelHasMany, "has_many"},
		{"RelManyToMany", RelManyToMany, "many_to_many"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("constant = %q, want %q", tt.value, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// paginationCalc Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestPaginationCalc(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "valid page 2 limit 10",
			page:       2,
			limit:      10,
			wantLimit:  10,
			wantOffset: 10,
		},
		{
			name:       "page 1 limit 20",
			page:       1,
			limit:      20,
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "zero limit falls back to default",
			page:       1,
			limit:      0,
			wantLimit:  defaultLimit,
			wantOffset: 0,
		},
		{
			name:       "negative limit falls back to default",
			page:       1,
			limit:      -5,
			wantLimit:  defaultLimit,
			wantOffset: 0,
		},
		{
			name:       "zero page falls back to page 1",
			page:       0,
			limit:      10,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "negative page falls back to page 1",
			page:       -1,
			limit:      10,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "page 3 limit 25",
			page:       3,
			limit:      25,
			wantLimit:  25,
			wantOffset: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := FindParams{Page: tt.page, Limit: tt.limit}
			limit, offset := paginationCalc(p)
			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// copyMap and keysOf Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestCopyMap(t *testing.T) {
	original := map[string]any{"a": 1, "b": "two", "c": true}
	copied := copyMap(original)

	if len(copied) != len(original) {
		t.Fatalf("copied length = %d, want %d", len(copied), len(original))
	}

	for k, v := range original {
		if copied[k] != v {
			t.Errorf("copied[%q] = %v, want %v", k, copied[k], v)
		}
	}

	// Verify it is a shallow copy (modifying copy does not affect original).
	copied["a"] = 999
	if original["a"] != 1 {
		t.Error("modifying copy affected original — not a shallow copy")
	}
}

func TestCopyMap_Empty(t *testing.T) {
	result := copyMap(map[string]any{})
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d items", len(result))
	}
}

func TestCopyMap_Nil(t *testing.T) {
	result := copyMap(nil)
	if len(result) != 0 {
		t.Errorf("expected empty map for nil input, got %d items", len(result))
	}
}

func TestKeysOf(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]any
		wantCount int
		wantKeys  []string
	}{
		{
			name:      "three keys",
			input:     map[string]any{"a": 1, "b": 2, "c": 3},
			wantCount: 3,
			wantKeys:  []string{"a", "b", "c"},
		},
		{
			name:      "empty map",
			input:     map[string]any{},
			wantCount: 0,
			wantKeys:  nil,
		},
		{
			name:      "nil map",
			input:     nil,
			wantCount: 0,
			wantKeys:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keysOf(tt.input)
			if len(got) != tt.wantCount {
				t.Errorf("keys count = %d, want %d", len(got), tt.wantCount)
			}
			// Check all expected keys are present (order is not guaranteed).
			gotSet := make(map[string]bool, len(got))
			for _, k := range got {
				gotSet[k] = true
			}
			for _, k := range tt.wantKeys {
				if !gotSet[k] {
					t.Errorf("missing key %q in result", k)
				}
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// respondJSON / respondError Response Format Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestRespondError_Format(t *testing.T) {
	w := httptest.NewRecorder()
	respondError(w, http.StatusBadRequest, "TEST_CODE", "test message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'error' object")
	}
	if errObj["code"] != "TEST_CODE" {
		t.Errorf("error.code = %v, want TEST_CODE", errObj["code"])
	}
	if errObj["message"] != "test message" {
		t.Errorf("error.message = %v, want 'test message'", errObj["message"])
	}
}

func TestRespondList_MetaCalculation(t *testing.T) {
	w := httptest.NewRecorder()
	data := []map[string]any{{"name": "item1"}, {"name": "item2"}}
	params := FindParams{Page: 1, Limit: 10}
	respondList(w, data, 25, params)

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	meta, _ := resp["meta"].(map[string]any)
	if total, _ := meta["total"].(float64); total != 25 {
		t.Errorf("meta.total = %v, want 25", total)
	}
	if totalPages, _ := meta["total_pages"].(float64); totalPages != 3 {
		t.Errorf("meta.total_pages = %v, want 3 (25 items / 10 per page = 3 pages)", totalPages)
	}
	if page, _ := meta["page"].(float64); page != 1 {
		t.Errorf("meta.page = %v, want 1", page)
	}
	if perPage, _ := meta["per_page"].(float64); perPage != 10 {
		t.Errorf("meta.per_page = %v, want 10", perPage)
	}
}

func TestRespondList_MetaCalculation_ExactDivision(t *testing.T) {
	w := httptest.NewRecorder()
	data := []map[string]any{}
	params := FindParams{Page: 2, Limit: 10}
	respondList(w, data, 20, params)

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	meta, _ := resp["meta"].(map[string]any)
	// 20 items / 10 per page = exactly 2 pages (no extra page).
	if totalPages, _ := meta["total_pages"].(float64); totalPages != 2 {
		t.Errorf("meta.total_pages = %v, want 2", totalPages)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// FindParams Struct Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestFindParams_Fields(t *testing.T) {
	filters := map[string]string{"status": "active", "name": "test"}
	p := FindParams{
		Filters: filters,
		Sort:    "-created_at",
		Page:    2,
		Limit:   50,
	}

	if p.Sort != "-created_at" {
		t.Errorf("Sort = %q, want '-created_at'", p.Sort)
	}
	if p.Page != 2 {
		t.Errorf("Page = %d, want 2", p.Page)
	}
	if p.Limit != 50 {
		t.Errorf("Limit = %d, want 50", p.Limit)
	}
	if len(p.Filters) != 2 {
		t.Errorf("Filters count = %d, want 2", len(p.Filters))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// RelDef Struct Test
// ──────────────────────────────────────────────────────────────────────────────

func TestRelDef_Fields(t *testing.T) {
	rd := RelDef{
		Domain:      "customers",
		Type:        RelBelongsTo,
		FK:          "customer_id",
		LocalKey:    "id",
		ForeignKey:  "customer_id",
		IsAutoload:  true,
		Fields:      []string{"name", "email"},
		MaxItems:    0,
		Junction:    "",
	}

	if rd.Domain != "customers" {
		t.Errorf("Domain = %q, want 'customers'", rd.Domain)
	}
	if rd.Type != RelBelongsTo {
		t.Errorf("Type = %q, want %q", rd.Type, RelBelongsTo)
	}
	if rd.FK != "customer_id" {
		t.Errorf("FK = %q, want 'customer_id'", rd.FK)
	}
	if !rd.IsAutoload {
		t.Error("IsAutoload should be true")
	}
	if len(rd.Fields) != 2 {
		t.Errorf("Fields count = %d, want 2", len(rd.Fields))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// stubDescriptor Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestStubDescriptor_TableName(t *testing.T) {
	d := &stubDescriptor{tableName: "orders"}
	if d.TableName() != "orders" {
		t.Errorf("TableName() = %q, want 'orders'", d.TableName())
	}
}

func TestStubDescriptor_DefaultRels(t *testing.T) {
	rels := map[string]RelDef{
		"customer": {Domain: "customers", Type: RelBelongsTo, IsAutoload: true},
	}
	d := &stubDescriptor{rels: rels}
	got := d.DefaultRels()
	if len(got) != 1 {
		t.Errorf("DefaultRels count = %d, want 1", len(got))
	}
	if r, ok := got["customer"]; !ok || r.Domain != "customers" {
		t.Error("missing 'customer' rel or wrong domain")
	}
}

func TestStubDescriptor_Validate(t *testing.T) {
	d := &stubDescriptor{}
	if err := d.Validate(map[string]any{"any": "data"}); err != nil {
		t.Errorf("stub Validate should return nil, got %v", err)
	}
}
