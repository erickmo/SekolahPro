// Package health_record adalah domain Vernon untuk catatan kesehatan siswa.
//
// HealthRecord memiliki 1 BelongsTo autoload: student.
// Digunakan untuk tracking riwayat kesehatan, vaksinasi, dan kondisi medis siswa.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package health_record

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID    = "student_id"
	FieldRecordType   = "record_type"
	FieldRecordDate   = "record_date"
	FieldDescription  = "description"
	FieldDiagnosis    = "diagnosis"
	FieldTreatment    = "treatment"
	FieldDoctorName   = "doctor_name"
	FieldFollowUpDate = "follow_up_date"
	FieldStatus       = "status"
)

// Record type constants.
const (
	RecordTypeGeneral    = "general"
	RecordTypeVaccination = "vaccination"
	RecordTypeAllergy    = "allergy"
	RecordTypeInjury     = "injury"
	RecordTypeChronic    = "chronic"
	RecordTypeVision     = "vision"
	RecordTypeDental     = "dental"
)

// Relation name constants.
const (
	RelStudent = "student"
)

var validRecordTypes = map[string]bool{
	RecordTypeGeneral: true, RecordTypeVaccination: true,
	RecordTypeAllergy: true, RecordTypeInjury: true,
	RecordTypeChronic: true, RecordTypeVision: true, RecordTypeDental: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk health_records.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "health_records" }

// DefaultRels mendefinisikan relasi domain ini.
// 1 BelongsTo autoload: student.
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
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	studentID, _ := data[FieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	recordType, _ := data[FieldRecordType].(string)
	if recordType != "" && !validRecordTypes[recordType] {
		return fmt.Errorf("record_type tidak valid: %q", recordType)
	}
	return nil
}
