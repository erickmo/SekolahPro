// Package teacher_substitution adalah domain Vernon untuk jadwal piket dan penggantian guru.
//
// Terdiri dari 3 deskriptor:
//   - DutyScheduleDescriptor        : jadwal piket harian (2 BelongsTo)
//   - TeacherSubstitutionDescriptor  : penggantian guru mengajar (multiple BelongsTo)
//   - SubstitutionLogDescriptor      : log aktivitas substitusi (2 BelongsTo)
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teacher_substitution

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Duty Schedule ───────────────────────────────────────────────────────────

// Field name constants — duty_schedules.
const (
	DSFieldTeacherID      = "teacher_id"
	DSFieldAcademicYearID = "academic_year_id"
	DSFieldSemester       = "semester"
	DSFieldDayOfWeek      = "day_of_week"
	DSFieldDutyType       = "duty_type"
	DSFieldStartTime      = "start_time"
	DSFieldEndTime        = "end_time"
	DSFieldNotes          = "notes"
)

// Duty type enum constants.
const (
	DutyPiketPagi   = "piket_pagi"
	DutyPiketKelas  = "piket_kelas"
	DutyPiketGerbang = "piket_gerbang"
	DutyPiketUpacara = "piket_upacara"
	DutyPiketSiang  = "piket_siang"
	DutyPiketAsrama = "piket_asrama"
	DutyPiketMalam  = "piket_malam"
)

// Semester enum constants.
const (
	SemesterGanjil = "ganjil"
	SemesterGenap  = "genap"
)

// Relation name constants — duty_schedules.
const (
	DSRelTeacher      = "teacher"
	DSRelAcademicYear = "academic_year"
)

var validDutyTypes = map[string]bool{
	DutyPiketPagi: true, DutyPiketKelas: true, DutyPiketGerbang: true,
	DutyPiketUpacara: true, DutyPiketSiang: true, DutyPiketAsrama: true,
	DutyPiketMalam: true,
}

var validSemesters = map[string]bool{
	SemesterGanjil: true, SemesterGenap: true,
}

// DutyScheduleDescriptor mengimplementasi vernon.DomainDescriptor untuk duty_schedules.
type DutyScheduleDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *DutyScheduleDescriptor) TableName() string { return "duty_schedules" }

// DefaultRels mendefinisikan relasi domain ini.
// 2 BelongsTo autoload: teacher, academic_year.
func (d *DutyScheduleDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		DSRelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         DSFieldTeacherID,
			LocalKey:   DSFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type", "role"},
		},
		DSRelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         DSFieldAcademicYearID,
			LocalKey:   DSFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant duty_schedules sebelum write.
func (d *DutyScheduleDescriptor) Validate(data map[string]any) error {
	if err := validateRequiredString(data, DSFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, DSFieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}

	semester, _ := data[DSFieldSemester].(string)
	if semester == "" {
		return fmt.Errorf("semester wajib diisi")
	}
	if !validSemesters[semester] {
		return fmt.Errorf("semester tidak valid: %q", semester)
	}

	if err := validateDayOfWeek(data); err != nil {
		return err
	}

	dutyType, _ := data[DSFieldDutyType].(string)
	if dutyType == "" {
		return fmt.Errorf("duty_type wajib diisi")
	}
	if !validDutyTypes[dutyType] {
		return fmt.Errorf("duty_type tidak valid: %q", dutyType)
	}
	return nil
}

func validateDayOfWeek(data map[string]any) error {
	dayVal, ok := data[DSFieldDayOfWeek]
	if !ok || dayVal == nil {
		return fmt.Errorf("day_of_week wajib diisi")
	}
	dayFloat, ok := dayVal.(float64)
	if !ok || dayFloat < 1 || dayFloat > 7 {
		return fmt.Errorf("day_of_week harus antara 1 dan 7")
	}
	return nil
}

// ── Teacher Substitution ────────────────────────────────────────────────────

// Field name constants — teacher_substitutions.
const (
	TSFieldOriginalTeacherID  = "original_teacher_id"
	TSFieldSubstituteTeacherID = "substitute_teacher_id"
	TSFieldAcademicYearID     = "academic_year_id"
	TSFieldSubstitutionDate   = "substitution_date"
	TSFieldReasonType         = "reason_type"
	TSFieldClassRoomID        = "class_room_id"
	TSFieldStatus             = "status"
	TSFieldScheduleEntryID    = "schedule_entry_id"
	TSFieldTimeSlotID         = "time_slot_id"
	TSFieldOriginalSubjectID  = "original_subject_id"
	TSFieldNotes              = "notes"
)

// Reason type enum constants.
const (
	ReasonSakit        = "sakit"
	ReasonCuti         = "cuti"
	ReasonIzin         = "izin"
	ReasonDinasLuar    = "dinas_luar"
	ReasonTugasBelajar = "tugas_belajar"
	ReasonTerlambat    = "terlambat"
	ReasonLainLain     = "lain_lain"
)

// Status enum constants — teacher_substitutions.
const (
	SubStatusPending    = "pending"
	SubStatusNotified   = "notified"
	SubStatusAccepted   = "accepted"
	SubStatusDeclined   = "declined"
	SubStatusInProgress = "in_progress"
	SubStatusCompleted  = "completed"
	SubStatusCancelled  = "cancelled"
)

// Relation name constants — teacher_substitutions.
const (
	TSRelOriginalTeacher  = "original_teacher"
	TSRelSubstituteTeacher = "substitute_teacher"
	TSRelAcademicYear     = "academic_year"
	TSRelClassRoom        = "class_room"
)

var validReasonTypes = map[string]bool{
	ReasonSakit: true, ReasonCuti: true, ReasonIzin: true,
	ReasonDinasLuar: true, ReasonTugasBelajar: true,
	ReasonTerlambat: true, ReasonLainLain: true,
}

var validSubStatuses = map[string]bool{
	SubStatusPending: true, SubStatusNotified: true, SubStatusAccepted: true,
	SubStatusDeclined: true, SubStatusInProgress: true,
	SubStatusCompleted: true, SubStatusCancelled: true,
}

// TeacherSubstitutionDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_substitutions.
type TeacherSubstitutionDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *TeacherSubstitutionDescriptor) TableName() string { return "teacher_substitutions" }

// DefaultRels mendefinisikan relasi domain ini.
// Multiple BelongsTo: original_teacher (teachers), substitute_teacher (teachers),
// academic_year, class_room.
func (d *TeacherSubstitutionDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		TSRelOriginalTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         TSFieldOriginalTeacherID,
			LocalKey:   TSFieldOriginalTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
		TSRelSubstituteTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         TSFieldSubstituteTeacherID,
			LocalKey:   TSFieldSubstituteTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
		TSRelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         TSFieldAcademicYearID,
			LocalKey:   TSFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
		TSRelClassRoom: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         TSFieldClassRoomID,
			LocalKey:   TSFieldClassRoomID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant teacher_substitutions sebelum write.
func (d *TeacherSubstitutionDescriptor) Validate(data map[string]any) error {
	if err := validateRequiredString(data, TSFieldOriginalTeacherID, "original_teacher_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, TSFieldSubstituteTeacherID, "substitute_teacher_id"); err != nil {
		return err
	}
	if err := validateDifferentTeachers(data); err != nil {
		return err
	}
	return validateSubstitutionFields(data)
}

func validateDifferentTeachers(data map[string]any) error {
	original, _ := data[TSFieldOriginalTeacherID].(string)
	substitute, _ := data[TSFieldSubstituteTeacherID].(string)
	if original != "" && substitute != "" && original == substitute {
		return fmt.Errorf("original_teacher_id dan substitute_teacher_id tidak boleh sama")
	}
	return nil
}

func validateSubstitutionFields(data map[string]any) error {
	if err := validateRequiredString(data, TSFieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, TSFieldSubstitutionDate, "substitution_date"); err != nil {
		return err
	}

	reasonType, _ := data[TSFieldReasonType].(string)
	if reasonType == "" {
		return fmt.Errorf("reason_type wajib diisi")
	}
	if !validReasonTypes[reasonType] {
		return fmt.Errorf("reason_type tidak valid: %q", reasonType)
	}

	if err := validateRequiredString(data, TSFieldClassRoomID, "class_room_id"); err != nil {
		return err
	}

	status, _ := data[TSFieldStatus].(string)
	if status != "" && !validSubStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}

// ── Substitution Log ────────────────────────────────────────────────────────

// Field name constants — substitution_logs.
const (
	SLFieldSubstitutionID = "substitution_id"
	SLFieldActorID        = "actor_id"
	SLFieldAction         = "action"
	SLFieldNotes          = "notes"
)

// Action enum constants — substitution_logs.
const (
	LogActionCreated    = "created"
	LogActionNotified   = "notified"
	LogActionAccepted   = "accepted"
	LogActionDeclined   = "declined"
	LogActionReassigned = "reassigned"
	LogActionStarted    = "started"
	LogActionCompleted  = "completed"
	LogActionCancelled  = "cancelled"
)

// Relation name constants — substitution_logs.
const (
	SLRelSubstitution = "substitution"
	SLRelActor        = "actor"
)

var validLogActions = map[string]bool{
	LogActionCreated: true, LogActionNotified: true, LogActionAccepted: true,
	LogActionDeclined: true, LogActionReassigned: true, LogActionStarted: true,
	LogActionCompleted: true, LogActionCancelled: true,
}

// SubstitutionLogDescriptor mengimplementasi vernon.DomainDescriptor untuk substitution_logs.
type SubstitutionLogDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *SubstitutionLogDescriptor) TableName() string { return "substitution_logs" }

// DefaultRels mendefinisikan relasi domain ini.
// 2 BelongsTo autoload: substitution, actor.
func (d *SubstitutionLogDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		SLRelSubstitution: {
			Domain:     "teacher_substitutions",
			Type:       vernon.RelBelongsTo,
			FK:         SLFieldSubstitutionID,
			LocalKey:   SLFieldSubstitutionID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"original_teacher_name", "substitute_teacher_name", "class_name", "subject_name"},
		},
		SLRelActor: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         SLFieldActorID,
			LocalKey:   SLFieldActorID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "role"},
		},
	}
}

// Validate memvalidasi invariant substitution_logs sebelum write.
func (d *SubstitutionLogDescriptor) Validate(data map[string]any) error {
	if err := validateRequiredString(data, SLFieldSubstitutionID, "substitution_id"); err != nil {
		return err
	}
	if err := validateRequiredString(data, SLFieldActorID, "actor_id"); err != nil {
		return err
	}

	action, _ := data[SLFieldAction].(string)
	if action == "" {
		return fmt.Errorf("action wajib diisi")
	}
	if !validLogActions[action] {
		return fmt.Errorf("action tidak valid: %q", action)
	}
	return nil
}

// ── Shared helpers ──────────────────────────────────────────────────────────

func validateRequiredString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}
