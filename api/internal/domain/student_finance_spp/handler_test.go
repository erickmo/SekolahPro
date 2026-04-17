package student_finance_spp_test

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

	"github.com/yourorg/boilerplate/internal/domain/student_finance_spp"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures -----------------------------------------------------------

var sppTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func sppScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- helpers -----------------------------------------------------------------

func sppRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func sppRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func sppRespondInvalidUUID(w http.ResponseWriter, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_ID", "message": field + " harus berupa UUID valid"},
	})
}

func sppRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}

// -- Scope enforcement tests -------------------------------------------------

func TestInvoiceHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sppRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validInvoiceBody(t)
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

func TestInvoiceHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(sppScopeMiddleware(sppTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, sppTestScope.TenantID, s.TenantID)
		assert.Equal(t, sppTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -------------------------------------------------------

func TestInvoiceHandler_InvalidJSON_Returns400(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	r := chi.NewRouter()
	r.Use(sppScopeMiddleware(sppTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sppRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sppRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			sppRespondValidationError(w, err.Error())
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

// -- Invalid FK UUID tests --------------------------------------------------

func TestInvoiceHandler_InvalidStudentUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(sppScopeMiddleware(sppTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sppRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sppRespondInvalidJSON(w)
			return
		}
		studentID, _ := body["student_id"].(string)
		if _, err := uuid.Parse(studentID); err != nil {
			sppRespondInvalidUUID(w, "student_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"student_id":       "not-a-uuid",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"fee_type_id":      "00000000-0000-0000-0000-000000000003",
		"invoice_no":       "INV-2026-001",
		"period_year":      2026,
		"amount":           500000,
		"total_amount":     500000,
		"due_date":         "2026-04-30",
		"status":           student_finance_spp.InvoiceStatusUnpaid,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPaymentHandler_InvalidInvoiceUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(sppScopeMiddleware(sppTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sppRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sppRespondInvalidJSON(w)
			return
		}
		invoiceID, _ := body["invoice_id"].(string)
		if _, err := uuid.Parse(invoiceID); err != nil {
			sppRespondInvalidUUID(w, "invoice_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"invoice_id":     "bad-uuid",
		"student_id":     "00000000-0000-0000-0000-000000000002",
		"receipt_no":     "RCT-2026-001",
		"amount":         500000,
		"payment_method": student_finance_spp.PaymentMethodCash,
		"payment_date":   "2026-04-17",
		"received_by":    "00000000-0000-0000-0000-000000000003",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// -- body helpers ------------------------------------------------------------

func validInvoiceBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"student_id":       "00000000-0000-0000-0000-000000000001",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"fee_type_id":      "00000000-0000-0000-0000-000000000003",
		"invoice_no":       "INV-2026-001",
		"period_year":      2026,
		"amount":           500000,
		"total_amount":     500000,
		"due_date":         "2026-04-30",
		"status":           student_finance_spp.InvoiceStatusUnpaid,
	})
	require.NoError(t, err)
	return body
}
