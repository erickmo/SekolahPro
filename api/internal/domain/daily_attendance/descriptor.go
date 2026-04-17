// Package daily_attendance adalah domain Vernon untuk kehadiran harian siswa.
//
// DailyAttendance memiliki 2 BelongsTo autoload: student dan class_room.
// Digunakan untuk tracking kehadiran siswa per hari per kelas.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package daily_attendance

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID      = "student_id"
	FieldClassRoomID    = "class_room_id"
	FieldAttendanceDate = "attendance_date"
	FieldStatus         = "status"
	FieldNotes          = "notes"
)

// Attendance status constants.
const (
	StatusPresent  = "present"
	StatusAbsent   = "absent"
	StatusLate     = "late"
	StatusExcused  = "excused"
	StatusSick     = "sick"
	StatusPermit   = "permit"
)

// Relation name constants.
const (
	RelStudent   = "student"
	RelClassRoom = "class_room"
)

var validStatuses = map[string]bool{
	StatusPresent: true, StatusAbsent: true, StatusLate: true,
	StatusExcused: true, StatusSick: true, StatusPermit: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk daily_attendances.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "daily_attendances" }

// DefaultRels mendefinisikan relasi domain ini.
// 2 BelongsTo autoload: student, class_room.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         FieldStudentID,
			LocalKey:   FieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
		RelClassRoom: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         FieldClassRoomID,
			LocalKey:   FieldClassRoomID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "grade_level"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	studentID, _ := data[FieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	classRoomID, _ := data[FieldClassRoomID].(string)
	if classRoomID == "" {
		return fmt.Errorf("class_room_id wajib diisi")
	}

	attendanceDate, _ := data[FieldAttendanceDate].(string)
	if attendanceDate == "" {
		return fmt.Errorf("attendance_date wajib diisi")
	}

	status, _ := data[FieldStatus].(string)
	if status == "" {
		return fmt.Errorf("status wajib diisi")
	}
	if !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
