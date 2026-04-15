# 06 — Testing Guide

## Overview

Boilerplate ini menggunakan tiga level testing:

| Level | Tools | Target | Speed |
|-------|-------|--------|-------|
| Unit (Go) | `go test`, `testify` | Command/Query handlers, domain logic | Fast (< 1s) |
| Integration (Go) | `testcontainers-go`, PostgreSQL | Repository implementations | Medium (~10-30s) |
| Unit/Component (Web) | `Vitest`, `MSW`, `React Testing Library` | React components, hooks, services | Fast (< 5s) |
| E2E (Web) | `Playwright` | Full user flows | Slow (~1-5min) |

---

## Go — Unit Tests

Unit test menarget logic di layer **command**, **query**, dan **domain** tanpa menyentuh database.

### Struktur Unit Test

```
api/
├── internal/
│   ├── command/create_example/
│   │   └── handler_test.go       # test command handler dengan mock repo
│   ├── query/get_example_by_id/
│   │   └── handler_test.go       # test query handler dengan mock repo
│   └── domain/example/
│       └── entity_test.go        # test domain logic (validasi, dll)
└── pkg/
    ├── commandbus/
    │   └── commandbus_test.go
    └── scope/
        └── scope_test.go
```

### Contoh Unit Test — Command Handler

```go
// internal/command/create_example/handler_test.go

package create_example_test

import (
    "context"
    "errors"
    "testing"

    "github.com/google/uuid"

    "github.com/yourorg/boilerplate/internal/command/create_example"
    "github.com/yourorg/boilerplate/internal/domain/example"
    "github.com/yourorg/boilerplate/pkg/scope"
)

// mockWriteRepo adalah mock sederhana untuk example.WriteRepository.
// Tidak perlu library mock — cukup struct dengan state.
type mockWriteRepo struct {
    savedEntity *example.Example
    saveErr     error
}

func (m *mockWriteRepo) Save(_ context.Context, _ scope.Scope, e *example.Example) error {
    m.savedEntity = e
    return m.saveErr
}

func (m *mockWriteRepo) Update(_ context.Context, _ scope.Scope, e *example.Example) error {
    return nil
}

func (m *mockWriteRepo) Delete(_ context.Context, _ scope.Scope, id uuid.UUID) error {
    return nil
}

// mockEventBus adalah mock untuk eventbus.EventBus.
type mockEventBus struct {
    published []any
    publishErr error
}

func (m *mockEventBus) Publish(_ context.Context, event any) error {
    m.published = append(m.published, event)
    return m.publishErr
}

func (m *mockEventBus) Subscribe(_ context.Context, _ string, _ any) error { return nil }

// testScope adalah scope default untuk testing.
var testScope = scope.Scope{
    TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
    CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
}

func TestCreateExample_Success(t *testing.T) {
    repo := &mockWriteRepo{}
    eb := &mockEventBus{}
    handler := create_example.NewHandler(repo, eb)

    ctx := scope.WithScope(context.Background(), testScope)
    err := handler.Handle(ctx, create_example.Command{
        Name:        "Test Example",
        Description: "Deskripsi test",
    })

    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if repo.savedEntity == nil {
        t.Fatal("expected entity to be saved")
    }
    if repo.savedEntity.Name != "Test Example" {
        t.Errorf("Name mismatch: got %q, want %q", repo.savedEntity.Name, "Test Example")
    }
    if len(eb.published) != 1 {
        t.Errorf("expected 1 event published, got %d", len(eb.published))
    }
}

func TestCreateExample_NameEmpty(t *testing.T) {
    repo := &mockWriteRepo{}
    eb := &mockEventBus{}
    handler := create_example.NewHandler(repo, eb)

    ctx := scope.WithScope(context.Background(), testScope)
    err := handler.Handle(ctx, create_example.Command{Name: ""})

    if !errors.Is(err, example.ErrNameEmpty) {
        t.Errorf("expected ErrNameEmpty, got: %v", err)
    }
    if repo.savedEntity != nil {
        t.Error("expected no entity saved on validation error")
    }
}

func TestCreateExample_MissingScope(t *testing.T) {
    repo := &mockWriteRepo{}
    eb := &mockEventBus{}
    handler := create_example.NewHandler(repo, eb)

    // Context tanpa scope
    err := handler.Handle(context.Background(), create_example.Command{Name: "Test"})

    if !errors.Is(err, scope.ErrMissingTenant) {
        t.Errorf("expected ErrMissingTenant, got: %v", err)
    }
}

func TestCreateExample_SaveError(t *testing.T) {
    dbErr := errors.New("connection refused")
    repo := &mockWriteRepo{saveErr: dbErr}
    eb := &mockEventBus{}
    handler := create_example.NewHandler(repo, eb)

    ctx := scope.WithScope(context.Background(), testScope)
    err := handler.Handle(ctx, create_example.Command{Name: "Test"})

    if err == nil {
        t.Fatal("expected error when repo fails")
    }
}
```

### Menjalankan Unit Tests

```bash
cd api

# Semua unit tests
make test
# Setara: go test ./internal/... ./pkg/... -v -timeout 30s

# Dengan race detector (deteksi data race)
make test-race
# Setara: go test -race ./... -timeout 60s

# Test package spesifik
go test ./internal/command/create_example/... -v

# Test fungsi spesifik
go test ./internal/command/create_example/... -run TestCreateExample_Success -v

# Test dengan coverage report
go test ./internal/... ./pkg/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Go — Integration Tests

Integration test menggunakan **Testcontainers** untuk spin up PostgreSQL container secara otomatis. Tidak perlu database yang sudah berjalan di luar.

### Prasyarat

- Docker harus berjalan
- Build tag `integration` diperlukan agar test tidak berjalan bersama unit test

### Struktur Integration Test

```go
// tests/integration/example_test.go

//go:build integration  // ← build tag wajib

package integration

import (
    "context"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
    // ...
)

// TestMain menyiapkan PostgreSQL container dan menjalankan semua test.
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Spin up PostgreSQL 17 container
    pgContainer, err = postgres.Run(ctx,
        "postgres:17-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpassword"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    // ...

    // Jalankan migrasi di container
    runMigrations(ctx)

    // Scope untuk semua test di file ini
    testScope = scope.Scope{
        TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
        CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
    }

    os.Exit(m.Run())
}
```

### Contoh Integration Test — Repository

```go
func TestExampleRepository_Save(t *testing.T) {
    repo := database.NewExampleRepository(testDB)
    ctx := context.Background()

    e := &example.Example{
        ID:          uuid.New(),
        Name:        "Test Example",
        Description: "Deskripsi test",
        IsActive:    true,
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    }

    if err := repo.Save(ctx, testScope, e); err != nil {
        t.Fatalf("Save gagal: %v", err)
    }

    // Verifikasi data tersimpan
    got, err := repo.GetByID(ctx, testScope, e.ID)
    if err != nil {
        t.Fatalf("GetByID gagal: %v", err)
    }
    if got.Name != e.Name {
        t.Errorf("Name mismatch: got %q, want %q", got.Name, e.Name)
    }
}

// Test isolasi scope — penting untuk keamanan multi-tenant
func TestExampleRepository_ScopeIsolation(t *testing.T) {
    repo := database.NewExampleRepository(testDB)
    ctx := context.Background()

    // Simpan dengan testScope
    e := &example.Example{ID: uuid.New(), Name: "Scope Test", IsActive: true,
        CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
    if err := repo.Save(ctx, testScope, e); err != nil {
        t.Fatalf("Save gagal: %v", err)
    }

    // Coba baca dengan scope berbeda — HARUS gagal
    otherScope := scope.Scope{TenantID: uuid.New(), CompanyID: uuid.New()}
    _, err := repo.GetByID(ctx, otherScope, e.ID)
    if err != example.ErrNotFound {
        t.Errorf("cross-scope access harus ErrNotFound, got: %v", err)
    }
}
```

### Menambah Integration Test untuk Domain Baru

Buat file baru di `tests/integration/`:

```go
// tests/integration/order_test.go

//go:build integration

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/yourorg/boilerplate/infrastructure/database"
    "github.com/yourorg/boilerplate/internal/domain/order"
)

func TestOrderRepository_CRUD(t *testing.T) {
    repo := database.NewOrderRepository(testDB)
    ctx := context.Background()

    // Create
    o := &order.Order{
        ID:          uuid.New(),
        CustomerID:  uuid.New(),
        Status:      order.StatusPending,
        TotalAmount: 150000,
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    }
    if err := repo.Save(ctx, testScope, o); err != nil {
        t.Fatalf("Save gagal: %v", err)
    }

    // Read
    got, err := repo.GetByID(ctx, testScope, o.ID)
    if err != nil {
        t.Fatalf("GetByID gagal: %v", err)
    }
    if got.Status != order.StatusPending {
        t.Errorf("Status mismatch: got %q, want %q", got.Status, order.StatusPending)
    }

    // Update
    o.Status = order.StatusConfirmed
    if err := repo.Update(ctx, testScope, o); err != nil {
        t.Fatalf("Update gagal: %v", err)
    }

    // Delete
    if err := repo.Delete(ctx, testScope, o.ID); err != nil {
        t.Fatalf("Delete gagal: %v", err)
    }
    _, err = repo.GetByID(ctx, testScope, o.ID)
    if err != order.ErrNotFound {
        t.Errorf("setelah delete, expected ErrNotFound, got: %v", err)
    }
}
```

### Menjalankan Integration Tests

```bash
cd api

# Jalankan semua integration tests (butuh Docker)
make test-integration
# Setara: go test ./tests/integration/... -v -timeout 120s -tags integration

# Test file spesifik
go test ./tests/integration/... -v -tags integration -run TestOrder

# Dengan verbose output Testcontainers
TESTCONTAINERS_RYUK_DISABLED=true go test ./tests/integration/... -v -tags integration
```

---

## Web Dashboard — Unit dan Component Tests

Menggunakan **Vitest** sebagai test runner dan **React Testing Library** + **MSW** untuk component testing.

### Setup

File konfigurasi sudah ada:
- `src/__ui_tests__/setup.ts` — global test setup
- `src/__ui_tests__/mocks/handlers.ts` — MSW request handlers
- `src/__ui_tests__/mocks/server.ts` — MSW server setup
- `src/__ui_tests__/test-utils.tsx` — custom render dengan providers

### Contoh Component Test

```tsx
// src/pages/Dashboard/__tests__/DashboardPage.test.tsx

import { describe, it, expect } from 'vitest'
import { screen } from '@testing-library/react'
import { render } from '@/__ui_tests__/test-utils'
import { DashboardPage } from '../DashboardPage'

describe('DashboardPage', () => {
  it('renders welcome message', () => {
    render(<DashboardPage />)
    expect(screen.getByText(/dashboard/i)).toBeInTheDocument()
  })

  it('shows loading state', () => {
    render(<DashboardPage />)
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
  })
})
```

### Contoh Service Test dengan MSW

```tsx
// src/services/__tests__/createEntityService.test.ts

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { server } from '@/__ui_tests__/mocks/server'
import { http, HttpResponse } from 'msw'
import { createEntityService } from '../createEntityService'

const userService = createEntityService<{ id: string; name: string }>('/api/users')

describe('createEntityService', () => {
  beforeAll(() => server.listen())
  afterEach(() => server.resetHandlers())
  afterAll(() => server.close())

  it('fetches list successfully', async () => {
    server.use(
      http.get('/api/users', () =>
        HttpResponse.json({ items: [{ id: '1', name: 'Alice' }], total: 1 })
      )
    )

    const result = await userService.list({ limit: 10 })
    expect(result.items).toHaveLength(1)
    expect(result.items[0].name).toBe('Alice')
  })

  it('creates entity successfully', async () => {
    server.use(
      http.post('/api/users', () =>
        HttpResponse.json({ id: '2', name: 'Bob' }, { status: 201 })
      )
    )

    const result = await userService.create({ name: 'Bob' })
    expect(result.id).toBe('2')
  })
})
```

### Contoh Hook Test

```tsx
// src/hooks/__tests__/useForm.test.ts

import { describe, it, expect } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useForm } from '../useForm'

describe('useForm', () => {
  it('initializes dengan values awal', () => {
    const { result } = renderHook(() =>
      useForm({ initialValues: { name: '', email: '' } })
    )

    expect(result.current.values.name).toBe('')
    expect(result.current.values.email).toBe('')
  })

  it('updates field value', () => {
    const { result } = renderHook(() =>
      useForm({ initialValues: { name: '' } })
    )

    act(() => {
      result.current.handleChange('name', 'Alice')
    })

    expect(result.current.values.name).toBe('Alice')
  })

  it('validates required fields', async () => {
    const { result } = renderHook(() =>
      useForm({
        initialValues: { name: '' },
        validate: (v) => v.name === '' ? { name: 'Required' } : {},
      })
    )

    await act(async () => {
      await result.current.handleSubmit(async () => {})
    })

    expect(result.current.errors.name).toBe('Required')
  })
})
```

### Menjalankan Web Tests

```bash
cd web-dashboard

# Semua unit/component tests
npm run test
# Setara: vitest

# Watch mode (untuk development)
npm run test -- --watch

# Coverage report
npm run test -- --coverage

# Test file spesifik
npm run test -- src/hooks/__tests__/useForm.test.ts

# UI mode (interactive)
npm run test -- --ui
```

---

## Web Dashboard — E2E Tests (Playwright)

### Setup

Playwright dikonfigurasi untuk test full user flow dengan browser nyata.

```typescript
// playwright.config.ts (atau e2e/)
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  use: {
    baseURL: 'http://localhost:5173',
  },
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
  },
})
```

### Contoh E2E Test

```typescript
// e2e/login.spec.ts

import { test, expect } from '@playwright/test'

test.describe('Login Flow', () => {
  test('login berhasil dan redirect ke dashboard', async ({ page }) => {
    await page.goto('/login')

    await page.fill('[name="email"]', 'admin@example.com')
    await page.fill('[name="password"]', 'password')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('h1')).toContainText('Dashboard')
  })

  test('login gagal menampilkan error', async ({ page }) => {
    await page.goto('/login')

    await page.fill('[name="email"]', 'wrong@example.com')
    await page.fill('[name="password"]', 'wrongpassword')
    await page.click('button[type="submit"]')

    await expect(page.locator('[role="alert"]')).toBeVisible()
    await expect(page).toHaveURL('/login')
  })

  test('halaman protected redirect ke login', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page).toHaveURL('/login')
  })
})
```

### Menjalankan E2E Tests

```bash
cd web-dashboard

# Jalankan semua E2E tests (headless)
npx playwright test

# Dengan UI mode (bisa lihat browser)
npx playwright test --ui

# Test file spesifik
npx playwright test e2e/login.spec.ts

# Debug mode
npx playwright test --debug

# Generate report HTML
npx playwright test --reporter=html
npx playwright show-report
```

---

## CI/CD Test Strategy

Rekomendasi urutan test di CI pipeline:

```yaml
# .github/workflows/ci.yml (contoh)

jobs:
  test-go:
    steps:
      - run: make lint                  # 1. Lint dulu
      - run: make test                  # 2. Unit tests (fast)
      - run: make test-integration      # 3. Integration tests (butuh Docker)

  test-web:
    steps:
      - run: npm run lint               # 1. ESLint
      - run: npm run test               # 2. Vitest unit tests
      - run: npx playwright install --with-deps
      - run: npx playwright test        # 3. E2E tests (butuh server)
```

---

## Tips Testing

### Mock yang Baik

```go
// Buat mock sederhana tanpa library tambahan untuk dependencies yang murni
type mockRepo struct {
    entities map[uuid.UUID]*example.Example
    err      error
}

func (m *mockRepo) GetByID(_ context.Context, _ scope.Scope, id uuid.UUID) (*example.Example, error) {
    if m.err != nil {
        return nil, m.err
    }
    e, ok := m.entities[id]
    if !ok {
        return nil, example.ErrNotFound
    }
    return e, nil
}
```

### Table-Driven Tests

```go
func TestValidateName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {"valid", "Test Name", nil},
        {"empty", "", example.ErrNameEmpty},
        {"too long", strings.Repeat("a", 256), example.ErrNameTooLong},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateName(tt.input)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Scope Isolation Selalu

Setiap integration test yang menyimpan data harus memverifikasi bahwa data tidak bisa diakses dari scope lain. Ini adalah test terpenting untuk keamanan multi-tenant.
