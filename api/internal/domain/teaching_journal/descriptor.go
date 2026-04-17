// Package teaching_journal adalah domain Vernon untuk jurnal mengajar harian guru.
//
// Teaching journal terdiri dari dua tabel:
//   - teaching_journals: catatan kegiatan belajar mengajar per sesi
//   - journal_session_attendances: kehadiran siswa per sesi jurnal
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teaching_journal

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field name constants ─────────────────────────────────────────────────────

// TeachingJournal field names.
const (
	FieldTeacherIDTJ      = "teacher_id"
	FieldSubjectIDTJ      = "subject_id"
	FieldClassRoomIDTJ    = "class_room_id"
	FieldAcademicYearIDTJ = "academic_year_id"
	FieldJournalDate      = "journal_date"
	FieldSemesterTJ       = "semester"
	FieldSlotStart        = "slot_start"
	FieldSlotEnd          = "slot_end"
	FieldTopicTaught      = "topic_taught"
	FieldStatusTJ         = "status"
	FieldTotalPresent     = "total_present"
	FieldTotalAbsent      = "total_absent"
	FieldTotalLate        = "total_late"
	FieldTotalSick        = "total_sick"
	FieldTotalPermitted   = "total_permitted"
	FieldTeachingMethodPesantrenTJ = "teaching_method_pesantren"
)

// JournalSessionAttendance field names.
const (
	FieldJournalID         = "journal_id"
	FieldStudentIDJSA      = "student_id"
	FieldAttendanceStatus  = "attendance_status"
	FieldLateMinutes       = "late_minutes"
)

// ── Enum constants ───────────────────────────────────────────────────────────

// Status constants.
const (
	JStatusDraft     = "draft"
	JStatusSubmitted = "submitted"
	JStatusVerified  = "verified"
)

// Attendance status constants.
const (
	AttendancePresent    = "present"
	AttendanceAbsent     = "absent"
	AttendanceLate       = "late"
	AttendanceSick       = "sick"
	AttendancePermitted  = "permitted"
	AttendanceDispensasi = "dispensasi"
)

// Teaching method pesantren constants.
const (
	MethodBandonganTJ  = "bandongan"
	MethodSoroganTJ    = "sorogan"
	MethodHalaqahTJ    = "halaqah"
	MethodMuhafadzahTJ = "muhafadzah"
	MethodMudzakarahTJ = "mudzakarah"
	MethodCeramahTJ    = "ceramah"
	MethodDemonstrasiTJ = "demonstrasi"
	MethodTanyaJawabTJ = "tanya_jawab"
)

// Semester constants.
const (
	SemesterGanjilTJ = "ganjil"
	SemesterGenapTJ  = "genap"
)

// ── Range constants ──────────────────────────────────────────────────────────

const (
	SlotRangeMin     = 1
	SlotRangeMax     = 15
	LateMinutesMin   = 1
	LateMinutesMax   = 120
)

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelTeacherTJ      = "teacher"
	RelSubjectTJ      = "subject"
	RelClassRoomTJ    = "class_room"
	RelAcademicYearTJ = "academic_year"
	RelJournal        = "journal"
	RelStudentJSA     = "student"
)

// ── Validation maps ──────────────────────────────────────────────────────────

var validJournalStatuses = map[string]bool{
	JStatusDraft: true, JStatusSubmitted: true, JStatusVerified: true,
}

var validAttendanceStatuses = map[string]bool{
	AttendancePresent: true, AttendanceAbsent: true, AttendanceLate: true,
	AttendanceSick: true, AttendancePermitted: true, AttendanceDispensasi: true,
}

var validTeachingMethodsTJ = map[string]bool{
	MethodBandonganTJ: true, MethodSoroganTJ: true, MethodHalaqahTJ: true,
	MethodMuhafadzahTJ: true, MethodMudzakarahTJ: true, MethodCeramahTJ: true,
	MethodDemonstrasiTJ: true, MethodTanyaJawabTJ: true,
}

var validSemestersTJ = map[string]bool{
	SemesterGanjilTJ: true, SemesterGenapTJ: true,
}

// ── TeachingJournalDescriptor ────────────────────────────────────────────────

// TeachingJournalDescriptor mengimplementasi vernon.DomainDescriptor untuk teaching_journals.
type TeachingJournalDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *TeachingJournalDescriptor) TableName() string { return "teaching_journals" }

// DefaultRels mendefinisikan relasi domain teaching_journals.
// 4 BelongsTo autoload: teacher, subject, class_room, academic_year.
func (d *TeachingJournalDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacherTJ: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTeacherIDTJ,
			LocalKey:   FieldTeacherIDTJ,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
		RelSubjectTJ: {
			Domain:     "subjects",
			Type:       vernon.RelBelongsTo,
			FK:         FieldSubjectIDTJ,
			LocalKey:   FieldSubjectIDTJ,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
		RelClassRoomTJ: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         FieldClassRoomIDTJ,
			LocalKey:   FieldClassRoomIDTJ,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "grade_level"},
		},
		RelAcademicYearTJ: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearIDTJ,
			LocalKey:   FieldAcademicYearIDTJ,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant domain teaching_journals sebelum write.
func (d *TeachingJournalDescriptor) Validate(data map[string]any) error {
	if err := validateTJRequired(data); err != nil {
		return err
	}
	return validateTJEnums(data)
}

// validateTJRequired memeriksa field wajib teaching_journals.
func validateTJRequired(data map[string]any) error {
	teacherID, _ := data[FieldTeacherIDTJ].(string)
	if teacherID == "" {
		return errors.New("teacher_id wajib diisi")
	}

	subjectID, _ := data[FieldSubjectIDTJ].(string)
	if subjectID == "" {
		return errors.New("subject_id wajib diisi")
	}

	classRoomID, _ := data[FieldClassRoomIDTJ].(string)
	if classRoomID == "" {
		return errors.New("class_room_id wajib diisi")
	}

	academicYearID, _ := data[FieldAcademicYearIDTJ].(string)
	if academicYearID == "" {
		return errors.New("academic_year_id wajib diisi")
	}

	if _, ok := data[FieldJournalDate]; !ok {
		return errors.New("journal_date wajib diisi")
	}

	semester, _ := data[FieldSemesterTJ].(string)
	if semester == "" {
		return errors.New("semester wajib diisi")
	}

	topicTaught, _ := data[FieldTopicTaught].(string)
	if topicTaught == "" {
		return errors.New("topic_taught wajib diisi")
	}
	return nil
}

// validateTJEnums memeriksa enum dan range constraints teaching_journals.
func validateTJEnums(data map[string]any) error {
	semester, _ := data[FieldSemesterTJ].(string)
	if semester != "" && !validSemestersTJ[semester] {
		return fmt.Errorf("semester tidak valid: %q (harus ganjil/genap)", semester)
	}

	status, _ := data[FieldStatusTJ].(string)
	if status != "" && !validJournalStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus draft/submitted/verified)", status)
	}

	method, _ := data[FieldTeachingMethodPesantrenTJ].(string)
	if method != "" && !validTeachingMethodsTJ[method] {
		return fmt.Errorf("teaching_method_pesantren tidak valid: %q", method)
	}

	if err := validateTJSlotRange(data); err != nil {
		return err
	}

	return validateTJAttendanceCounts(data)
}

// validateTJSlotRange memeriksa slot_start, slot_end range dan start <= end.
func validateTJSlotRange(data map[string]any) error {
	slotStart, startOK := data[FieldSlotStart].(float64)
	slotEnd, endOK := data[FieldSlotEnd].(float64)

	if startOK && (slotStart < SlotRangeMin || slotStart > SlotRangeMax) {
		return fmt.Errorf("slot_start harus antara %d dan %d", SlotRangeMin, SlotRangeMax)
	}
	if endOK && (slotEnd < SlotRangeMin || slotEnd > SlotRangeMax) {
		return fmt.Errorf("slot_end harus antara %d dan %d", SlotRangeMin, SlotRangeMax)
	}
	if startOK && endOK && slotStart > slotEnd {
		return errors.New("slot_start tidak boleh lebih besar dari slot_end")
	}
	return nil
}

// validateTJAttendanceCounts memeriksa semua attendance count >= 0.
func validateTJAttendanceCounts(data map[string]any) error {
	countFields := []string{
		FieldTotalPresent, FieldTotalAbsent, FieldTotalLate,
		FieldTotalSick, FieldTotalPermitted,
	}
	for _, field := range countFields {
		count, ok := data[field].(float64)
		if ok && count < 0 {
			return fmt.Errorf("%s tidak boleh negatif", field)
		}
	}
	return nil
}

// ── JournalSessionAttendanceDescriptor ───────────────────────────────────────

// JournalSessionAttendanceDescriptor mengimplementasi vernon.DomainDescriptor untuk journal_session_attendances.
type JournalSessionAttendanceDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *JournalSessionAttendanceDescriptor) TableName() string { return "journal_session_attendances" }

// DefaultRels mendefinisikan relasi domain journal_session_attendances.
// 2 BelongsTo autoload: journal, student.
func (d *JournalSessionAttendanceDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelJournal: {
			Domain:     "teaching_journals",
			Type:       vernon.RelBelongsTo,
			FK:         FieldJournalID,
			LocalKey:   FieldJournalID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"journal_date", "topic_taught", "status"},
		},
		RelStudentJSA: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         FieldStudentIDJSA,
			LocalKey:   FieldStudentIDJSA,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
	}
}

// Validate memvalidasi invariant domain journal_session_attendances sebelum write.
func (d *JournalSessionAttendanceDescriptor) Validate(data map[string]any) error {
	if err := validateJSARequired(data); err != nil {
		return err
	}
	return validateJSAEnums(data)
}

// validateJSARequired memeriksa field wajib journal_session_attendances.
func validateJSARequired(data map[string]any) error {
	journalID, _ := data[FieldJournalID].(string)
	if journalID == "" {
		return errors.New("journal_id wajib diisi")
	}

	studentID, _ := data[FieldStudentIDJSA].(string)
	if studentID == "" {
		return errors.New("student_id wajib diisi")
	}

	attendanceStatus, _ := data[FieldAttendanceStatus].(string)
	if attendanceStatus == "" {
		return errors.New("attendance_status wajib diisi")
	}
	return nil
}

// validateJSAEnums memeriksa enum dan range constraints journal_session_attendances.
func validateJSAEnums(data map[string]any) error {
	attendanceStatus, _ := data[FieldAttendanceStatus].(string)
	if attendanceStatus != "" && !validAttendanceStatuses[attendanceStatus] {
		return fmt.Errorf("attendance_status tidak valid: %q", attendanceStatus)
	}

	return validateJSALateMinutes(data)
}

// validateJSALateMinutes memeriksa late_minutes: 1-120, hanya jika status=late.
func validateJSALateMinutes(data map[string]any) error {
	attendanceStatus, _ := data[FieldAttendanceStatus].(string)
	lateMinutes, hasLateMinutes := data[FieldLateMinutes].(float64)

	if hasLateMinutes {
		if lateMinutes < LateMinutesMin || lateMinutes > LateMinutesMax {
			return fmt.Errorf("late_minutes harus antara %d dan %d", LateMinutesMin, LateMinutesMax)
		}
	}

	if attendanceStatus == AttendanceLate && !hasLateMinutes {
		return errors.New("late_minutes wajib diisi jika attendance_status = late")
	}
	return nil
}
