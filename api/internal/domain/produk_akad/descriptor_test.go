package produk_akad_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/produk_akad"
)

func TestDescriptor_TableName(t *testing.T) {
	d := &produk_akad.Descriptor{}
	if got := d.TableName(); got != "produk_akad" {
		t.Errorf("TableName() = %q, want %q", got, "produk_akad")
	}
}

func TestDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &produk_akad.Descriptor{}
	rels := d.DefaultRels()
	if len(rels) != 0 {
		t.Errorf("DefaultRels() length = %d, want 0 (root entity)", len(rels))
	}
}

func TestDescriptor_Validate_Valid(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		"name":       "Tabungan Wadiah",
		"code":       "TW-001",
		"type":       "simpanan",
		"akad_type":  "wadiah",
		"min_amount": float64(10000),
		"max_amount": float64(100000000),
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestDescriptor_Validate_MissingRequired_Name(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		"code":      "TW-001",
		"type":      "simpanan",
		"akad_type": "wadiah",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing name")
	}
}

func TestDescriptor_Validate_MissingRequired_Code(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		"name":      "Tabungan Wadiah",
		"type":      "simpanan",
		"akad_type": "wadiah",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing code")
	}
}

func TestDescriptor_Validate_InvalidEnums_Type(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		"name":      "Tabungan Wadiah",
		"code":      "TW-001",
		"type":      "deposito",
		"akad_type": "wadiah",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid type")
	}
}

func TestDescriptor_Validate_InvalidEnums_AkadType(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		"name":      "Tabungan Wadiah",
		"code":      "TW-001",
		"type":      "simpanan",
		"akad_type": "riba",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid akad_type")
	}
}

func TestDescriptor_Validate_NegativeMinAmount(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		"name":       "Tabungan Wadiah",
		"code":       "TW-001",
		"type":       "simpanan",
		"akad_type":  "wadiah",
		"min_amount": float64(-1000),
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject negative min_amount")
	}
}

func TestDescriptor_Validate_MaxLessThanMin(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		"name":       "Tabungan Wadiah",
		"code":       "TW-001",
		"type":       "simpanan",
		"akad_type":  "wadiah",
		"min_amount": float64(100000),
		"max_amount": float64(50000),
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject max_amount < min_amount")
	}
}
