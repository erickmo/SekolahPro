package student_counseling_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/student_counseling"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── CaseDescriptor metadata tests ──────────────────────────────────────────────

func TestCaseDescriptor_TableName(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	assert.Equal(t, "counseling_cases", d.TableName())
}

func TestCaseDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &student_counseling.CaseDescriptor{}
}

// ── CaseDescriptor DefaultRels tests ───────────────────────────────────────────

func TestCaseDescriptor_DefaultRels_Count(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "counseling_cases should have 2 BelongsTo relations")
}

func TestCaseDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_counseling.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_counseling.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestCaseDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_counseling.RelAcademicYear]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_counseling.FieldAcademicYearID, rel.FK)
	assert.True(t, rel.IsAutoload, "academic_year should be autoloaded")
	assert.Equal(t, []string{"name", "code"}, rel.Fields)
}

// ── CaseDescriptor Validation tests ────────────────────────────────────────────

func validCaseData() map[string]any {
	return map[string]any{
		student_counseling.FieldStudentID:      "00000000-0000-0000-0000-000000000001",
		student_counseling.FieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		student_counseling.FieldCaseNo:         "BK-2026-001",
		student_counseling.FieldCategory:       student_counseling.CategoryAcademic,
		student_counseling.FieldTitle:          "Kesulitan Fokus Belajar",
		student_counseling.FieldSeverity:       student_counseling.SeverityMedium,
		student_counseling.FieldOpenedDate:     "2026-04-17",
		student_counseling.FieldCounselorID:    "00000000-0000-0000-0000-000000000003",
	}
}

func TestCaseDescriptor_Validate_Success(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	err := d.Validate(validCaseData())
	assert.NoError(t, err)
}

func TestCaseDescriptor_Validate_AllCategories(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	categories := []string{
		student_counseling.CategoryAcademic, student_counseling.CategorySocial,
		student_counseling.CategoryPersonal, student_counseling.CategoryCareer,
		student_counseling.CategoryBehavioral, student_counseling.CategoryFamily,
		student_counseling.CategoryOther,
	}
	for _, cat := range categories {
		data := validCaseData()
		data[student_counseling.FieldCategory] = cat
		err := d.Validate(data)
		assert.NoError(t, err, "category=%q should be valid", cat)
	}
}

func TestCaseDescriptor_Validate_AllSeverities(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	severities := []string{
		student_counseling.SeverityLow, student_counseling.SeverityMedium,
		student_counseling.SeverityHigh, student_counseling.SeverityCritical,
	}
	for _, sev := range severities {
		data := validCaseData()
		data[student_counseling.FieldSeverity] = sev
		err := d.Validate(data)
		assert.NoError(t, err, "severity=%q should be valid", sev)
	}
}

func TestCaseDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestCaseDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestCaseDescriptor_Validate_MissingCaseNo(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldCaseNo)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "case_no wajib diisi")
}

func TestCaseDescriptor_Validate_MissingCategory(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldCategory)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "category wajib diisi")
}

func TestCaseDescriptor_Validate_InvalidCategory(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	data[student_counseling.FieldCategory] = "health"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "category tidak valid")
}

func TestCaseDescriptor_Validate_MissingTitle(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldTitle)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "title wajib diisi")
}

func TestCaseDescriptor_Validate_MissingSeverity(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldSeverity)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "severity wajib diisi")
}

func TestCaseDescriptor_Validate_InvalidSeverity(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	data[student_counseling.FieldSeverity] = "urgent"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "severity tidak valid")
}

func TestCaseDescriptor_Validate_MissingOpenedDate(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldOpenedDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "opened_date wajib diisi")
}

func TestCaseDescriptor_Validate_MissingCounselorID(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	data := validCaseData()
	delete(data, student_counseling.FieldCounselorID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "counselor_id wajib diisi")
}

// ── SessionDescriptor metadata tests ───────────────────────────────────────────

func TestSessionDescriptor_TableName(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	assert.Equal(t, "counseling_sessions", d.TableName())
}

func TestSessionDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &student_counseling.SessionDescriptor{}
}

// ── SessionDescriptor DefaultRels tests ────────────────────────────────────────

func TestSessionDescriptor_DefaultRels_Count(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "counseling_sessions should have 2 BelongsTo relations")
}

func TestSessionDescriptor_DefaultRels_CaseRelation(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_counseling.RelCase]
	require.True(t, ok, "should have case relation")

	assert.Equal(t, "counseling_cases", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_counseling.FieldCaseID, rel.FK)
	assert.True(t, rel.IsAutoload, "case should be autoloaded")
	assert.Equal(t, []string{"case_no", "title", "status"}, rel.Fields)
}

func TestSessionDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_counseling.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_counseling.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
}

// ── SessionDescriptor Validation tests ─────────────────────────────────────────

func validSessionData() map[string]any {
	return map[string]any{
		student_counseling.FieldCaseID:      "00000000-0000-0000-0000-000000000010",
		student_counseling.FieldStudentID:   "00000000-0000-0000-0000-000000000001",
		student_counseling.FieldSessionDate: "2026-04-17",
		student_counseling.FieldSessionType: student_counseling.SessionTypeIndividual,
		student_counseling.FieldNotes:       "Siswa bercerita tentang masalah keluarga",
		student_counseling.FieldCounselorID: "00000000-0000-0000-0000-000000000003",
	}
}

func TestSessionDescriptor_Validate_Success(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	err := d.Validate(validSessionData())
	assert.NoError(t, err)
}

func TestSessionDescriptor_Validate_AllSessionTypes(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	types := []string{
		student_counseling.SessionTypeIndividual, student_counseling.SessionTypeGroup,
		student_counseling.SessionTypeHomeVisit, student_counseling.SessionTypeParentConf,
		student_counseling.SessionTypeReferral,
	}
	for _, st := range types {
		data := validSessionData()
		data[student_counseling.FieldSessionType] = st
		err := d.Validate(data)
		assert.NoError(t, err, "session_type=%q should be valid", st)
	}
}

func TestSessionDescriptor_Validate_MissingCaseID(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	data := validSessionData()
	delete(data, student_counseling.FieldCaseID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "case_id wajib diisi")
}

func TestSessionDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	data := validSessionData()
	delete(data, student_counseling.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestSessionDescriptor_Validate_MissingSessionDate(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	data := validSessionData()
	delete(data, student_counseling.FieldSessionDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "session_date wajib diisi")
}

func TestSessionDescriptor_Validate_MissingSessionType(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	data := validSessionData()
	delete(data, student_counseling.FieldSessionType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "session_type wajib diisi")
}

func TestSessionDescriptor_Validate_InvalidSessionType(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	data := validSessionData()
	data[student_counseling.FieldSessionType] = "phone_call"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "session_type tidak valid")
}

func TestSessionDescriptor_Validate_MissingNotes(t *testing.T) {
	d := &student_counseling.SessionDescriptor{}
	data := validSessionData()
	delete(data, student_counseling.FieldNotes)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "notes wajib diisi")
}
