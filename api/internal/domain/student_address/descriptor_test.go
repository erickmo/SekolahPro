package student_address_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/student_address"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &student_address.Descriptor{}
	if got := d.TableName(); got != "student_addresses" {
		t.Errorf("TableName() = %q, want %q", got, "student_addresses")
	}
}

func TestDescriptor_DefaultRels(t *testing.T) {
	d := &student_address.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 1 {
		t.Errorf("DefaultRels() length = %d, want 1", len(rels))
	}
	if _, ok := rels["student"]; !ok {
		t.Error("DefaultRels() missing student relation")
	}
}

func TestDescriptor_Rels_StudentAutoload(t *testing.T) {
	d := &student_address.Descriptor{}
	rels := d.DefaultRels()
	student := rels["student"]
	if !student.IsAutoload {
		t.Error("student relation should be autoload")
	}
	if student.Domain != "students" {
		t.Errorf("student.Domain = %q, want %q", student.Domain, "students")
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &student_address.Descriptor{}
	data := map[string]any{
		"student_id":   "018f0000-0000-7000-8000-000000000001",
		"address_type": "domisili",
		"street":       "Jl. Merdeka No. 10",
		"city":         "Bandung",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_StudentID(t *testing.T) {
	d := &student_address.Descriptor{}
	data := map[string]any{
		"address_type": "domisili",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing student_id")
	}
}

func TestDescriptor_Validate_MissingRequired_AddressType(t *testing.T) {
	d := &student_address.Descriptor{}
	data := map[string]any{
		"student_id": "018f0000-0000-7000-8000-000000000001",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing address_type")
	}
}

func TestDescriptor_Validate_InvalidEnums_AddressType(t *testing.T) {
	d := &student_address.Descriptor{}
	data := map[string]any{
		"student_id":   "018f0000-0000-7000-8000-000000000001",
		"address_type": "kantor",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid address_type")
	}
}
