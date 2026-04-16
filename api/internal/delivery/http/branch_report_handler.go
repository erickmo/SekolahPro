package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createbranchreport "github.com/yourorg/boilerplate/internal/command/create_branch_report"
	deletebranchreport "github.com/yourorg/boilerplate/internal/command/delete_branch_report"
	updatebranchreport "github.com/yourorg/boilerplate/internal/command/update_branch_report"
	approvebranchreport "github.com/yourorg/boilerplate/internal/command/approve_branch_report"
	createeliminationrule "github.com/yourorg/boilerplate/internal/command/create_elimination_rule"
	toggleeliminationrule "github.com/yourorg/boilerplate/internal/command/toggle_elimination_rule"
	getbranchreport "github.com/yourorg/boilerplate/internal/query/get_branch_report_by_id"
	listbranchreports "github.com/yourorg/boilerplate/internal/query/list_branch_reports"
	getconsolidated "github.com/yourorg/boilerplate/internal/query/get_consolidated_report"
	listeliminationrules "github.com/yourorg/boilerplate/internal/query/list_elimination_rules"
	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// BranchReportHandler menangani HTTP request untuk domain BranchReport.
type BranchReportHandler struct {
	BaseHandler
}

var branchReportSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"period_start": "period_start",
		"created_at":   "created_at",
		"period_type":  "period_type",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

var eliminationRuleSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"rule_name":  "rule_name",
		"rule_type":  "rule_type",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewBranchReportHandler membuat instance BranchReportHandler baru.
func NewBranchReportHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *BranchReportHandler {
	return &BranchReportHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain BranchReport.
func (h *BranchReportHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/branch-reports", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/consolidated", h.GetConsolidated)
		r.Get("/elimination-rules", h.ListEliminationRules)
		r.Post("/elimination-rules", h.CreateEliminationRule)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Post("/{id}/approve", h.Approve)
		r.Delete("/{id}", h.Delete)
		r.Put("/elimination-rules/{ruleId}/toggle", h.ToggleEliminationRule)
	})
}

// List godoc
// @Summary      List branch financial reports
// @Tags         branch-reports
// @Produce      json
// @Param        branch_id    query     string  false  "Branch UUID (empty = all)"
// @Param        period_type  query     string  false  "daily|weekly|monthly|quarterly|yearly"
// @Param        start_date   query     string  false  "Period start >= (YYYY-MM-DD)"
// @Param        end_date     query     string  false  "Period end <= (YYYY-MM-DD)"
// @Param        approved_by  query     string  false  "Approved by user UUID"
// @Success      200  {object}  listbranchreports.Result
// @Router       /api/v1/branch-reports [get]
func (h *BranchReportHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, branchReportSortConfig)

	var branchID *uuid.UUID
	if bid := r.URL.Query().Get("branch_id"); bid != "" {
		parsed, err := uuid.Parse(bid)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "branch_id tidak valid")
			return
		}
		branchID = &parsed
	}

	var periodType *branch_report.PeriodType
	if pt := r.URL.Query().Get("period_type"); pt != "" {
		v, ok := branch_report.ValidPeriodTypes[pt]
		if !ok {
			h.RespondError(w, http.StatusBadRequest, "period_type tidak valid")
			return
		}
		periodType = &v
	}

	var startDate, endDate *time.Time
	if sd := r.URL.Query().Get("start_date"); sd != "" {
		t, err := time.Parse("2006-01-02", sd)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "start_date tidak valid (format YYYY-MM-DD)")
			return
		}
		startDate = &t
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "end_date tidak valid (format YYYY-MM-DD)")
			return
		}
		endDate = &t
	}

	var approvedBy *uuid.UUID
	if ab := r.URL.Query().Get("approved_by"); ab != "" {
		parsed, err := uuid.Parse(ab)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "approved_by tidak valid")
			return
		}
		approvedBy = &parsed
	}

	result, err := querybus.Dispatch[*listbranchreports.Result](r.Context(), h.queryBus, listbranchreports.Query{
		Params:     params,
		BranchID:   branchID,
		PeriodType: periodType,
		StartDate:  startDate,
		EndDate:    endDate,
		ApprovedBy: approvedBy,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary      Get branch report by ID
// @Tags         branch-reports
// @Produce      json
// @Param        id   path      string  true  "Report ID"
// @Success      200  {object}  getbranchreport.Result
// @Router       /api/v1/branch-reports/{id} [get]
func (h *BranchReportHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getbranchreport.Result](r.Context(), h.queryBus, getbranchreport.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create branch financial report
// @Tags         branch-reports
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/branch-reports [post]
func (h *BranchReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BranchID              *string `json:"branch_id"`
		PeriodType            string  `json:"period_type"`
		PeriodStart           string  `json:"period_start"`
		PeriodEnd             string  `json:"period_end"`
		PendapatanOperasional int64   `json:"pendapatan_operasional"`
		PendapatanBungaMargin int64   `json:"pendapatan_bunga_margin"`
		PendapatanLain        int64   `json:"pendapatan_lain"`
		TotalPendapatan       int64   `json:"total_pendapatan"`
		BiayaOperasional      int64   `json:"biaya_operasional"`
		BiayaPersonel         int64   `json:"biaya_personel"`
		BiayaAdministrasi     int64   `json:"biaya_administrasi"`
		BebanPPAP             int64   `json:"beban_ppap"`
		TotalBiaya            int64   `json:"total_biaya"`
		LabaRugiBersih        int64   `json:"laba_rugi_bersih"`
		TotalAset             int64   `json:"total_aset"`
		KasDanBank            int64   `json:"kas_dan_bank"`
		PinjamanDiberikan     int64   `json:"pinjaman_diberikan"`
		SimpananDiterima      int64   `json:"simpanan_diterima"`
		TotalKewajiban        int64   `json:"total_kewajiban"`
		ModalSendiri          int64   `json:"modal_sendiri"`
		NPLRatio              *float64 `json:"npl_ratio"`
		BOPORatio             *float64 `json:"bopo_ratio"`
		ROA                   *float64 `json:"roa"`
		CAR                   *float64 `json:"car"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	var branchID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" {
		parsed, err := uuid.Parse(*req.BranchID)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "branch_id tidak valid")
			return
		}
		branchID = &parsed
	}

	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "period_start tidak valid (format YYYY-MM-DD)")
		return
	}
	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "period_end tidak valid (format YYYY-MM-DD)")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createbranchreport.Command{
		BranchID:               branchID,
		PeriodType:             req.PeriodType,
		PeriodStart:            periodStart,
		PeriodEnd:              periodEnd,
		PendapatanOperasional:  req.PendapatanOperasional,
		PendapatanBungaMargin:  req.PendapatanBungaMargin,
		PendapatanLain:         req.PendapatanLain,
		TotalPendapatan:        req.TotalPendapatan,
		BiayaOperasional:       req.BiayaOperasional,
		BiayaPersonel:          req.BiayaPersonel,
		BiayaAdministrasi:      req.BiayaAdministrasi,
		BebanPPAP:              req.BebanPPAP,
		TotalBiaya:             req.TotalBiaya,
		LabaRugiBersih:         req.LabaRugiBersih,
		TotalAset:              req.TotalAset,
		KasDanBank:             req.KasDanBank,
		PinjamanDiberikan:      req.PinjamanDiberikan,
		SimpananDiterima:       req.SimpananDiterima,
		TotalKewajiban:         req.TotalKewajiban,
		ModalSendiri:           req.ModalSendiri,
		NPLRatio:               req.NPLRatio,
		BOPORatio:              req.BOPORatio,
		ROA:                    req.ROA,
		CAR:                    req.CAR,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update branch financial report
// @Tags         branch-reports
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Report ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/branch-reports/{id} [put]
func (h *BranchReportHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		PendapatanOperasional int64   `json:"pendapatan_operasional"`
		PendapatanBungaMargin int64   `json:"pendapatan_bunga_margin"`
		PendapatanLain        int64   `json:"pendapatan_lain"`
		TotalPendapatan       int64   `json:"total_pendapatan"`
		BiayaOperasional      int64   `json:"biaya_operasional"`
		BiayaPersonel         int64   `json:"biaya_personel"`
		BiayaAdministrasi     int64   `json:"biaya_administrasi"`
		BebanPPAP             int64   `json:"beban_ppap"`
		TotalBiaya            int64   `json:"total_biaya"`
		LabaRugiBersih        int64   `json:"laba_rugi_bersih"`
		TotalAset             int64   `json:"total_aset"`
		KasDanBank            int64   `json:"kas_dan_bank"`
		PinjamanDiberikan     int64   `json:"pinjaman_diberikan"`
		SimpananDiterima      int64   `json:"simpanan_diterima"`
		TotalKewajiban        int64   `json:"total_kewajiban"`
		ModalSendiri          int64   `json:"modal_sendiri"`
		NPLRatio              *float64 `json:"npl_ratio"`
		BOPORatio             *float64 `json:"bopo_ratio"`
		ROA                   *float64 `json:"roa"`
		CAR                   *float64 `json:"car"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updatebranchreport.Command{
		ID:                     id,
		PendapatanOperasional:  req.PendapatanOperasional,
		PendapatanBungaMargin:  req.PendapatanBungaMargin,
		PendapatanLain:         req.PendapatanLain,
		TotalPendapatan:        req.TotalPendapatan,
		BiayaOperasional:       req.BiayaOperasional,
		BiayaPersonel:          req.BiayaPersonel,
		BiayaAdministrasi:      req.BiayaAdministrasi,
		BebanPPAP:              req.BebanPPAP,
		TotalBiaya:             req.TotalBiaya,
		LabaRugiBersih:         req.LabaRugiBersih,
		TotalAset:              req.TotalAset,
		KasDanBank:             req.KasDanBank,
		PinjamanDiberikan:      req.PinjamanDiberikan,
		SimpananDiterima:       req.SimpananDiterima,
		TotalKewajiban:         req.TotalKewajiban,
		ModalSendiri:           req.ModalSendiri,
		NPLRatio:               req.NPLRatio,
		BOPORatio:              req.BOPORatio,
		ROA:                    req.ROA,
		CAR:                    req.CAR,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Approve godoc
// @Summary      Approve branch financial report
// @Tags         branch-reports
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Report ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/branch-reports/{id}/approve [post]
func (h *BranchReportHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		ApprovedBy string `json:"approved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	approvedBy, err := uuid.Parse(req.ApprovedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "approved_by tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), approvebranchreport.Command{
		ID:         id,
		ApprovedBy: approvedBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete branch financial report
// @Tags         branch-reports
// @Param        id   path      string  true  "Report ID"
// @Success      204
// @Router       /api/v1/branch-reports/{id} [delete]
func (h *BranchReportHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletebranchreport.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetConsolidated godoc
// @Summary      Get consolidated report (aggregated all branches)
// @Tags         branch-reports
// @Produce      json
// @Param        period_type   query     string  true  "daily|weekly|monthly|quarterly|yearly"
// @Param        period_start  query     string  true  "Period start (YYYY-MM-DD)"
// @Success      200  {object}  getconsolidated.Result
// @Router       /api/v1/branch-reports/consolidated [get]
func (h *BranchReportHandler) GetConsolidated(w http.ResponseWriter, r *http.Request) {
	ptStr := r.URL.Query().Get("period_type")
	if ptStr == "" {
		h.RespondError(w, http.StatusBadRequest, "period_type wajib diisi")
		return
	}
	pt, ok := branch_report.ValidPeriodTypes[ptStr]
	if !ok {
		h.RespondError(w, http.StatusBadRequest, "period_type tidak valid")
		return
	}

	psStr := r.URL.Query().Get("period_start")
	if psStr == "" {
		h.RespondError(w, http.StatusBadRequest, "period_start wajib diisi")
		return
	}
	periodStart, err := time.Parse("2006-01-02", psStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "period_start tidak valid (format YYYY-MM-DD)")
		return
	}

	result, err := querybus.Dispatch[*getconsolidated.Result](r.Context(), h.queryBus, getconsolidated.Query{
		PeriodType:  pt,
		PeriodStart: periodStart,
	})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// ListEliminationRules godoc
// @Summary      List elimination rules
// @Tags         branch-reports
// @Produce      json
// @Success      200  {object}  listeliminationrules.Result
// @Router       /api/v1/branch-reports/elimination-rules [get]
func (h *BranchReportHandler) ListEliminationRules(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, eliminationRuleSortConfig)

	result, err := querybus.Dispatch[*listeliminationrules.Result](r.Context(), h.queryBus, listeliminationrules.Query{
		Params: params,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateEliminationRule godoc
// @Summary      Create elimination rule
// @Tags         branch-reports
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/branch-reports/elimination-rules [post]
func (h *BranchReportHandler) CreateEliminationRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RuleName     string `json:"rule_name"`
		RuleType     string `json:"rule_type"`
		FromBranchID string `json:"from_branch_id"`
		ToBranchID   string `json:"to_branch_id"`
		Description  string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	fromBranchID, err := uuid.Parse(req.FromBranchID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "from_branch_id tidak valid")
		return
	}
	toBranchID, err := uuid.Parse(req.ToBranchID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "to_branch_id tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createeliminationrule.Command{
		RuleName:     req.RuleName,
		RuleType:     req.RuleType,
		FromBranchID: fromBranchID,
		ToBranchID:   toBranchID,
		Description:  req.Description,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ToggleEliminationRule godoc
// @Summary      Toggle elimination rule active status
// @Tags         branch-reports
// @Param        ruleId   path      string  true  "Rule ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/branch-reports/elimination-rules/{ruleId}/toggle [put]
func (h *BranchReportHandler) ToggleEliminationRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := uuid.Parse(chi.URLParam(r, "ruleId"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "ruleId tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), toggleeliminationrule.Command{ID: ruleID}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": ruleID.String()})
}
