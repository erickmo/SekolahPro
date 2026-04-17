package teacher_workload_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/teacher_workload"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ──────────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	assert.Equal(t, "teacher_workloads", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_workload.Descriptor{}
}

// ── Descriptor DefaultRels tests ───────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "teacher_workloads should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_workload.RelTeacher]
	require.True(t, ok, "should have teacher relation")

	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_workload.FieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload, "teacher should be autoloaded")
	assert.Equal(t, []string{"full_name", "nip", "employee_type", "role"}, rel.Fields)
}

func TestDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_workload.RelAcademicYear]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_workload.FieldAcademicYearID, rel.FK)
	assert.True(t, rel.IsAutoload, "academic_year should be autoloaded")
	assert.Equal(t, []string{"name"}, rel.Fields)
}

// ── Descriptor Validation tests ────────────────────────────────────────────────

func validWorkloadData() map[string]any {
	return map[string]any{
		teacher_workload.FieldTeacherID:      "00000000-0000-0000-0000-000000000001",
		teacher_workload.FieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		teacher_workload.FieldSemester:       teacher_workload.SemesterGanjil,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	err := d.Validate(validWorkloadData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllSemesters(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	semesters := []string{teacher_workload.SemesterGanjil, teacher_workload.SemesterGenap}
	for _, sem := range semesters {
		data := validWorkloadData()
		data[teacher_workload.FieldSemester] = sem
		err := d.Validate(data)
		assert.NoError(t, err, "semester=%q should be valid", sem)
	}
}

func TestDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	data := validWorkloadData()
	delete(data, teacher_workload.FieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	data := validWorkloadData()
	delete(data, teacher_workload.FieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestDescriptor_Validate_MissingSemester(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	data := validWorkloadData()
	delete(data, teacher_workload.FieldSemester)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester wajib diisi")
}

func TestDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &teacher_workload.Descriptor{}
	data := validWorkloadData()
	data[teacher_workload.FieldSemester] = "midterm"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}

// ── ItemDescriptor metadata tests ──────────────────────────────────────────────

func TestItemDescriptor_TableName(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	assert.Equal(t, "teacher_workload_items", d.TableName())
}

func TestItemDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_workload.ItemDescriptor{}
}

// ── ItemDescriptor DefaultRels tests ───────────────────────────────────────────

func TestItemDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "teacher_workload_items should have 2 BelongsTo relations")
}

func TestItemDescriptor_DefaultRels_WorkloadRelation(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_workload.RelWorkload]
	require.True(t, ok, "should have workload relation")

	assert.Equal(t, "teacher_workloads", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_workload.FieldWorkloadID, rel.FK)
	assert.True(t, rel.IsAutoload, "workload should be autoloaded")
	assert.Equal(t, []string{"total_hours", "is_fulfilled"}, rel.Fields)
}

func TestItemDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_workload.RelTeacher]
	require.True(t, ok, "should have teacher relation")

	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_workload.FieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload, "teacher should be autoloaded")
	assert.Equal(t, []string{"full_name"}, rel.Fields)
}

// ── ItemDescriptor Validation tests ────────────────────────────────────────────

func validItemData() map[string]any {
	return map[string]any{
		teacher_workload.FieldWorkloadID:     "00000000-0000-0000-0000-000000000010",
		teacher_workload.FieldTeacherID:      "00000000-0000-0000-0000-000000000001",
		teacher_workload.FieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		teacher_workload.FieldItemType:       teacher_workload.ItemTypeMengajar,
		teacher_workload.FieldDescription:    "Mengajar Matematika Kelas X-A",
		teacher_workload.FieldHoursPerWeek:   float64(6),
	}
}

func TestItemDescriptor_Validate_Success(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	err := d.Validate(validItemData())
	assert.NoError(t, err)
}

func TestItemDescriptor_Validate_AllItemTypes(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	types := []string{
		teacher_workload.ItemTypeMengajar, teacher_workload.ItemTypeWaliKelas,
		teacher_workload.ItemTypePembinaEkskul, teacher_workload.ItemTypeGuruBK,
		teacher_workload.ItemTypeKepalaSekolah, teacher_workload.ItemTypeWakilKepsek,
		teacher_workload.ItemTypeKoordinator, teacher_workload.ItemTypePanitia,
		teacher_workload.ItemTypeTugasTambahan,
	}
	for _, it := range types {
		data := validItemData()
		data[teacher_workload.FieldItemType] = it
		err := d.Validate(data)
		assert.NoError(t, err, "item_type=%q should be valid", it)
	}
}

func TestItemDescriptor_Validate_MissingWorkloadID(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	data := validItemData()
	delete(data, teacher_workload.FieldWorkloadID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "workload_id wajib diisi")
}

func TestItemDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	data := validItemData()
	delete(data, teacher_workload.FieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestItemDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	data := validItemData()
	delete(data, teacher_workload.FieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestItemDescriptor_Validate_MissingItemType(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	data := validItemData()
	delete(data, teacher_workload.FieldItemType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "item_type wajib diisi")
}

func TestItemDescriptor_Validate_InvalidItemType(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	data := validItemData()
	data[teacher_workload.FieldItemType] = "freelance"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "item_type tidak valid")
}

func TestItemDescriptor_Validate_MissingDescription(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	data := validItemData()
	delete(data, teacher_workload.FieldDescription)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "description wajib diisi")
}

func TestItemDescriptor_Validate_MissingHoursPerWeek(t *testing.T) {
	d := &teacher_workload.ItemDescriptor{}
	data := validItemData()
	delete(data, teacher_workload.FieldHoursPerWeek)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "hours_per_week wajib diisi")
}
