// Package teacher_attendance adalah domain Vernon untuk absensi guru & staff.
//
// Terdiri dari 2 tabel: teacher_attendances dan teacher_attendance_configs.
// teacher_attendances memiliki 2 BelongsTo autoload: teacher dan academic_year.
// teacher_attendance_configs adalah tabel konfigurasi standalone (tanpa BelongsTo).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teacher_attendance

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Attendance field name constants ────────────────────────────────────────────

const (
	FieldTeacherID      = "teacher_id"
	FieldAcademicYearID = "academic_year_id"
	FieldAttendanceDate = "attendance_date"
	FieldStatus         = "status"
	FieldClockInMethod  = "clock_in_method"
	FieldClockOutMethod = "clock_out_method"
	FieldLateMinutes    = "late_minutes"
	FieldEarlyLeave     = "early_leave_minutes"
	FieldNote           = "note"
)

// ── Config field name constants ────────────────────────────────────────────────

const (
	FieldWorkStartTime        = "work_start_time"
	FieldWorkEndTime          = "work_end_time"
	FieldLateToleranceMinutes = "late_tolerance_minutes"
)

// ── Attendance status constants ────────────────────────────────────────────────

const (
	AttStatusPresent   = "present"
	AttStatusSick      = "sick"
	AttStatusPermitted = "permitted"
	AttStatusAbsent    = "absent"
	AttStatusDinasLuar = "dinas_luar"
	AttStatusCuti      = "cuti"
	AttStatusLibur     = "libur"
)

// ── Clock method constants ─────────────────────────────────────────────────────

const (
	MethodFingerprint    = "fingerprint"
	MethodFaceRecognition = "face_recognition"
	MethodGPS            = "gps"
	MethodManual         = "manual"
	MethodQRCode         = "qr_code"
)

// ── Relation name constants ────────────────────────────────────────────────────

const (
	RelTeacher      = "teacher"
	RelAcademicYear = "academic_year"
)

var validAttendanceStatuses = map[string]bool{
	AttStatusPresent: true, AttStatusSick: true, AttStatusPermitted: true,
	AttStatusAbsent: true, AttStatusDinasLuar: true,
	AttStatusCuti: true, AttStatusLibur: true,
}

var validClockMethods = map[string]bool{
	MethodFingerprint: true, MethodFaceRecognition: true,
	MethodGPS: true, MethodManual: true, MethodQRCode: true,
}

// ── Descriptor (teacher_attendances) ───────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk teacher_attendances.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "teacher_attendances" }

// DefaultRels mendefinisikan relasi teacher_attendances: teacher + academic_year.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTeacherID,
			LocalKey:   FieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type", "role"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearID,
			LocalKey:   FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant teacher_attendances sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	teacherID, _ := data[FieldTeacherID].(string)
	if teacherID == "" {
		return fmt.Errorf("teacher_id wajib diisi")
	}

	academicYearID, _ := data[FieldAcademicYearID].(string)
	if academicYearID == "" {
		return fmt.Errorf("academic_year_id wajib diisi")
	}

	attendanceDate, _ := data[FieldAttendanceDate].(string)
	if attendanceDate == "" {
		return fmt.Errorf("attendance_date wajib diisi")
	}

	status, _ := data[FieldStatus].(string)
	if status == "" {
		return fmt.Errorf("status wajib diisi")
	}
	if !validAttendanceStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	if err := validateOptionalEnum(data, FieldClockInMethod, "clock_in_method", validClockMethods); err != nil {
		return err
	}
	return validateOptionalEnum(data, FieldClockOutMethod, "clock_out_method", validClockMethods)
}

// ── ConfigDescriptor (teacher_attendance_configs) ──────────────────────────────

// ConfigDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_attendance_configs.
// Tabel konfigurasi tanpa relasi BelongsTo.
type ConfigDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL untuk konfigurasi absensi.
func (d *ConfigDescriptor) TableName() string { return "teacher_attendance_configs" }

// DefaultRels mengembalikan relasi kosong — tabel konfigurasi standalone.
func (d *ConfigDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant teacher_attendance_configs sebelum write.
func (d *ConfigDescriptor) Validate(data map[string]any) error {
	workStart, _ := data[FieldWorkStartTime].(string)
	if workStart == "" {
		return fmt.Errorf("work_start_time wajib diisi")
	}

	workEnd, _ := data[FieldWorkEndTime].(string)
	if workEnd == "" {
		return fmt.Errorf("work_end_time wajib diisi")
	}

	tolerance, ok := data[FieldLateToleranceMinutes]
	if !ok {
		return fmt.Errorf("late_tolerance_minutes wajib diisi")
	}
	toleranceNum, isNum := tolerance.(float64)
	if !isNum || toleranceNum < 0 {
		return fmt.Errorf("late_tolerance_minutes harus berupa angka >= 0")
	}
	return nil
}

// ── shared validation helpers ──────────────────────────────────────────────────

func validateOptionalEnum(data map[string]any, field, label string, valid map[string]bool) error {
	val, _ := data[field].(string)
	if val == "" {
		return nil
	}
	if !valid[val] {
		return fmt.Errorf("%s tidak valid: %q", label, val)
	}
	return nil
}
