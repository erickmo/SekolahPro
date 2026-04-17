package teaching_journal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/teaching_journal"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── TeachingJournalDescriptor metadata tests ─────────────────────────────────

func TestTeachingJournalDescriptor_TableName(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	assert.Equal(t, "teaching_journals", d.TableName())
}

func TestTeachingJournalDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teaching_journal.TeachingJournalDescriptor{}
}

func TestTeachingJournalDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 4, "teaching_journals should have 4 BelongsTo relations")
}

func TestTeachingJournalDescriptor_DefaultRels_AllAutoload(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	rels := d.DefaultRels()

	expected := []string{
		teaching_journal.RelTeacherTJ, teaching_journal.RelSubjectTJ,
		teaching_journal.RelClassRoomTJ, teaching_journal.RelAcademicYearTJ,
	}
	for _, name := range expected {
		rel, ok := rels[name]
		require.True(t, ok, "should have %s relation", name)
		assert.Equal(t, vernon.RelBelongsTo, rel.Type, "%s should be belongs_to", name)
		assert.True(t, rel.IsAutoload, "%s should be autoloaded", name)
	}
}

// ── TeachingJournalDescriptor validation tests ───────────────────────────────

func validTeachingJournalData() map[string]any {
	return map[string]any{
		teaching_journal.FieldTeacherIDTJ:      "00000000-0000-0000-0000-000000000001",
		teaching_journal.FieldSubjectIDTJ:      "00000000-0000-0000-0000-000000000002",
		teaching_journal.FieldClassRoomIDTJ:    "00000000-0000-0000-0000-000000000003",
		teaching_journal.FieldAcademicYearIDTJ: "00000000-0000-0000-0000-000000000004",
		teaching_journal.FieldJournalDate:      "2026-01-15",
		teaching_journal.FieldSemesterTJ:       teaching_journal.SemesterGanjilTJ,
		teaching_journal.FieldSlotStart:        float64(1),
		teaching_journal.FieldSlotEnd:          float64(3),
		teaching_journal.FieldTopicTaught:      "Bilangan Bulat",
		teaching_journal.FieldStatusTJ:         teaching_journal.JStatusDraft,
		teaching_journal.FieldTotalPresent:     float64(25),
		teaching_journal.FieldTotalAbsent:      float64(2),
		teaching_journal.FieldTotalLate:        float64(1),
		teaching_journal.FieldTotalSick:        float64(0),
		teaching_journal.FieldTotalPermitted:   float64(0),
	}
}

func TestTeachingJournalDescriptor_Validate_Success(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	err := d.Validate(validTeachingJournalData())
	assert.NoError(t, err)
}

func TestTeachingJournalDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	statuses := []string{
		teaching_journal.JStatusDraft, teaching_journal.JStatusSubmitted,
		teaching_journal.JStatusVerified,
	}
	for _, status := range statuses {
		data := validTeachingJournalData()
		data[teaching_journal.FieldStatusTJ] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestTeachingJournalDescriptor_Validate_AllTeachingMethods(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	methods := []string{
		teaching_journal.MethodBandonganTJ, teaching_journal.MethodSoroganTJ,
		teaching_journal.MethodHalaqahTJ, teaching_journal.MethodMuhafadzahTJ,
		teaching_journal.MethodMudzakarahTJ, teaching_journal.MethodCeramahTJ,
		teaching_journal.MethodDemonstrasiTJ, teaching_journal.MethodTanyaJawabTJ,
	}
	for _, method := range methods {
		data := validTeachingJournalData()
		data[teaching_journal.FieldTeachingMethodPesantrenTJ] = method
		err := d.Validate(data)
		assert.NoError(t, err, "teaching_method_pesantren=%q should be valid", method)
	}
}

func TestTeachingJournalDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	delete(data, teaching_journal.FieldTeacherIDTJ)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestTeachingJournalDescriptor_Validate_MissingSubjectID(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	delete(data, teaching_journal.FieldSubjectIDTJ)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "subject_id wajib diisi")
}

func TestTeachingJournalDescriptor_Validate_MissingClassRoomID(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	delete(data, teaching_journal.FieldClassRoomIDTJ)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "class_room_id wajib diisi")
}

func TestTeachingJournalDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	delete(data, teaching_journal.FieldAcademicYearIDTJ)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestTeachingJournalDescriptor_Validate_MissingJournalDate(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	delete(data, teaching_journal.FieldJournalDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "journal_date wajib diisi")
}

func TestTeachingJournalDescriptor_Validate_MissingSemester(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	delete(data, teaching_journal.FieldSemesterTJ)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester wajib diisi")
}

func TestTeachingJournalDescriptor_Validate_MissingTopicTaught(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	delete(data, teaching_journal.FieldTopicTaught)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "topic_taught wajib diisi")
}

func TestTeachingJournalDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldSemesterTJ] = "summer"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}

func TestTeachingJournalDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldStatusTJ] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestTeachingJournalDescriptor_Validate_InvalidTeachingMethod(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldTeachingMethodPesantrenTJ] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teaching_method_pesantren tidak valid")
}

func TestTeachingJournalDescriptor_Validate_SlotStartOutOfRange(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldSlotStart] = float64(20)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "slot_start harus antara")
}

func TestTeachingJournalDescriptor_Validate_SlotEndOutOfRange(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldSlotEnd] = float64(20)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "slot_end harus antara")
}

func TestTeachingJournalDescriptor_Validate_SlotStartGreaterThanEnd(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldSlotStart] = float64(5)
	data[teaching_journal.FieldSlotEnd] = float64(3)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "slot_start tidak boleh lebih besar dari slot_end")
}

func TestTeachingJournalDescriptor_Validate_NegativeAttendanceCount(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldTotalPresent] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tidak boleh negatif")
}

func TestTeachingJournalDescriptor_Validate_NegativeTotalAbsent(t *testing.T) {
	d := &teaching_journal.TeachingJournalDescriptor{}
	data := validTeachingJournalData()
	data[teaching_journal.FieldTotalAbsent] = float64(-5)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tidak boleh negatif")
}

// ── JournalSessionAttendanceDescriptor metadata tests ────────────────────────

func TestJournalSessionAttendanceDescriptor_TableName(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	assert.Equal(t, "journal_session_attendances", d.TableName())
}

func TestJournalSessionAttendanceDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teaching_journal.JournalSessionAttendanceDescriptor{}
}

func TestJournalSessionAttendanceDescriptor_DefaultRels_Count(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "journal_session_attendances should have 2 BelongsTo relations")
}

func TestJournalSessionAttendanceDescriptor_DefaultRels_AllAutoload(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	rels := d.DefaultRels()

	expected := []string{teaching_journal.RelJournal, teaching_journal.RelStudentJSA}
	for _, name := range expected {
		rel, ok := rels[name]
		require.True(t, ok, "should have %s relation", name)
		assert.Equal(t, vernon.RelBelongsTo, rel.Type, "%s should be belongs_to", name)
		assert.True(t, rel.IsAutoload, "%s should be autoloaded", name)
	}
}

// ── JournalSessionAttendanceDescriptor validation tests ──────────────────────

func validJSAData() map[string]any {
	return map[string]any{
		teaching_journal.FieldJournalID:        "00000000-0000-0000-0000-000000000001",
		teaching_journal.FieldStudentIDJSA:     "00000000-0000-0000-0000-000000000002",
		teaching_journal.FieldAttendanceStatus: teaching_journal.AttendancePresent,
	}
}

func TestJournalSessionAttendanceDescriptor_Validate_Success(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	err := d.Validate(validJSAData())
	assert.NoError(t, err)
}

func TestJournalSessionAttendanceDescriptor_Validate_AllAttendanceStatuses(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	statuses := []string{
		teaching_journal.AttendancePresent, teaching_journal.AttendanceAbsent,
		teaching_journal.AttendanceLate, teaching_journal.AttendanceSick,
		teaching_journal.AttendancePermitted, teaching_journal.AttendanceDispensasi,
	}
	for _, status := range statuses {
		data := validJSAData()
		data[teaching_journal.FieldAttendanceStatus] = status
		if status == teaching_journal.AttendanceLate {
			data[teaching_journal.FieldLateMinutes] = float64(10)
		}
		err := d.Validate(data)
		assert.NoError(t, err, "attendance_status=%q should be valid", status)
	}
}

func TestJournalSessionAttendanceDescriptor_Validate_LateWithLateMinutes(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	data[teaching_journal.FieldAttendanceStatus] = teaching_journal.AttendanceLate
	data[teaching_journal.FieldLateMinutes] = float64(15)
	err := d.Validate(data)
	assert.NoError(t, err)
}

func TestJournalSessionAttendanceDescriptor_Validate_MissingJournalID(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	delete(data, teaching_journal.FieldJournalID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "journal_id wajib diisi")
}

func TestJournalSessionAttendanceDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	delete(data, teaching_journal.FieldStudentIDJSA)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestJournalSessionAttendanceDescriptor_Validate_MissingAttendanceStatus(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	delete(data, teaching_journal.FieldAttendanceStatus)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "attendance_status wajib diisi")
}

func TestJournalSessionAttendanceDescriptor_Validate_InvalidAttendanceStatus(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	data[teaching_journal.FieldAttendanceStatus] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "attendance_status tidak valid")
}

func TestJournalSessionAttendanceDescriptor_Validate_LateWithoutLateMinutes(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	data[teaching_journal.FieldAttendanceStatus] = teaching_journal.AttendanceLate
	err := d.Validate(data)
	assert.ErrorContains(t, err, "late_minutes wajib diisi jika attendance_status = late")
}

func TestJournalSessionAttendanceDescriptor_Validate_LateMinutesOutOfRange(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	data[teaching_journal.FieldAttendanceStatus] = teaching_journal.AttendanceLate
	data[teaching_journal.FieldLateMinutes] = float64(200)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "late_minutes harus antara")
}

func TestJournalSessionAttendanceDescriptor_Validate_LateMinutesZero(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	data[teaching_journal.FieldAttendanceStatus] = teaching_journal.AttendanceLate
	data[teaching_journal.FieldLateMinutes] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "late_minutes harus antara")
}

func TestJournalSessionAttendanceDescriptor_Validate_NonLateWithLateMinutesAllowed(t *testing.T) {
	d := &teaching_journal.JournalSessionAttendanceDescriptor{}
	data := validJSAData()
	data[teaching_journal.FieldAttendanceStatus] = teaching_journal.AttendancePresent
	data[teaching_journal.FieldLateMinutes] = float64(10)
	err := d.Validate(data)
	assert.NoError(t, err, "non-late status with late_minutes should be allowed")
}
