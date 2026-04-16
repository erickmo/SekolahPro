//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/academic_year"
	"github.com/yourorg/boilerplate/internal/domain/class_room"
	"github.com/yourorg/boilerplate/internal/domain/teacher"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Vernon Test Helpers ─────────────────────────────────────────────────────

func newTestEventBus(t *testing.T) eventbus.EventBus {
	t.Helper()
	eb, err := eventbus.NewInMemoryEventBus()
	require.NoError(t, err)
	return eb
}

func newTestLogger() zerolog.Logger {
	return zerolog.Nop()
}

type vernonTestCtx struct {
	repo *vernon.BaseRepository
	svc  *vernon.BaseService
}

func setupVernonService(t *testing.T, desc vernon.DomainDescriptor) vernonTestCtx {
	t.Helper()
	repo := vernon.NewBaseRepository(testDB, desc.TableName())
	svc := vernon.NewBaseService(testDB, repo, desc, newTestEventBus(t), newTestLogger())
	return vernonTestCtx{repo: repo, svc: svc}
}

func setupAllVernonServices(t *testing.T) (
	aySvc *vernon.BaseService,
	tSvc *vernon.BaseService,
	crSvc *vernon.BaseService,
) {
	t.Helper()
	ayRepo := vernon.NewBaseRepository(testDB, "academic_years")
	aySvc = vernon.NewBaseService(testDB, ayRepo, &academic_year.Descriptor{}, newTestEventBus(t), newTestLogger())

	tRepo := vernon.NewBaseRepository(testDB, "teachers")
	tSvc = vernon.NewBaseService(testDB, tRepo, &teacher.Descriptor{}, newTestEventBus(t), newTestLogger())

	crRepo := vernon.NewBaseRepository(testDB, "class_rooms")
	crSvc = vernon.NewBaseService(testDB, crRepo, &class_room.Descriptor{}, newTestEventBus(t), newTestLogger())

	return aySvc, tSvc, crSvc
}

// ════════════════════════════════════════════════════════════════════════════
// ACADEMIC YEARS
// ════════════════════════════════════════════════════════════════════════════

func TestVernonAcademicYear_Create_Success(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	data := map[string]any{
		"name":            "2025/2026",
		"code":            "2526",
		"status":          "planning",
		"is_active":       true,
		"start_date":      "2025-07-14",
		"end_date":        "2026-06-20",
		"semester1_start": "2025-07-14",
		"semester1_end":   "2025-12-19",
		"semester2_start": "2026-01-05",
		"semester2_end":   "2026-06-20",
	}

	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)
	assert.NotNil(t, entity)
	assert.NotEqual(t, uuid.Nil, entity.ID)
	assert.Equal(t, testScope.TenantID, entity.TenantID)
	assert.Equal(t, testScope.CompanyID, entity.CompanyID)
	assert.Equal(t, "2025/2026", entity.Data["name"])
	assert.Equal(t, "2526", entity.Data["code"])
	assert.Equal(t, "planning", entity.Data["status"])
	assert.Equal(t, vernon.SyncStatusSynced, entity.SyncStatus)
}

func TestVernonAcademicYear_Create_ValidationFail(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	data := map[string]any{
		"code":   "2526",
		"status": "planning",
	}

	_, err := ctx.svc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name wajib diisi")
}

func TestVernonAcademicYear_GetByID_Success(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	data := map[string]any{
		"name": "2026/2027", "code": "2627", "status": "planning",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	got, err := ctx.svc.GetByID(context.Background(), testScope, entity.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.ID, got.ID)
	assert.Equal(t, "2026/2027", got.Data["name"])
}

func TestVernonAcademicYear_GetByID_NotFound(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	_, err := ctx.svc.GetByID(context.Background(), testScope, uuid.New())
	assert.ErrorIs(t, err, vernon.ErrNotFound)
}

func TestVernonAcademicYear_Update_Success(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	data := map[string]any{
		"name": "2025/2026", "code": "2526", "status": "planning",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	updated := map[string]any{
		"name": "2025/2026 (Updated)", "code": "2526", "status": "active",
	}
	got, err := ctx.svc.Update(context.Background(), testScope, entity.ID, updated)
	require.NoError(t, err)
	assert.Equal(t, "active", got.Data["status"])
	assert.Equal(t, "2025/2026 (Updated)", got.Data["name"])
}

func TestVernonAcademicYear_Patch_Success(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	data := map[string]any{
		"name": "2025/2026", "code": "2526", "status": "planning",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	patch := map[string]any{"status": "active"}
	got, err := ctx.svc.Patch(context.Background(), testScope, entity.ID, patch)
	require.NoError(t, err)
	assert.Equal(t, "active", got.Data["status"])
	assert.Equal(t, "2526", got.Data["code"]) // unchanged
}

func TestVernonAcademicYear_Delete_Success(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	data := map[string]any{
		"name": "To Delete", "code": "DEL", "status": "planning",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	err = ctx.svc.Delete(context.Background(), testScope, entity.ID)
	require.NoError(t, err)

	_, err = ctx.svc.GetByID(context.Background(), testScope, entity.ID)
	assert.ErrorIs(t, err, vernon.ErrNotFound)
}

func TestVernonAcademicYear_List(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	for i := range 3 {
		data := map[string]any{
			"name": fmt.Sprintf("List Test %d", i), "code": fmt.Sprintf("LT%d", i),
			"status": "planning",
		}
		_, err := ctx.svc.Create(context.Background(), testScope, data)
		require.NoError(t, err)
	}

	items, total, err := ctx.svc.List(context.Background(), testScope, vernon.FindParams{
		Limit: 10, Page: 1,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(3))
	assert.GreaterOrEqual(t, len(items), 3)
}

func TestVernonAcademicYear_ScopeIsolation(t *testing.T) {
	ctx := setupVernonService(t, &academic_year.Descriptor{})

	data := map[string]any{
		"name": "Scoped Year", "code": "SCOPE", "status": "planning",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	otherScope := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	_, err = ctx.svc.GetByID(context.Background(), otherScope, entity.ID)
	assert.ErrorIs(t, err, vernon.ErrNotFound)
}

// ════════════════════════════════════════════════════════════════════════════
// TEACHERS
// ════════════════════════════════════════════════════════════════════════════

func TestVernonTeacher_Create_Success(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"full_name":      "Budi Santoso",
		"gender":         "L",
		"employee_type":  "pns",
		"role":           "guru_mapel",
		"nip":            "198501012010011001",
		"status":         "active",
		"religion":       "islam",
		"phone":          "081234567890",
	}

	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)
	assert.NotNil(t, entity)
	assert.NotEqual(t, uuid.Nil, entity.ID)
	assert.Equal(t, "Budi Santoso", entity.Data["full_name"])
	assert.Equal(t, "guru_mapel", entity.Data["role"])
}

func TestVernonTeacher_Create_ValidationFail_MissingFullName(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"gender": "L", "employee_type": "pns", "role": "guru_mapel",
	}
	_, err := ctx.svc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "full_name wajib diisi")
}

func TestVernonTeacher_Create_ValidationFail_InvalidGender(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"full_name": "Test", "gender": "X", "employee_type": "pns", "role": "guru_mapel",
	}
	_, err := ctx.svc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gender tidak valid")
}

func TestVernonTeacher_Create_ValidationFail_InvalidRole(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"full_name": "Test", "gender": "L", "employee_type": "pns", "role": "superhero",
	}
	_, err := ctx.svc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role tidak valid")
}

func TestVernonTeacher_GetByID_Success(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"full_name": "Siti Aminah", "gender": "P",
		"employee_type": "honorer", "role": "guru_bk",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	got, err := ctx.svc.GetByID(context.Background(), testScope, entity.ID)
	require.NoError(t, err)
	assert.Equal(t, "Siti Aminah", got.Data["full_name"])
	assert.Equal(t, "P", got.Data["gender"])
}

func TestVernonTeacher_Update_Success(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"full_name": "Old Name", "gender": "L",
		"employee_type": "pns", "role": "guru_mapel",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	updated := map[string]any{
		"full_name": "New Name", "gender": "L",
		"employee_type": "pns", "role": "kepala_sekolah",
	}
	got, err := ctx.svc.Update(context.Background(), testScope, entity.ID, updated)
	require.NoError(t, err)
	assert.Equal(t, "New Name", got.Data["full_name"])
	assert.Equal(t, "kepala_sekolah", got.Data["role"])
}

func TestVernonTeacher_Delete_Success(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"full_name": "ToDelete", "gender": "L",
		"employee_type": "kontrak", "role": "staff_umum",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	err = ctx.svc.Delete(context.Background(), testScope, entity.ID)
	require.NoError(t, err)

	_, err = ctx.svc.GetByID(context.Background(), testScope, entity.ID)
	assert.ErrorIs(t, err, vernon.ErrNotFound)
}

func TestVernonTeacher_ScopeIsolation(t *testing.T) {
	ctx := setupVernonService(t, &teacher.Descriptor{})

	data := map[string]any{
		"full_name": "Scoped Teacher", "gender": "L",
		"employee_type": "pns", "role": "guru_mapel",
	}
	entity, err := ctx.svc.Create(context.Background(), testScope, data)
	require.NoError(t, err)

	otherScope := scope.Scope{TenantID: uuid.New(), CompanyID: uuid.New()}
	_, err = ctx.svc.GetByID(context.Background(), otherScope, entity.ID)
	assert.ErrorIs(t, err, vernon.ErrNotFound)
}

// ════════════════════════════════════════════════════════════════════════════
// CLASS ROOMS
// ════════════════════════════════════════════════════════════════════════════

func TestVernonClassRoom_Create_Success(t *testing.T) {
	aySvc, _, crSvc := setupAllVernonServices(t)

	// Create prerequisite academic year
	ayData := map[string]any{
		"name": "2025/2026", "code": "2526", "status": "active",
	}
	ay, err := aySvc.Create(context.Background(), testScope, ayData)
	require.NoError(t, err)

	// Create class room with academic_year reference
	crData := map[string]any{
		"name":             "VII-A",
		"grade_level":      "7",
		"academic_year_id": ay.ID.String(),
		"capacity":         float64(36),
		"is_active":        true,
	}
	entity, err := crSvc.Create(context.Background(), testScope, crData)
	require.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "VII-A", entity.Data["name"])
	assert.Equal(t, "7", entity.Data["grade_level"])
}

func TestVernonClassRoom_Create_WithAutoload(t *testing.T) {
	aySvc, tSvc, crSvc := setupAllVernonServices(t)

	// Create prerequisite academic year
	ayData := map[string]any{
		"name": "2025/2026", "code": "2526", "status": "active",
	}
	ay, err := aySvc.Create(context.Background(), testScope, ayData)
	require.NoError(t, err)

	// Create prerequisite teacher
	tData := map[string]any{
		"full_name": "Pak Budi", "gender": "L",
		"employee_type": "pns", "role": "guru_mapel",
		"nip": "123456789", "nuptk": "987654321",
	}
	teacherEntity, err := tSvc.Create(context.Background(), testScope, tData)
	require.NoError(t, err)

	// Create class room with both relations
	crData := map[string]any{
		"name":                "X-IPA-1",
		"grade_level":         "10",
		"academic_year_id":    ay.ID.String(),
		"homeroom_teacher_id": teacherEntity.ID.String(),
		"capacity":            float64(36),
		"major":               "ipa",
	}
	entity, err := crSvc.Create(context.Background(), testScope, crData)
	require.NoError(t, err)

	// Verify autoloaded data is embedded in _data
	assert.NotNil(t, entity.Data["academic_year"])
	assert.NotNil(t, entity.Data["homeroom_teacher"])

	ayEmbedded, ok := entity.Data["academic_year"].(map[string]any)
	require.True(t, ok, "academic_year should be a map")
	assert.Equal(t, "2025/2026", ayEmbedded["name"])

	tEmbedded, ok := entity.Data["homeroom_teacher"].(map[string]any)
	require.True(t, ok, "homeroom_teacher should be a map")
	assert.Equal(t, "Pak Budi", tEmbedded["full_name"])
}

func TestVernonClassRoom_Create_ValidationFail_MissingName(t *testing.T) {
	_, _, crSvc := setupAllVernonServices(t)

	data := map[string]any{
		"grade_level": "7", "academic_year_id": uuid.New().String(),
	}
	_, err := crSvc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name wajib diisi")
}

func TestVernonClassRoom_Create_ValidationFail_InvalidGradeLevel(t *testing.T) {
	_, _, crSvc := setupAllVernonServices(t)

	data := map[string]any{
		"name": "Test Class", "grade_level": "13",
		"academic_year_id": uuid.New().String(),
	}
	_, err := crSvc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grade_level tidak valid")
}

func TestVernonClassRoom_Create_ValidationFail_MissingAcademicYearID(t *testing.T) {
	_, _, crSvc := setupAllVernonServices(t)

	data := map[string]any{
		"name": "No Year Class", "grade_level": "7",
	}
	_, err := crSvc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "academic_year_id wajib diisi")
}

func TestVernonClassRoom_Create_ValidationFail_ZeroCapacity(t *testing.T) {
	_, _, crSvc := setupAllVernonServices(t)

	data := map[string]any{
		"name": "Zero Cap", "grade_level": "7",
		"academic_year_id": uuid.New().String(),
		"capacity": float64(0),
	}
	_, err := crSvc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "capacity harus lebih dari 0")
}

func TestVernonClassRoom_Create_ValidationFail_InvalidMajor(t *testing.T) {
	_, _, crSvc := setupAllVernonServices(t)

	data := map[string]any{
		"name": "Bad Major Class", "grade_level": "10",
		"academic_year_id": uuid.New().String(),
		"major": "ekonomi",
	}
	_, err := crSvc.Create(context.Background(), testScope, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "major tidak valid")
}

func TestVernonClassRoom_GetByID_Success(t *testing.T) {
	aySvc, _, crSvc := setupAllVernonServices(t)

	ayData := map[string]any{"name": "AY", "code": "AY1", "status": "active"}
	ay, _ := aySvc.Create(context.Background(), testScope, ayData)

	crData := map[string]any{
		"name": "VIII-B", "grade_level": "8",
		"academic_year_id": ay.ID.String(), "capacity": float64(32),
	}
	entity, err := crSvc.Create(context.Background(), testScope, crData)
	require.NoError(t, err)

	got, err := crSvc.GetByID(context.Background(), testScope, entity.ID)
	require.NoError(t, err)
	assert.Equal(t, "VIII-B", got.Data["name"])
}

func TestVernonClassRoom_Delete_Success(t *testing.T) {
	aySvc, _, crSvc := setupAllVernonServices(t)

	ayData := map[string]any{"name": "AY Del", "code": "AYD", "status": "active"}
	ay, _ := aySvc.Create(context.Background(), testScope, ayData)

	crData := map[string]any{
		"name": "Del Class", "grade_level": "9",
		"academic_year_id": ay.ID.String(),
	}
	entity, err := crSvc.Create(context.Background(), testScope, crData)
	require.NoError(t, err)

	err = crSvc.Delete(context.Background(), testScope, entity.ID)
	require.NoError(t, err)

	_, err = crSvc.GetByID(context.Background(), testScope, entity.ID)
	assert.ErrorIs(t, err, vernon.ErrNotFound)
}

func TestVernonClassRoom_ScopeIsolation(t *testing.T) {
	aySvc, _, crSvc := setupAllVernonServices(t)

	ayData := map[string]any{"name": "AY ISO", "code": "AI", "status": "active"}
	ay, _ := aySvc.Create(context.Background(), testScope, ayData)

	crData := map[string]any{
		"name": "Scoped Class", "grade_level": "7",
		"academic_year_id": ay.ID.String(),
	}
	entity, err := crSvc.Create(context.Background(), testScope, crData)
	require.NoError(t, err)

	otherScope := scope.Scope{TenantID: uuid.New(), CompanyID: uuid.New()}
	_, err = crSvc.GetByID(context.Background(), otherScope, entity.ID)
	assert.ErrorIs(t, err, vernon.ErrNotFound)
}

func TestVernonClassRoom_List_WithFilter(t *testing.T) {
	aySvc, _, crSvc := setupAllVernonServices(t)

	ayData := map[string]any{"name": "AY Filter", "code": "AF", "status": "active"}
	ay, _ := aySvc.Create(context.Background(), testScope, ayData)

	for i := range 3 {
		crData := map[string]any{
			"name": fmt.Sprintf("Filter-%d", i), "grade_level": "7",
			"academic_year_id": ay.ID.String(),
		}
		_, err := crSvc.Create(context.Background(), testScope, crData)
		require.NoError(t, err)
	}

	items, total, err := crSvc.List(context.Background(), testScope, vernon.FindParams{
		Limit: 10, Page: 1,
		Filters: map[string]string{"name": "Filter"},
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(3))
	assert.GreaterOrEqual(t, len(items), 3)
}
