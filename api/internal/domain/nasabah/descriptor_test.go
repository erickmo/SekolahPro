package nasabah_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/nasabah"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &nasabah.Descriptor{}
	if got := d.TableName(); got != "nasabah" {
		t.Errorf("TableName() = %q, want %q", got, "nasabah")
	}
}

func TestDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &nasabah.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 0 {
		t.Errorf("DefaultRels() length = %d, want 0 (root entity)", len(rels))
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"nik":       "3201234567890001",
		"full_name": "Budi Santoso",
		"type":      "siswa",
		"gender":    "L",
		"status":    "active",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_NIK(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"full_name": "Budi Santoso",
		"type":      "siswa",
		"gender":    "L",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing nik")
	}
}

func TestDescriptor_Validate_MissingRequired_Type(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"nik":       "3201234567890001",
		"full_name": "Budi Santoso",
		"gender":    "L",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing type")
	}
}

func TestDescriptor_Validate_MissingRequired_Gender(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"nik":       "3201234567890001",
		"full_name": "Budi Santoso",
		"type":      "siswa",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing gender")
	}
}

func TestDescriptor_Validate_InvalidEnums_NIK(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"nik":       "123",
		"full_name": "Budi Santoso",
		"type":      "siswa",
		"gender":    "L",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid NIK format")
	}
}

func TestDescriptor_Validate_InvalidEnums_Type(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"nik":       "3201234567890001",
		"full_name": "Budi Santoso",
		"type":      "mahasiswa",
		"gender":    "L",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid type")
	}
}

func TestDescriptor_Validate_InvalidEnums_Gender(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"nik":       "3201234567890001",
		"full_name": "Budi Santoso",
		"type":      "siswa",
		"gender":    "X",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid gender")
	}
}

func TestDescriptor_Validate_InvalidEnums_Status(t *testing.T) {
	d := &nasabah.Descriptor{}
	data := map[string]any{
		"nik":       "3201234567890001",
		"full_name": "Budi Santoso",
		"type":      "siswa",
		"gender":    "L",
		"status":    "unknown",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}
