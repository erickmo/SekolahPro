package student_document_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/student_document"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &student_document.Descriptor{}
	if got := d.TableName(); got != "student_documents" {
		t.Errorf("TableName() = %q, want %q", got, "student_documents")
	}
}

func TestDescriptor_DefaultRels(t *testing.T) {
	d := &student_document.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 1 {
		t.Errorf("DefaultRels() length = %d, want 1", len(rels))
	}
	if _, ok := rels["student"]; !ok {
		t.Error("DefaultRels() missing student relation")
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &student_document.Descriptor{}
	data := map[string]any{
		"student_id":     "018f0000-0000-7000-8000-000000000001",
		"document_type":  "akta_kelahiran",
		"file_url":       "https://storage.example.com/doc.pdf",
		"status":         "pending",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_StudentID(t *testing.T) {
	d := &student_document.Descriptor{}
	data := map[string]any{
		"document_type": "akta_kelahiran",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing student_id")
	}
}

func TestDescriptor_Validate_MissingRequired_DocumentType(t *testing.T) {
	d := &student_document.Descriptor{}
	data := map[string]any{
		"student_id": "018f0000-0000-7000-8000-000000000001",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing document_type")
	}
}

func TestDescriptor_Validate_InvalidEnums_DocumentType(t *testing.T) {
	d := &student_document.Descriptor{}
	data := map[string]any{
		"student_id":     "018f0000-0000-7000-8000-000000000001",
		"document_type":  "surat_nikah",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid document_type")
	}
}

func TestDescriptor_Validate_InvalidEnums_Status(t *testing.T) {
	d := &student_document.Descriptor{}
	data := map[string]any{
		"student_id":     "018f0000-0000-7000-8000-000000000001",
		"document_type":  "akta_kelahiran",
		"status":         "unknown",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}
