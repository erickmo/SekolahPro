package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	approvereservetransaction "github.com/yourorg/boilerplate/internal/command/approve_reserve_transaction"
	createreservefund "github.com/yourorg/boilerplate/internal/command/create_reserve_fund"
	createreservetransaction "github.com/yourorg/boilerplate/internal/command/create_reserve_transaction"
	deletereservefund "github.com/yourorg/boilerplate/internal/command/delete_reserve_fund"
	freezereservefund "github.com/yourorg/boilerplate/internal/command/freeze_reserve_fund"
	rejectreservetransaction "github.com/yourorg/boilerplate/internal/command/reject_reserve_transaction"
	updatereservefund "github.com/yourorg/boilerplate/internal/command/update_reserve_fund"
	getreservefund "github.com/yourorg/boilerplate/internal/query/get_reserve_fund_by_id"
	listreservefunds "github.com/yourorg/boilerplate/internal/query/list_reserve_funds"
	listreservetransactions "github.com/yourorg/boilerplate/internal/query/list_reserve_transactions"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// ReserveHandler menangani HTTP request untuk domain Reserve Fund.
type ReserveHandler struct {
	BaseHandler
}

var reserveFundSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"name":       "name",
		"fund_type":  "fund_type",
		"status":     "status",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewReserveHandler membuat instance ReserveHandler baru.
func NewReserveHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *ReserveHandler {
	return &ReserveHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Reserve Fund.
func (h *ReserveHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/reserve-funds", func(r chi.Router) {
		// Fund CRUD
		r.Get("/", h.ListFunds)
		r.Post("/", h.CreateFund)
		r.Get("/{id}", h.GetFundByID)
		r.Put("/{id}", h.UpdateFund)
		r.Delete("/{id}", h.DeleteFund)
		r.Put("/{id}/freeze", h.FreezeFund)

		// Transactions
		r.Post("/transactions", h.CreateTransaction)
		r.Get("/transactions", h.ListTransactions)
		r.Put("/transactions/{id}/approve", h.ApproveTransaction)
		r.Put("/transactions/{id}/reject", h.RejectTransaction)
	})
}

// ListFunds godoc
// @Summary      List reserve funds
// @Tags         reserve-funds
// @Produce      json
// @Param        fund_type  query  string  false  "Filter by fund type"
// @Param        status     query  string  false  "Filter by status"
// @Success      200  {object}  listreservefunds.Result
// @Router       /api/v1/reserve-funds [get]
func (h *ReserveHandler) ListFunds(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, reserveFundSortConfig)
	filter := h.parseFundFilter(r)
	result, err := querybus.Dispatch[*listreservefunds.Result](r.Context(), h.queryBus, listreservefunds.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetFundByID godoc
// @Summary      Get reserve fund by ID
// @Tags         reserve-funds
// @Produce      json
// @Param        id   path      string  true  "Reserve Fund ID"
// @Success      200  {object}  getreservefund.Result
// @Router       /api/v1/reserve-funds/{id} [get]
func (h *ReserveHandler) GetFundByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getreservefund.Result](r.Context(), h.queryBus, getreservefund.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateFund godoc
// @Summary      Create reserve fund
// @Tags         reserve-funds
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/reserve-funds [post]
func (h *ReserveHandler) CreateFund(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string  `json:"name"`
		FundType        string  `json:"fund_type"`
		TargetAmount    int64   `json:"target_amount"`
		MinimumBalance  int64   `json:"minimum_balance"`
		ContributionPct float64 `json:"contribution_pct"`
		Description     string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createreservefund.Command{
		Name:            req.Name,
		FundType:        req.FundType,
		TargetAmount:    req.TargetAmount,
		MinimumBalance:  req.MinimumBalance,
		ContributionPct: req.ContributionPct,
		Description:     req.Description,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateFund godoc
// @Summary      Update reserve fund
// @Tags         reserve-funds
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Reserve Fund ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/reserve-funds/{id} [put]
func (h *ReserveHandler) UpdateFund(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Name            string  `json:"name"`
		FundType        string  `json:"fund_type"`
		TargetAmount    int64   `json:"target_amount"`
		MinimumBalance  int64   `json:"minimum_balance"`
		ContributionPct float64 `json:"contribution_pct"`
		Description     string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), updatereservefund.Command{
		ID:              id,
		Name:            req.Name,
		FundType:        req.FundType,
		TargetAmount:    req.TargetAmount,
		MinimumBalance:  req.MinimumBalance,
		ContributionPct: req.ContributionPct,
		Description:     req.Description,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// DeleteFund godoc
// @Summary      Delete reserve fund
// @Tags         reserve-funds
// @Param        id   path      string  true  "Reserve Fund ID"
// @Success      204
// @Router       /api/v1/reserve-funds/{id} [delete]
func (h *ReserveHandler) DeleteFund(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletereservefund.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// FreezeFund godoc
// @Summary      Freeze reserve fund
// @Tags         reserve-funds
// @Param        id   path      string  true  "Reserve Fund ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/reserve-funds/{id}/freeze [put]
func (h *ReserveHandler) FreezeFund(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), freezereservefund.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// CreateTransaction godoc
// @Summary      Create reserve fund transaction
// @Tags         reserve-funds
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/reserve-funds/transactions [post]
func (h *ReserveHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ReserveFundID   string `json:"reserve_fund_id"`
		TransactionType string `json:"transaction_type"`
		Amount          int64  `json:"amount"`
		ReferenceNo     string `json:"reference_no"`
		Description     string `json:"description"`
		ProcessedBy     string `json:"processed_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	fundID, err := uuid.Parse(req.ReserveFundID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "reserve_fund_id tidak valid")
		return
	}
	processedBy, err := uuid.Parse(req.ProcessedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "processed_by tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createreservetransaction.Command{
		ReserveFundID:   fundID,
		TransactionType: reserve.TransactionType(req.TransactionType),
		Amount:          req.Amount,
		ReferenceNo:     req.ReferenceNo,
		Description:     req.Description,
		ProcessedBy:     processedBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ListTransactions godoc
// @Summary      List reserve fund transactions
// @Tags         reserve-funds
// @Produce      json
// @Param        reserve_fund_id    query  string  false  "Filter by reserve fund ID"
// @Param        transaction_type   query  string  false  "Filter by transaction type"
// @Success      200  {object}  listreservetransactions.Result
// @Router       /api/v1/reserve-funds/transactions [get]
func (h *ReserveHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, reserveFundSortConfig)
	filter := h.parseTransactionFilter(r)
	result, err := querybus.Dispatch[*listreservetransactions.Result](r.Context(), h.queryBus, listreservetransactions.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// ApproveTransaction godoc
// @Summary      Approve reserve fund transaction
// @Tags         reserve-funds
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/reserve-funds/transactions/{id}/approve [put]
func (h *ReserveHandler) ApproveTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), approvereservetransaction.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// RejectTransaction godoc
// @Summary      Reject reserve fund transaction
// @Tags         reserve-funds
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/reserve-funds/transactions/{id}/reject [put]
func (h *ReserveHandler) RejectTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), rejectreservetransaction.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// parseFundFilter membaca filter fund dari query parameters.
func (h *ReserveHandler) parseFundFilter(r *http.Request) reserve.ReserveFundFilter {
	var filter reserve.ReserveFundFilter
	q := r.URL.Query()

	if ft := q.Get("fund_type"); ft != "" {
		filter.FundType = &ft
	}
	if st := q.Get("status"); st != "" {
		s := reserve.FundStatus(st)
		filter.Status = &s
	}

	return filter
}

// parseTransactionFilter membaca filter transaksi dari query parameters.
func (h *ReserveHandler) parseTransactionFilter(r *http.Request) reserve.ReserveFundTransactionFilter {
	var filter reserve.ReserveFundTransactionFilter
	q := r.URL.Query()

	if fid := q.Get("reserve_fund_id"); fid != "" {
		id, err := uuid.Parse(fid)
		if err == nil {
			filter.ReserveFundID = &id
		}
	}
	if tt := q.Get("transaction_type"); tt != "" {
		t := reserve.TransactionType(tt)
		filter.TransactionType = &t
	}

	return filter
}
