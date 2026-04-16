package rekening_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/rekening"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &rekening.Descriptor{}
	if got := d.TableName(); got != "rekening" {
		t.Errorf("TableName() = %q, want %q", got, "rekening")
	}
}

func TestDescriptor_DefaultRels(t *testing.T) {
	d := &rekening.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 2 {
		t.Errorf("DefaultRels() length = %d, want 2", len(rels))
	}
	if _, ok := rels["nasabah"]; !ok {
		t.Error("DefaultRels() missing nasabah relation")
	}
	if _, ok := rels["produk_akad"]; !ok {
		t.Error("DefaultRels() missing produk_akad relation")
	}
}

func TestDescriptor_Rels_NasabahAutoload(t *testing.T) {
	d := &rekening.Descriptor{}
	rels := d.DefaultRels()
	nasabah := rels["nasabah"]
	if !nasabah.IsAutoload {
		t.Error("nasabah relation should be autoload")
	}
	if nasabah.Domain != "nasabah" {
		t.Errorf("nasabah.Domain = %q, want %q", nasabah.Domain, "nasabah")
	}
}

func TestDescriptor_Rels_ProdukAkadAutoload(t *testing.T) {
	d := &rekening.Descriptor{}
	rels := d.DefaultRels()
	pa := rels["produk_akad"]
	if !pa.IsAutoload {
		t.Error("produk_akad relation should be autoload")
	}
	if pa.Domain != "produk_akad" {
		t.Errorf("produk_akad.Domain = %q, want %q", pa.Domain, "produk_akad")
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		"nasabah_id":     "018f0000-0000-7000-8000-000000000001",
		"produk_akad_id": "018f0000-0000-7000-8000-000000000002",
		"no_rekening":    "REK-001",
		"balance":        float64(500000),
		"hold_balance":   float64(0),
		"status":         "active",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_NegativeBalance(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		"balance": float64(-1000),
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject negative balance")
	}
}

func TestDescriptor_Validate_HoldExceedsBalance(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		"balance":      float64(1000),
		"hold_balance": float64(5000),
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject hold_balance > balance")
	}
}

func TestDescriptor_Validate_InvalidEnums_Status(t *testing.T) {
	d := &rekening.Descriptor{}
	data := map[string]any{
		"status": "unknown",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}
