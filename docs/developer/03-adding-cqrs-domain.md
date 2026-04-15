# 03 — Adding a CQRS Domain

Gunakan pola ini untuk domain dengan **business logic kompleks**, workflow multi-step, atau kebutuhan strong consistency. Contoh di bawah menggunakan domain `Order` sebagai ilustrasi.

## Checklist

```
[ ] 1. Buat internal/domain/order/ (entity, errors, events)
[ ] 2. Buat internal/command/create_order/handler.go
[ ] 3. Buat internal/command/update_order/handler.go
[ ] 4. Buat internal/command/delete_order/handler.go
[ ] 5. Buat internal/query/get_order_by_id/handler.go
[ ] 6. Buat internal/query/list_orders/handler.go
[ ] 7. Buat internal/delivery/http/order_handler.go
[ ] 8. Buat infrastructure/database/order_repository.go
[ ] 9. Buat migration SQL
[10. Register di cmd/api/main.go
```

---

## Step 1 — Domain Layer

### `internal/domain/order/entity.go`

```go
// Package order adalah domain untuk manajemen pesanan.
// Aturan layer: ZERO dependency ke package lain kecuali pkg/scope.
package order

import (
    "context"
    "time"

    "github.com/google/uuid"

    "github.com/yourorg/boilerplate/pkg/scope"
)

// Status order.
const (
    StatusPending   = "pending"
    StatusConfirmed = "confirmed"
    StatusCancelled = "cancelled"
    StatusCompleted = "completed"
)

// Order adalah entity utama untuk domain ini.
type Order struct {
    ID          uuid.UUID  `db:"id"`
    CustomerID  uuid.UUID  `db:"customer_id"`
    Status      string     `db:"status"`
    TotalAmount int64      `db:"total_amount"` // dalam satuan terkecil (sen/rupiah)
    Notes       string     `db:"notes"`
    CreatedAt   time.Time  `db:"created_at"`
    UpdatedAt   time.Time  `db:"updated_at"`
    DeletedAt   *time.Time `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain Order.
type WriteRepository interface {
    Save(ctx context.Context, s scope.Scope, o *Order) error
    Update(ctx context.Context, s scope.Scope, o *Order) error
    Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain Order.
type ReadRepository interface {
    GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*Order, error)
    List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*Order, int, error)
}
```

### `internal/domain/order/errors.go`

```go
package order

import "errors"

var (
    ErrNotFound          = errors.New("order tidak ditemukan")
    ErrInvalidStatus     = errors.New("status order tidak valid")
    ErrCustomerIDEmpty   = errors.New("customer_id tidak boleh kosong")
    ErrInvalidAmount     = errors.New("total amount tidak boleh negatif")
    ErrCannotCancel      = errors.New("order sudah selesai tidak bisa dibatalkan")
)
```

### `internal/domain/order/events.go`

```go
package order

import "github.com/google/uuid"

// OrderCreatedEvent dipublikasikan saat Order baru berhasil dibuat.
// TenantID dan CompanyID wajib ada untuk isolasi event handling.
type OrderCreatedEvent struct {
    TenantID    uuid.UUID `json:"tenant_id"`
    CompanyID   uuid.UUID `json:"company_id"`
    ID          uuid.UUID `json:"id"`
    CustomerID  uuid.UUID `json:"customer_id"`
    TotalAmount int64     `json:"total_amount"`
}

func (e OrderCreatedEvent) EventName() string   { return "order.created" }
func (e OrderCreatedEvent) AggregateID() string { return e.ID.String() }

// OrderStatusChangedEvent dipublikasikan saat status Order berubah.
type OrderStatusChangedEvent struct {
    TenantID  uuid.UUID `json:"tenant_id"`
    CompanyID uuid.UUID `json:"company_id"`
    ID        uuid.UUID `json:"id"`
    OldStatus string    `json:"old_status"`
    NewStatus string    `json:"new_status"`
}

func (e OrderStatusChangedEvent) EventName() string   { return "order.status_changed" }
func (e OrderStatusChangedEvent) AggregateID() string { return e.ID.String() }
```

---

## Step 2 — Command Handlers

### `internal/command/create_order/handler.go`

```go
// Package create_order menangani command untuk membuat Order baru.
package create_order

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"

    "github.com/yourorg/boilerplate/internal/domain/order"
    "github.com/yourorg/boilerplate/pkg/commandbus"
    "github.com/yourorg/boilerplate/pkg/eventbus"
    "github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_order"

// Command berisi data yang dibutuhkan untuk membuat Order baru.
type Command struct {
    CustomerID  uuid.UUID
    TotalAmount int64
    Notes       string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_order.
type Handler struct {
    repo     order.WriteRepository
    eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo order.WriteRepository, eb eventbus.EventBus) *Handler {
    return &Handler{repo: repo, eventBus: eb}
}

// Handle memvalidasi command, membuat entity, menyimpan, dan publish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    // Application layer membaca Scope dan meneruskan sebagai parameter eksplisit.
    s, ok := scope.ScopeFromContext(ctx)
    if !ok {
        return scope.ErrMissingTenant
    }
    if err := s.IsValid(); err != nil {
        return err
    }

    // Validasi domain
    if cmd.CustomerID == uuid.Nil {
        return order.ErrCustomerIDEmpty
    }
    if cmd.TotalAmount < 0 {
        return order.ErrInvalidAmount
    }

    id := uuid.New()
    now := time.Now().UTC()

    entity := &order.Order{
        ID:          id,
        CustomerID:  cmd.CustomerID,
        Status:      order.StatusPending,
        TotalAmount: cmd.TotalAmount,
        Notes:       cmd.Notes,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    if err := h.repo.Save(ctx, s, entity); err != nil {
        return fmt.Errorf("save order: %w", err)
    }

    commandbus.SetCreatedID(ctx, id)

    if err := h.eventBus.Publish(ctx, order.OrderCreatedEvent{
        TenantID:    s.TenantID,
        CompanyID:   s.CompanyID,
        ID:          id,
        CustomerID:  cmd.CustomerID,
        TotalAmount: cmd.TotalAmount,
    }); err != nil {
        return fmt.Errorf("publish order.created: %w", err)
    }

    return nil
}
```

### `internal/command/update_order/handler.go`

```go
// Package update_order menangani command untuk mengupdate status Order.
package update_order

import (
    "context"
    "fmt"

    "github.com/google/uuid"

    "github.com/yourorg/boilerplate/internal/domain/order"
    "github.com/yourorg/boilerplate/pkg/eventbus"
    "github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_order"

// Command berisi data untuk mengupdate Order.
type Command struct {
    ID     uuid.UUID
    Status string
    Notes  string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_order.
type Handler struct {
    readRepo  order.ReadRepository
    writeRepo order.WriteRepository
    eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr order.ReadRepository, wr order.WriteRepository, eb eventbus.EventBus) *Handler {
    return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengupdate entity, dan publish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    s, ok := scope.ScopeFromContext(ctx)
    if !ok {
        return scope.ErrMissingTenant
    }
    if err := s.IsValid(); err != nil {
        return err
    }

    validStatuses := map[string]bool{
        order.StatusPending:   true,
        order.StatusConfirmed: true,
        order.StatusCancelled: true,
        order.StatusCompleted: true,
    }
    if !validStatuses[cmd.Status] {
        return order.ErrInvalidStatus
    }

    entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
    if err != nil {
        return fmt.Errorf("get order: %w", err)
    }

    // Business rule: order yang completed tidak bisa dibatalkan
    if entity.Status == order.StatusCompleted && cmd.Status == order.StatusCancelled {
        return order.ErrCannotCancel
    }

    oldStatus := entity.Status
    entity.Status = cmd.Status
    entity.Notes = cmd.Notes

    if err := h.writeRepo.Update(ctx, s, entity); err != nil {
        return fmt.Errorf("update order: %w", err)
    }

    if oldStatus != cmd.Status {
        if err := h.eventBus.Publish(ctx, order.OrderStatusChangedEvent{
            TenantID:  s.TenantID,
            CompanyID: s.CompanyID,
            ID:        cmd.ID,
            OldStatus: oldStatus,
            NewStatus: cmd.Status,
        }); err != nil {
            return fmt.Errorf("publish order.status_changed: %w", err)
        }
    }

    return nil
}
```

---

## Step 3 — Query Handlers

### `internal/query/get_order_by_id/handler.go`

```go
// Package get_order_by_id menangani query untuk mengambil Order berdasarkan ID.
package get_order_by_id

import (
    "context"
    "fmt"

    "github.com/google/uuid"

    "github.com/yourorg/boilerplate/internal/domain/order"
    "github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_order_by_id"

// Query berisi parameter untuk mengambil Order berdasarkan ID.
type Query struct {
    ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
    ID          string `json:"id"`
    CustomerID  string `json:"customer_id"`
    Status      string `json:"status"`
    TotalAmount int64  `json:"total_amount"`
    Notes       string `json:"notes"`
    CreatedAt   string `json:"created_at"`
    UpdatedAt   string `json:"updated_at"`
}

// Handler menangani Query get_order_by_id.
type Handler struct {
    repo order.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo order.ReadRepository) *Handler {
    return &Handler{repo: repo}
}

// Handle mengambil Order berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
    s, ok := scope.ScopeFromContext(ctx)
    if !ok {
        return nil, scope.ErrMissingTenant
    }
    if err := s.IsValid(); err != nil {
        return nil, err
    }

    entity, err := h.repo.GetByID(ctx, s, qry.ID)
    if err != nil {
        return nil, fmt.Errorf("get order by id: %w", err)
    }

    return &Result{
        ID:          entity.ID.String(),
        CustomerID:  entity.CustomerID.String(),
        Status:      entity.Status,
        TotalAmount: entity.TotalAmount,
        Notes:       entity.Notes,
        CreatedAt:   entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
        UpdatedAt:   entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
    }, nil
}
```

### `internal/query/list_orders/handler.go`

```go
// Package list_orders menangani query untuk mengambil daftar Order dengan pagination.
package list_orders

import (
    "context"
    "fmt"

    "github.com/yourorg/boilerplate/internal/domain/order"
    "github.com/yourorg/boilerplate/pkg/pagination"
    "github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_orders"

// Query berisi parameter pagination dan filter untuk list orders.
type Query struct {
    Params pagination.Params
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
    Items      []*OrderItem `json:"items"`
    Total      int          `json:"total"`
    Page       int          `json:"page"`
    PerPage    int          `json:"per_page"`
    TotalPages int          `json:"total_pages"`
}

// OrderItem adalah representasi satu Order dalam list.
type OrderItem struct {
    ID          string `json:"id"`
    CustomerID  string `json:"customer_id"`
    Status      string `json:"status"`
    TotalAmount int64  `json:"total_amount"`
    CreatedAt   string `json:"created_at"`
}

// Handler menangani Query list_orders.
type Handler struct {
    repo order.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo order.ReadRepository) *Handler {
    return &Handler{repo: repo}
}

// Handle mengambil daftar Order dengan pagination.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
    s, ok := scope.ScopeFromContext(ctx)
    if !ok {
        return nil, scope.ErrMissingTenant
    }
    if err := s.IsValid(); err != nil {
        return nil, err
    }

    p := qry.Params
    entities, total, err := h.repo.List(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order)
    if err != nil {
        return nil, fmt.Errorf("list orders: %w", err)
    }

    items := make([]*OrderItem, 0, len(entities))
    for _, e := range entities {
        items = append(items, &OrderItem{
            ID:          e.ID.String(),
            CustomerID:  e.CustomerID.String(),
            Status:      e.Status,
            TotalAmount: e.TotalAmount,
            CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
        })
    }

    return &Result{
        Items:      items,
        Total:      total,
        Page:       p.Page,
        PerPage:    p.Limit,
        TotalPages: pagination.TotalPages(total, p.Limit),
    }, nil
}
```

---

## Step 4 — HTTP Handler

### `internal/delivery/http/order_handler.go`

```go
package http

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"

    createorder "github.com/yourorg/boilerplate/internal/command/create_order"
    deleteorder "github.com/yourorg/boilerplate/internal/command/delete_order"
    updateorder "github.com/yourorg/boilerplate/internal/command/update_order"
    getorder "github.com/yourorg/boilerplate/internal/query/get_order_by_id"
    listorders "github.com/yourorg/boilerplate/internal/query/list_orders"
    "github.com/yourorg/boilerplate/pkg/commandbus"
    "github.com/yourorg/boilerplate/pkg/querybus"
)

// OrderHandler menangani HTTP request untuk domain Order.
// Aturan: TIDAK ada business logic di sini — hanya decode, dispatch, encode.
type OrderHandler struct {
    BaseHandler
}

var orderSortConfig = SortConfig{
    AllowedFields: map[string]string{
        "created_at":   "created_at",
        "total_amount": "total_amount",
        "status":       "status",
    },
    DefaultField: "created_at",
    DefaultOrder: "desc",
}

// NewOrderHandler membuat instance OrderHandler baru.
func NewOrderHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *OrderHandler {
    return &OrderHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Order.
func (h *OrderHandler) RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/orders", func(r chi.Router) {
        r.Get("/", h.List)
        r.Post("/", h.Create)
        r.Get("/{id}", h.GetByID)
        r.Put("/{id}", h.Update)
        r.Delete("/{id}", h.Delete)
    })
}

// List godoc
// @Summary      List orders
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  listorders.Result
// @Router       /api/v1/orders [get]
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
    params := h.ParsePagination(r, orderSortConfig)
    result, err := querybus.Dispatch[*listorders.Result](r.Context(), h.queryBus, listorders.Query{
        Params: params,
    })
    if err != nil {
        h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
        return
    }
    h.RespondJSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary      Get order by ID
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  getorder.Result
// @Router       /api/v1/orders/{id} [get]
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        h.RespondError(w, http.StatusBadRequest, "id tidak valid")
        return
    }
    result, err := querybus.Dispatch[*getorder.Result](r.Context(), h.queryBus, getorder.Query{ID: id})
    if err != nil {
        h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
        return
    }
    h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201  {object}  map[string]string
// @Router       /api/v1/orders [post]
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct {
        CustomerID  string `json:"customer_id"`
        TotalAmount int64  `json:"total_amount"`
        Notes       string `json:"notes"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.RespondError(w, http.StatusBadRequest, "request tidak valid")
        return
    }
    customerID, err := uuid.Parse(req.CustomerID)
    if err != nil {
        h.RespondError(w, http.StatusBadRequest, "customer_id tidak valid")
        return
    }
    ctx, holder := commandbus.WithResultID(r.Context())
    if err := h.commandBus.Dispatch(ctx, createorder.Command{
        CustomerID:  customerID,
        TotalAmount: req.TotalAmount,
        Notes:       req.Notes,
    }); err != nil {
        h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update order status
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/orders/{id} [put]
func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        h.RespondError(w, http.StatusBadRequest, "id tidak valid")
        return
    }
    var req struct {
        Status string `json:"status"`
        Notes  string `json:"notes"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.RespondError(w, http.StatusBadRequest, "request tidak valid")
        return
    }
    if err := h.commandBus.Dispatch(r.Context(), updateorder.Command{
        ID: id, Status: req.Status, Notes: req.Notes,
    }); err != nil {
        h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete order
// @Tags         orders
// @Security     BearerAuth
// @Param        id   path      string  true  "Order ID"
// @Success      204
// @Router       /api/v1/orders/{id} [delete]
func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        h.RespondError(w, http.StatusBadRequest, "id tidak valid")
        return
    }
    if err := h.commandBus.Dispatch(r.Context(), deleteorder.Command{ID: id}); err != nil {
        h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

---

## Step 5 — Repository Implementation

### `infrastructure/database/order_repository.go`

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"

    "github.com/yourorg/boilerplate/internal/domain/order"
    "github.com/yourorg/boilerplate/pkg/scope"
)

// OrderRepository adalah concrete implementation dari order.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data terjamin.
type OrderRepository struct {
    db *sqlx.DB
}

// NewOrderRepository membuat instance baru OrderRepository.
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
    return &OrderRepository{db: db}
}

// Save menyimpan entity Order baru ke database.
func (r *OrderRepository) Save(ctx context.Context, s scope.Scope, o *order.Order) error {
    const q = `
        INSERT INTO orders (id, tenant_id, company_id, customer_id, status, total_amount, notes, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

    _, err := r.db.ExecContext(ctx, q,
        o.ID, s.TenantID, s.CompanyID,
        o.CustomerID, o.Status, o.TotalAmount, o.Notes,
        o.CreatedAt, o.UpdatedAt,
    )
    if err != nil {
        return fmt.Errorf("save order: %w", err)
    }
    return nil
}

// Update mengupdate Order yang sudah ada berdasarkan ID + scope.
func (r *OrderRepository) Update(ctx context.Context, s scope.Scope, o *order.Order) error {
    const q = `
        UPDATE orders
        SET status = $4, notes = $5, updated_at = $6
        WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

    res, err := r.db.ExecContext(ctx, q,
        s.TenantID, s.CompanyID, o.ID,
        o.Status, o.Notes, time.Now().UTC(),
    )
    if err != nil {
        return fmt.Errorf("update order: %w", err)
    }
    if n, _ := res.RowsAffected(); n == 0 {
        return order.ErrNotFound
    }
    return nil
}

// Delete melakukan soft delete.
func (r *OrderRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
    const q = `
        UPDATE orders SET deleted_at = NOW(), updated_at = NOW()
        WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

    res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
    if err != nil {
        return fmt.Errorf("delete order: %w", err)
    }
    if n, _ := res.RowsAffected(); n == 0 {
        return order.ErrNotFound
    }
    return nil
}

// GetByID mengambil satu Order berdasarkan ID + scope.
func (r *OrderRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*order.Order, error) {
    const q = `
        SELECT id, customer_id, status, total_amount, notes, created_at, updated_at, deleted_at
        FROM orders
        WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

    var o order.Order
    err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
        &o.ID, &o.CustomerID, &o.Status, &o.TotalAmount, &o.Notes,
        &o.CreatedAt, &o.UpdatedAt, &o.DeletedAt,
    )
    if err == sql.ErrNoRows {
        return nil, order.ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get order by id: %w", err)
    }
    return &o, nil
}

// List mengambil daftar Order dengan pagination.
func (r *OrderRepository) List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, ord string) ([]*order.Order, int, error) {
    const countQ = `SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
    var total int
    if err := r.db.QueryRowContext(ctx, countQ, s.TenantID, s.CompanyID).Scan(&total); err != nil {
        return nil, 0, fmt.Errorf("count orders: %w", err)
    }

    allowedCols := map[string]string{
        "created_at":   "created_at",
        "total_amount": "total_amount",
        "status":       "status",
    }
    col := safeColumn(sortBy, allowedCols, "created_at")
    dir := safeOrder(ord)

    q := fmt.Sprintf(`
        SELECT id, customer_id, status, total_amount, notes, created_at, updated_at, deleted_at
        FROM orders
        WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
        ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir)

    rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
    if err != nil {
        return nil, 0, fmt.Errorf("list orders: %w", err)
    }
    defer rows.Close()

    var results []*order.Order
    for rows.Next() {
        var o order.Order
        if err := rows.Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalAmount,
            &o.Notes, &o.CreatedAt, &o.UpdatedAt, &o.DeletedAt); err != nil {
            return nil, 0, err
        }
        results = append(results, &o)
    }
    return results, total, rows.Err()
}
```

---

## Step 6 — Migration SQL

Buat file migration:

```bash
cd api
make migrate-create name=create_orders
# Output: Created migrations/005_create_orders.sql
```

Edit file yang baru dibuat:

### `migrations/005_create_orders.sql`

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS orders (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID         NOT NULL,
    company_id   UUID         NOT NULL,
    customer_id  UUID         NOT NULL,
    status       VARCHAR(50)  NOT NULL DEFAULT 'pending',
    total_amount BIGINT       NOT NULL DEFAULT 0,
    notes        TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

-- tenant_id sebagai leading column di semua composite index
CREATE INDEX IF NOT EXISTS idx_orders_tenant_company_created
    ON orders(tenant_id, company_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_orders_tenant_company_status
    ON orders(tenant_id, company_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_orders_tenant_company_customer
    ON orders(tenant_id, company_id, customer_id)
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS orders;
```

---

## Step 7 — Register di main.go

Edit `cmd/api/main.go`:

```go
import (
    // ... existing imports ...

    createorder "github.com/yourorg/boilerplate/internal/command/create_order"
    deleteorder "github.com/yourorg/boilerplate/internal/command/delete_order"
    updateorder "github.com/yourorg/boilerplate/internal/command/update_order"
    getorder    "github.com/yourorg/boilerplate/internal/query/get_order_by_id"
    listorders  "github.com/yourorg/boilerplate/internal/query/list_orders"
    deliveryhttp "github.com/yourorg/boilerplate/internal/delivery/http"
)

// Di dalam fx.Provide:
deliveryhttp.NewOrderHandler,

// Di dalam registerCommandHandlers():
func registerCommandHandlers(bus *commandbus.CommandBus, db *sqlx.DB, eb eventbus.EventBus, logger zerolog.Logger) {
    // ... existing handlers ...
    orderRepo := database.NewOrderRepository(db)
    commandbus.Register(bus, createorder.NewHandler(orderRepo, eb))
    commandbus.Register(bus, updateorder.NewHandler(orderRepo, orderRepo, eb))
    commandbus.Register(bus, deleteorder.NewHandler(orderRepo))
}

// Di dalam registerQueryHandlers():
func registerQueryHandlers(bus *querybus.QueryBus, db *sqlx.DB, logger zerolog.Logger) {
    // ... existing handlers ...
    orderRepo := database.NewOrderRepository(db)
    querybus.Register(bus, getorder.NewHandler(orderRepo))
    querybus.Register(bus, listorders.NewHandler(orderRepo))
}

// Di dalam newRouter() (server.go):
func newRouter(..., orderHandler *deliveryhttp.OrderHandler, ...) chi.Router {
    // ...
    r.Group(func(r chi.Router) {
        // ... auth middleware ...
        orderHandler.RegisterRoutes(r)
    })
}
```

---

## Step 8 — Jalankan Migrasi

```bash
make migrate-up
```

Verifikasi:

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"<uuid>","total_amount":150000,"notes":"test order"}'
```

---

## Aturan yang Tidak Boleh Dilanggar

| Larangan | Alasan |
|----------|--------|
| Business logic di HTTP handler | Melanggar SRP, sulit ditest |
| Baca DB langsung dari command/query handler | Melanggar dependency rule |
| Repo baca scope dari `ctx.Value()` | Hidden dependency, sulit di-mock |
| Magic string untuk event name | Gunakan konstanta di `events.go` |
| `TenantID`/`CompanyID` di dalam entity struct | Tenancy adalah infrastruktur concern |
| JOIN antar tabel di command handler | Bukan tanggung jawab application layer |
