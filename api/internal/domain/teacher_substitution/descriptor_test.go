package teacher_substitution_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/teacher_substitution"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── DutySchedule: metadata tests ────────────────────────────────────────────

func TestDutyScheduleDescriptor_TableName(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	assert.Equal(t, "duty_schedules", d.TableName())
}

func TestDutyScheduleDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_substitution.DutyScheduleDescriptor{}
}

func TestDutyScheduleDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "duty_schedules should have 2 BelongsTo relations")
}

func TestDutyScheduleDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_substitution.DSRelTeacher]
	require.True(t, ok, "should have teacher relation")
	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_substitution.DSFieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

func TestDutyScheduleDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_substitution.DSRelAcademicYear]
	require.True(t, ok, "should have academic_year relation")
	assert.Equal(t, "academic_years", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

// ── DutySchedule: validation tests ──────────────────────────────────────────

func validDutyScheduleData() map[string]any {
	return map[string]any{
		teacher_substitution.DSFieldTeacherID:      "00000000-0000-0000-0000-000000000001",
		teacher_substitution.DSFieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		teacher_substitution.DSFieldSemester:       teacher_substitution.SemesterGanjil,
		teacher_substitution.DSFieldDayOfWeek:      float64(1),
		teacher_substitution.DSFieldDutyType:       teacher_substitution.DutyPiketPagi,
	}
}

func TestDutyScheduleDescriptor_Validate_Success(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	err := d.Validate(validDutyScheduleData())
	assert.NoError(t, err)
}

func TestDutyScheduleDescriptor_Validate_AllDutyTypes(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	types := []string{
		teacher_substitution.DutyPiketPagi, teacher_substitution.DutyPiketKelas,
		teacher_substitution.DutyPiketGerbang, teacher_substitution.DutyPiketUpacara,
		teacher_substitution.DutyPiketSiang, teacher_substitution.DutyPiketAsrama,
		teacher_substitution.DutyPiketMalam,
	}
	for _, dt := range types {
		data := validDutyScheduleData()
		data[teacher_substitution.DSFieldDutyType] = dt
		err := d.Validate(data)
		assert.NoError(t, err, "duty_type=%q should be valid", dt)
	}
}

func TestDutyScheduleDescriptor_Validate_AllSemesters(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	semesters := []string{
		teacher_substitution.SemesterGanjil, teacher_substitution.SemesterGenap,
	}
	for _, sem := range semesters {
		data := validDutyScheduleData()
		data[teacher_substitution.DSFieldSemester] = sem
		err := d.Validate(data)
		assert.NoError(t, err, "semester=%q should be valid", sem)
	}
}

func TestDutyScheduleDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	data := validDutyScheduleData()
	delete(data, teacher_substitution.DSFieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestDutyScheduleDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	data := validDutyScheduleData()
	data[teacher_substitution.DSFieldSemester] = "semester_3"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}

func TestDutyScheduleDescriptor_Validate_InvalidDayOfWeek(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	data := validDutyScheduleData()
	data[teacher_substitution.DSFieldDayOfWeek] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "day_of_week harus antara 1 dan 7")
}

func TestDutyScheduleDescriptor_Validate_InvalidDutyType(t *testing.T) {
	d := &teacher_substitution.DutyScheduleDescriptor{}
	data := validDutyScheduleData()
	data[teacher_substitution.DSFieldDutyType] = "piket_lab"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "duty_type tidak valid")
}

// ── TeacherSubstitution: metadata tests ─────────────────────────────────────

func TestTeacherSubstitutionDescriptor_TableName(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	assert.Equal(t, "teacher_substitutions", d.TableName())
}

func TestTeacherSubstitutionDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_substitution.TeacherSubstitutionDescriptor{}
}

func TestTeacherSubstitutionDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 4, "teacher_substitutions should have 4 BelongsTo relations")
}

func TestTeacherSubstitutionDescriptor_DefaultRels_OriginalTeacher(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_substitution.TSRelOriginalTeacher]
	require.True(t, ok, "should have original_teacher relation")
	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_substitution.TSFieldOriginalTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"full_name", "nip"}, rel.Fields)
}

func TestTeacherSubstitutionDescriptor_DefaultRels_SubstituteTeacher(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_substitution.TSRelSubstituteTeacher]
	require.True(t, ok, "should have substitute_teacher relation")
	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, teacher_substitution.TSFieldSubstituteTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

func TestTeacherSubstitutionDescriptor_DefaultRels_ClassRoom(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_substitution.TSRelClassRoom]
	require.True(t, ok, "should have class_room relation")
	assert.Equal(t, "class_rooms", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

// ── TeacherSubstitution: validation tests ───────────────────────────────────

func validSubstitutionData() map[string]any {
	return map[string]any{
		teacher_substitution.TSFieldOriginalTeacherID:  "00000000-0000-0000-0000-000000000001",
		teacher_substitution.TSFieldSubstituteTeacherID: "00000000-0000-0000-0000-000000000002",
		teacher_substitution.TSFieldAcademicYearID:     "00000000-0000-0000-0000-000000000003",
		teacher_substitution.TSFieldSubstitutionDate:   "2026-04-20",
		teacher_substitution.TSFieldReasonType:         teacher_substitution.ReasonSakit,
		teacher_substitution.TSFieldClassRoomID:        "00000000-0000-0000-0000-000000000004",
		teacher_substitution.TSFieldStatus:             teacher_substitution.SubStatusPending,
	}
}

func TestTeacherSubstitutionDescriptor_Validate_Success(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	err := d.Validate(validSubstitutionData())
	assert.NoError(t, err)
}

func TestTeacherSubstitutionDescriptor_Validate_AllReasonTypes(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	reasons := []string{
		teacher_substitution.ReasonSakit, teacher_substitution.ReasonCuti,
		teacher_substitution.ReasonIzin, teacher_substitution.ReasonDinasLuar,
		teacher_substitution.ReasonTugasBelajar, teacher_substitution.ReasonTerlambat,
		teacher_substitution.ReasonLainLain,
	}
	for _, reason := range reasons {
		data := validSubstitutionData()
		data[teacher_substitution.TSFieldReasonType] = reason
		err := d.Validate(data)
		assert.NoError(t, err, "reason_type=%q should be valid", reason)
	}
}

func TestTeacherSubstitutionDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	statuses := []string{
		teacher_substitution.SubStatusPending, teacher_substitution.SubStatusNotified,
		teacher_substitution.SubStatusAccepted, teacher_substitution.SubStatusDeclined,
		teacher_substitution.SubStatusInProgress, teacher_substitution.SubStatusCompleted,
		teacher_substitution.SubStatusCancelled,
	}
	for _, status := range statuses {
		data := validSubstitutionData()
		data[teacher_substitution.TSFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestTeacherSubstitutionDescriptor_Validate_SameTeacherRejected(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	data := validSubstitutionData()
	data[teacher_substitution.TSFieldSubstituteTeacherID] = "00000000-0000-0000-0000-000000000001"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tidak boleh sama")
}

func TestTeacherSubstitutionDescriptor_Validate_MissingOriginalTeacherID(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	data := validSubstitutionData()
	delete(data, teacher_substitution.TSFieldOriginalTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "original_teacher_id wajib diisi")
}

func TestTeacherSubstitutionDescriptor_Validate_InvalidReasonType(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	data := validSubstitutionData()
	data[teacher_substitution.TSFieldReasonType] = "malas"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "reason_type tidak valid")
}

func TestTeacherSubstitutionDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	data := validSubstitutionData()
	data[teacher_substitution.TSFieldStatus] = "waiting"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

// ── SubstitutionLog: metadata tests ─────────────────────────────────────────

func TestSubstitutionLogDescriptor_TableName(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	assert.Equal(t, "substitution_logs", d.TableName())
}

func TestSubstitutionLogDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_substitution.SubstitutionLogDescriptor{}
}

func TestSubstitutionLogDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "substitution_logs should have 2 BelongsTo relations")
}

func TestSubstitutionLogDescriptor_DefaultRels_Substitution(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_substitution.SLRelSubstitution]
	require.True(t, ok, "should have substitution relation")
	assert.Equal(t, "teacher_substitutions", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload)
}

func TestSubstitutionLogDescriptor_DefaultRels_Actor(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[teacher_substitution.SLRelActor]
	require.True(t, ok, "should have actor relation")
	assert.Equal(t, "teachers", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

// ── SubstitutionLog: validation tests ───────────────────────────────────────

func validSubstitutionLogData() map[string]any {
	return map[string]any{
		teacher_substitution.SLFieldSubstitutionID: "00000000-0000-0000-0000-000000000001",
		teacher_substitution.SLFieldActorID:        "00000000-0000-0000-0000-000000000002",
		teacher_substitution.SLFieldAction:         teacher_substitution.LogActionCreated,
	}
}

func TestSubstitutionLogDescriptor_Validate_Success(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	err := d.Validate(validSubstitutionLogData())
	assert.NoError(t, err)
}

func TestSubstitutionLogDescriptor_Validate_AllActions(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	actions := []string{
		teacher_substitution.LogActionCreated, teacher_substitution.LogActionNotified,
		teacher_substitution.LogActionAccepted, teacher_substitution.LogActionDeclined,
		teacher_substitution.LogActionReassigned, teacher_substitution.LogActionStarted,
		teacher_substitution.LogActionCompleted, teacher_substitution.LogActionCancelled,
	}
	for _, action := range actions {
		data := validSubstitutionLogData()
		data[teacher_substitution.SLFieldAction] = action
		err := d.Validate(data)
		assert.NoError(t, err, "action=%q should be valid", action)
	}
}

func TestSubstitutionLogDescriptor_Validate_MissingSubstitutionID(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	data := validSubstitutionLogData()
	delete(data, teacher_substitution.SLFieldSubstitutionID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "substitution_id wajib diisi")
}

func TestSubstitutionLogDescriptor_Validate_MissingAction(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	data := validSubstitutionLogData()
	delete(data, teacher_substitution.SLFieldAction)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "action wajib diisi")
}

func TestSubstitutionLogDescriptor_Validate_InvalidAction(t *testing.T) {
	d := &teacher_substitution.SubstitutionLogDescriptor{}
	data := validSubstitutionLogData()
	data[teacher_substitution.SLFieldAction] = "forwarded"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "action tidak valid")
}
