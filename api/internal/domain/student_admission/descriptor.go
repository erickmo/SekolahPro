// Package student_admission adalah domain Vernon untuk penerimaan siswa baru (PPDB).
//
// Student admission memiliki 2 BelongsTo autoload: student dan academic_year.
// Mengelola proses dari pendaftaran hingga diterima/ditolak.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (student_admissions).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student_admission

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldRegistrationNumber = "registration_number"
	FieldStudentID          = "student_id"
	FieldAcademicYearID     = "academic_year_id"
	FieldAdmissionType      = "admission_type"
	FieldStatus             = "status"
	FieldRegistrationDate   = "registration_date"
	FieldTestDate           = "test_date"
	FieldTestScore          = "test_score"
	FieldInterviewScore     = "interview_score"
	FieldFinalScore         = "final_score"
	FieldNotes              = "notes"
	FieldApprovedBy         = "approved_by"
)

// Admission type constants.
const (
	TypeRegular  = "regular"
	TypePindahan = "pindahan"
	TypeAfirmasi = "afirmasi"
	TypePrestasi = "prestasi"
)

// Status constants.
const (
	StatusPending     = "pending"
	StatusDocReview   = "document_review"
	StatusTest        = "test"
	StatusWrittenTest = "written_test"
	StatusInterview   = "interview"
	StatusAccepted    = "accepted"
	StatusRejected    = "rejected"
)

// Relation name constants.
const (
	RelStudent      = "student"
	RelAcademicYear = "academic_year"
)

var validAdmissionTypes = map[string]bool{
	TypeRegular: true, TypePindahan: true, TypeAfirmasi: true, TypePrestasi: true,
}

var validStatuses = map[string]bool{
	StatusPending: true, StatusDocReview: true, StatusTest: true,
	StatusWrittenTest: true, StatusInterview: true,
	StatusAccepted: true, StatusRejected: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk student_admissions.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "student_admissions" }

// DefaultRels mendefinisikan relasi domain ini.
// 2 BelongsTo autoload: student, academic_year.
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
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearID,
			LocalKey:   FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	regNumber, _ := data[FieldRegistrationNumber].(string)
	if regNumber == "" {
		return errors.New("registration_number wajib diisi")
	}

	admissionType, _ := data[FieldAdmissionType].(string)
	if admissionType == "" {
		return errors.New("admission_type wajib diisi")
	}
	if !validAdmissionTypes[admissionType] {
		return fmt.Errorf("admission_type tidak valid: %q", admissionType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	return validateScores(data)
}

// validateScores memeriksa skor tidak boleh negatif.
func validateScores(data map[string]any) error {
	scoreFields := []string{FieldTestScore, FieldInterviewScore, FieldFinalScore}
	for _, field := range scoreFields {
		if score, ok := data[field].(float64); ok && score < 0 {
			return fmt.Errorf("%s tidak boleh negatif", field)
		}
	}
	return nil
}
