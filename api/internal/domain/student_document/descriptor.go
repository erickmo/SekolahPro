// Package student_document adalah domain Vernon untuk dokumen siswa.
//
// Student document belongs_to students dengan autoload. Menyimpan
// berbagai jenis dokumen: akta, KK, ijazah, foto, KTP ortu.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (student_documents).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student_document

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID     = "student_id"
	FieldDocumentType  = "document_type"
	FieldFileURL       = "file_url"
	FieldFileName      = "file_name"
	FieldFileSize      = "file_size"
	FieldMimeType      = "mime_type"
	FieldVerifiedBy    = "verified_by"
	FieldVerifiedAt    = "verified_at"
	FieldStatus        = "status"
	FieldNotes         = "notes"
)

// Document type constants.
const (
	TypeAkta    = "akta_kelahiran"
	TypeKK      = "kartu_keluarga"
	TypeIjazah  = "ijazah"
	TypeSKHUN   = "skhun"
	TypePhoto   = "pass_photo"
	TypeKTPOrtu = "ktp_ortu"
)

// Status constants.
const (
	StatusPending  = "pending"
	StatusVerified = "verified"
	StatusRejected = "rejected"
)

// Relation name constants.
const (
	RelStudent = "student"
)

var validDocTypes = map[string]bool{
	TypeAkta: true, TypeKK: true, TypeIjazah: true,
	TypeSKHUN: true, TypePhoto: true, TypeKTPOrtu: true,
}

var validStatuses = map[string]bool{
	StatusPending: true, StatusVerified: true, StatusRejected: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk student_documents.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "student_documents" }

// DefaultRels mendefinisikan relasi domain ini.
// student_documents belongs_to students — autoload.
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
		return errors.New("student_id wajib diisi")
	}

	docType, _ := data[FieldDocumentType].(string)
	if docType == "" {
		return errors.New("document_type wajib diisi")
	}
	if !validDocTypes[docType] {
		return fmt.Errorf("document_type tidak valid: %q", docType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus pending/verified/rejected)", status)
	}
	return nil
}
