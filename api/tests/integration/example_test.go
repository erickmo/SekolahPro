//go:build integration

// Package integration berisi integration tests yang membutuhkan PostgreSQL container.
// Jalankan dengan: go test ./tests/integration/... -v -tags integration
package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yourorg/boilerplate/infrastructure/database"
	"github.com/yourorg/boilerplate/internal/domain/example"
	"github.com/yourorg/boilerplate/pkg/scope"
)

var (
	testDB      *sqlx.DB
	testScope   scope.Scope
	pgContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer pgContainer.Terminate(ctx) //nolint:errcheck

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	testDB, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test db: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := runMigrations(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "migrations failed: %v\n", err)
		os.Exit(1)
	}

	testScope = scope.Scope{
		TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	}

	os.Exit(m.Run())
}

// runMigrations membuat tabel yang dibutuhkan untuk testing.
func runMigrations(ctx context.Context) error {
	const createExamplesTable = `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS examples (
			id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id   UUID         NOT NULL,
			company_id  UUID         NOT NULL,
			branch_id   UUID,
			name        VARCHAR(255) NOT NULL,
			description TEXT         NOT NULL DEFAULT '',
			is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			deleted_at  TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_examples_tenant_company
			ON examples(tenant_id, company_id, created_at DESC);
	`
	_, err := testDB.ExecContext(ctx, createExamplesTable)
	return err
}

// ── Tests ─────────────────────────────────────────────────────────────────────

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

	got, err := repo.GetByID(ctx, testScope, e.ID)
	if err != nil {
		t.Fatalf("GetByID gagal: %v", err)
	}
	if got.Name != e.Name {
		t.Errorf("Name mismatch: got %q, want %q", got.Name, e.Name)
	}
	if got.Description != e.Description {
		t.Errorf("Description mismatch: got %q, want %q", got.Description, e.Description)
	}
}

func TestExampleRepository_Update(t *testing.T) {
	repo := database.NewExampleRepository(testDB)
	ctx := context.Background()

	e := &example.Example{
		ID: uuid.New(), Name: "Before Update", Description: "Old",
		IsActive: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Save(ctx, testScope, e); err != nil {
		t.Fatalf("Save gagal: %v", err)
	}

	e.Name = "After Update"
	e.Description = "New"
	if err := repo.Update(ctx, testScope, e); err != nil {
		t.Fatalf("Update gagal: %v", err)
	}

	got, err := repo.GetByID(ctx, testScope, e.ID)
	if err != nil {
		t.Fatalf("GetByID gagal: %v", err)
	}
	if got.Name != "After Update" {
		t.Errorf("Name mismatch setelah update: got %q", got.Name)
	}
}

func TestExampleRepository_Delete(t *testing.T) {
	repo := database.NewExampleRepository(testDB)
	ctx := context.Background()

	e := &example.Example{
		ID: uuid.New(), Name: "To Delete",
		IsActive: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Save(ctx, testScope, e); err != nil {
		t.Fatalf("Save gagal: %v", err)
	}

	if err := repo.Delete(ctx, testScope, e.ID); err != nil {
		t.Fatalf("Delete gagal: %v", err)
	}

	_, err := repo.GetByID(ctx, testScope, e.ID)
	if err != example.ErrNotFound {
		t.Errorf("setelah delete, expected ErrNotFound, got: %v", err)
	}
}

func TestExampleRepository_List(t *testing.T) {
	repo := database.NewExampleRepository(testDB)
	ctx := context.Background()

	// Buat beberapa examples
	for i := range 3 {
		e := &example.Example{
			ID: uuid.New(), Name: fmt.Sprintf("List Test %d", i),
			IsActive: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		if err := repo.Save(ctx, testScope, e); err != nil {
			t.Fatalf("Save gagal untuk item %d: %v", i, err)
		}
	}

	items, total, err := repo.List(ctx, testScope, 100, 0, "created_at", "desc")
	if err != nil {
		t.Fatalf("List gagal: %v", err)
	}
	if total < 3 {
		t.Errorf("total mismatch: got %d, want >= 3", total)
	}
	if len(items) < 3 {
		t.Errorf("len(items) mismatch: got %d, want >= 3", len(items))
	}
}

func TestExampleRepository_GetByID_NotFound(t *testing.T) {
	repo := database.NewExampleRepository(testDB)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testScope, uuid.New())
	if err != example.ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestExampleRepository_ScopeisolationID(t *testing.T) {
	repo := database.NewExampleRepository(testDB)
	ctx := context.Background()

	// Simpan dengan testScope
	e := &example.Example{
		ID: uuid.New(), Name: "Scope Isolation Test",
		IsActive: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Save(ctx, testScope, e); err != nil {
		t.Fatalf("Save gagal: %v", err)
	}

	// Coba baca dengan scope berbeda — harus gagal
	otherScope := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	_, err := repo.GetByID(ctx, otherScope, e.ID)
	if err != example.ErrNotFound {
		t.Errorf("cross-scope access seharusnya ErrNotFound, got: %v", err)
	}
}
