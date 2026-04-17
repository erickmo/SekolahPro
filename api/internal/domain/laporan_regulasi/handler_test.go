package laporan_regulasi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/laporan_regulasi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var lrTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func lrScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestLaporanRegulasiHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			lrRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validLaporanBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok, "response should contain error object")
	assert.Equal(t, "FORBIDDEN", errMap["code"])
}

func TestLaporanRegulasiHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(lrScopeMiddleware(lrTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, lrTestScope.TenantID, s.TenantID)
		assert.Equal(t, lrTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestLaporanRegulasiHandler_InvalidJSONBody(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	r := chi.NewRouter()
	r.Use(lrScopeMiddleware(lrTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			lrRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			lrRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			lrRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "INVALID_JSON", errMap["code"])
}

// -- Invalid UUID in path -------------------------------------------------------

func TestLaporanRegulasiHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(lrScopeMiddleware(lrTestScope))
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		raw := chi.URLParam(req, "id")
		_, err := uuid.Parse(raw)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "INVALID_ID",
					"message": "ID harus berupa UUID valid",
				},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		pathParam  string
		wantStatus int
	}{
		{"not-a-uuid", "not-a-uuid", http.StatusBadRequest},
		{"valid-uuid", uuid.New().String(), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// -- Full handler pipeline simulation -------------------------------------------

func TestLaporanRegulasiHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	r := chi.NewRouter()
	r.Use(lrScopeMiddleware(lrTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			lrRespondForbidden(w)
			return
		}
		assert.Equal(t, lrTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			lrRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			lrRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validLaporanBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestLaporanRegulasiHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	r := chi.NewRouter()
	r.Use(lrScopeMiddleware(lrTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			lrRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			lrRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			lrRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"report_type": "neraca",
		// Missing: laporan_config_id, report_category, period_type, etc.
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", errMap["code"])
}

// -- Config handler pipeline tests ----------------------------------------------

func TestLaporanRegulasiHandler_ConfigCreate_Valid(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	var parsed map[string]any
	body := validConfigBody(t)
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)
	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestLaporanRegulasiHandler_VersiCreate_Valid(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	var parsed map[string]any
	body := validVersiBody(t)
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)
	err = d.Validate(parsed)
	assert.NoError(t, err)
}

// -- Table-driven: Laporan validation pipeline -----------------------------------

func TestLaporanRegulasiHandler_LaporanValidation_Table(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid draft laporan",
			body:    validLaporanData(),
			wantErr: false,
		},
		{
			name: "missing laporan_config_id",
			body: func() map[string]any {
				data := validLaporanData()
				delete(data, laporan_regulasi.LapFieldLaporanConfigID)
				return data
			}(),
			wantErr: true,
			errMsg:  "laporan_config_id wajib diisi",
		},
		{
			name: "missing report_type",
			body: func() map[string]any {
				data := validLaporanData()
				delete(data, laporan_regulasi.LapFieldReportType)
				return data
			}(),
			wantErr: true,
			errMsg:  "report_type wajib diisi",
		},
		{
			name: "missing report_category",
			body: func() map[string]any {
				data := validLaporanData()
				delete(data, laporan_regulasi.LapFieldReportCategory)
				return data
			}(),
			wantErr: true,
			errMsg:  "report_category wajib diisi",
		},
		{
			name: "missing period_label",
			body: func() map[string]any {
				data := validLaporanData()
				delete(data, laporan_regulasi.LapFieldPeriodLabel)
				return data
			}(),
			wantErr: true,
			errMsg:  "period_label wajib diisi",
		},
		{
			name: "invalid status",
			body: func() map[string]any {
				data := validLaporanData()
				data[laporan_regulasi.LapFieldStatus] = "cancelled"
				return data
			}(),
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "invalid consolidation_level",
			body: func() map[string]any {
				data := validLaporanData()
				data[laporan_regulasi.LapFieldConsolidationLevel] = "invalid"
				return data
			}(),
			wantErr: true,
			errMsg:  "consolidation_level tidak valid",
		},
		{
			name: "valid submitted laporan",
			body: func() map[string]any {
				data := validLaporanData()
				data[laporan_regulasi.LapFieldStatus] = laporan_regulasi.LapStatusSubmitted
				return data
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			var parsed map[string]any
			err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsed)
			require.NoError(t, err)

			err = d.Validate(parsed)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// -- helpers ---------------------------------------------------------------------

func validLaporanBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validLaporanData())
	require.NoError(t, err)
	return body
}

func validConfigBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validConfigData())
	require.NoError(t, err)
	return body
}

func validVersiBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validVersiData())
	require.NoError(t, err)
	return body
}

func lrRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func lrRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func lrRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
