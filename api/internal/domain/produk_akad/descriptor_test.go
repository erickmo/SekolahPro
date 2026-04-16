package produk_akad_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/produk_akad"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &produk_akad.Descriptor{}
	assert.Equal(t, "produk_akad", d.TableName())
}

func TestDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &produk_akad.Descriptor{}
	rels := d.DefaultRels()
	assert.Empty(t, rels, "produk_akad is a root entity — should have no relations")
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	// Compile-time check: Descriptor must implement DomainDescriptor.
	var _ vernon.DomainDescriptor = &produk_akad.Descriptor{}
}

// ── Valid data — all enum combinations (type x akad_type) ───────────────────

func TestDescriptor_Validate_AllTypeAkadTypeCombinations(t *testing.T) {
	d := &produk_akad.Descriptor{}

	types := []string{
		produk_akad.TypeSimpanan,
		produk_akad.TypePembiayaan,
	}
	akadTypes := []string{
		produk_akad.AkadWadiah,
		produk_akad.AkadMudharabah,
		produk_akad.AkadMusyarakah,
		produk_akad.AkadMurabahah,
		produk_akad.AkadIjarah,
		produk_akad.AkadQard,
	}

	for _, typ := range types {
		for _, akad := range akadTypes {
			data := map[string]any{
				produk_akad.FieldName: "Produk Test",
				produk_akad.FieldCode: "TST-001",
				produk_akad.FieldType: typ,
				produk_akad.FieldAkadType: akad,
			}
			err := d.Validate(data)
			assert.NoError(t, err, "type=%q akad_type=%q should be valid", typ, akad)
		}
	}
}

func TestDescriptor_Validate_ValidFullData(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:         "Tabungan Wadiah",
		produk_akad.FieldCode:         "TW-001",
		produk_akad.FieldType:         produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType:     produk_akad.AkadWadiah,
		produk_akad.FieldMinAmount:    float64(10000),
		produk_akad.FieldMaxAmount:    float64(100000000),
		produk_akad.FieldMinTenor:     float64(1),
		produk_akad.FieldMaxTenor:     float64(12),
		produk_akad.FieldTenorUnit:    "bulan",
		produk_akad.FieldInterestRate: float64(3.5),
		produk_akad.FieldDescription:  "Tabungan wadiah untuk siswa",
		produk_akad.FieldStatus:       "active",
	}
	err := d.Validate(data)
	assert.NoError(t, err)
}

// ── Missing required fields ─────────────────────────────────────────────────

func TestDescriptor_Validate_MissingRequired_Name(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldCode:     "TW-001",
		produk_akad.FieldType:     produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType: produk_akad.AkadWadiah,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name wajib diisi")
}

func TestDescriptor_Validate_MissingRequired_Code(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:     "Tabungan Wadiah",
		produk_akad.FieldType:     produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType: produk_akad.AkadWadiah,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "code wajib diisi")
}

func TestDescriptor_Validate_MissingRequired_Type(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:     "Tabungan Wadiah",
		produk_akad.FieldCode:     "TW-001",
		produk_akad.FieldAkadType: produk_akad.AkadWadiah,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type wajib diisi")
}

func TestDescriptor_Validate_MissingRequired_AkadType(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName: "Tabungan Wadiah",
		produk_akad.FieldCode: "TW-001",
		produk_akad.FieldType: produk_akad.TypeSimpanan,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "akad_type wajib diisi")
}

// ── Invalid enum values ─────────────────────────────────────────────────────

func TestDescriptor_Validate_InvalidType(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:     "Tabungan Wadiah",
		produk_akad.FieldCode:     "TW-001",
		produk_akad.FieldType:     "deposito",
		produk_akad.FieldAkadType: produk_akad.AkadWadiah,
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type tidak valid")
}

func TestDescriptor_Validate_InvalidAkadType(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:     "Tabungan Wadiah",
		produk_akad.FieldCode:     "TW-001",
		produk_akad.FieldType:     produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType: "riba",
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "akad_type tidak valid")
}

// ── Amount boundary conditions ──────────────────────────────────────────────

func TestDescriptor_Validate_NegativeMinAmount(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:      "Tabungan Wadiah",
		produk_akad.FieldCode:      "TW-001",
		produk_akad.FieldType:      produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType:  produk_akad.AkadWadiah,
		produk_akad.FieldMinAmount: float64(-1000),
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "min_amount tidak boleh negatif")
}

func TestDescriptor_Validate_NegativeMaxAmount(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:      "Tabungan Wadiah",
		produk_akad.FieldCode:      "TW-001",
		produk_akad.FieldType:      produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType:  produk_akad.AkadWadiah,
		produk_akad.FieldMaxAmount: float64(-500),
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max_amount tidak boleh negatif")
}

func TestDescriptor_Validate_MaxLessThanMin(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:      "Tabungan Wadiah",
		produk_akad.FieldCode:      "TW-001",
		produk_akad.FieldType:      produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType:  produk_akad.AkadWadiah,
		produk_akad.FieldMinAmount: float64(100000),
		produk_akad.FieldMaxAmount: float64(50000),
	}
	err := d.Validate(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max_amount harus >= min_amount")
}

func TestDescriptor_Validate_ZeroAmounts(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:      "Tabungan Wadiah",
		produk_akad.FieldCode:      "TW-001",
		produk_akad.FieldType:      produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType:  produk_akad.AkadWadiah,
		produk_akad.FieldMinAmount: float64(0),
		produk_akad.FieldMaxAmount: float64(0),
	}
	err := d.Validate(data)
	assert.NoError(t, err, "zero amounts should be valid")
}

func TestDescriptor_Validate_EqualMinMaxAmounts(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:      "Tabungan Wadiah",
		produk_akad.FieldCode:      "TW-001",
		produk_akad.FieldType:      produk_akad.TypeSimpanan,
		produk_akad.FieldAkadType:  produk_akad.AkadWadiah,
		produk_akad.FieldMinAmount: float64(50000),
		produk_akad.FieldMaxAmount: float64(50000),
	}
	err := d.Validate(data)
	assert.NoError(t, err, "equal min/max amounts should be valid")
}

func TestDescriptor_Validate_VeryLargeAmounts(t *testing.T) {
	d := &produk_akad.Descriptor{}
	data := map[string]any{
		produk_akad.FieldName:      "Pembiayaan Besar",
		produk_akad.FieldCode:      "PB-999",
		produk_akad.FieldType:      produk_akad.TypePembiayaan,
		produk_akad.FieldAkadType:  produk_akad.AkadMurabahah,
		produk_akad.FieldMinAmount: float64(999999999999),
		produk_akad.FieldMaxAmount: float64(999999999999),
	}
	err := d.Validate(data)
	assert.NoError(t, err, "very large amounts should be valid")
}

// ── Empty/nil data map ──────────────────────────────────────────────────────

func TestDescriptor_Validate_NilData(t *testing.T) {
	d := &produk_akad.Descriptor{}
	err := d.Validate(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name wajib diisi")
}

func TestDescriptor_Validate_EmptyData(t *testing.T) {
	d := &produk_akad.Descriptor{}
	err := d.Validate(map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name wajib diisi")
}

// ── Wrong type for field values ─────────────────────────────────────────────

func TestDescriptor_Validate_EmptyStringRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantErr  string
	}{
		{
			name: "empty name",
			data: map[string]any{
				produk_akad.FieldName:     "",
				produk_akad.FieldCode:     "TW-001",
				produk_akad.FieldType:     produk_akad.TypeSimpanan,
				produk_akad.FieldAkadType: produk_akad.AkadWadiah,
			},
			wantErr: "name wajib diisi",
		},
		{
			name: "empty code",
			data: map[string]any{
				produk_akad.FieldName:     "Tabungan Wadiah",
				produk_akad.FieldCode:     "",
				produk_akad.FieldType:     produk_akad.TypeSimpanan,
				produk_akad.FieldAkadType: produk_akad.AkadWadiah,
			},
			wantErr: "code wajib diisi",
		},
		{
			name: "empty type",
			data: map[string]any{
				produk_akad.FieldName:     "Tabungan Wadiah",
				produk_akad.FieldCode:     "TW-001",
				produk_akad.FieldType:     "",
				produk_akad.FieldAkadType: produk_akad.AkadWadiah,
			},
			wantErr: "type wajib diisi",
		},
		{
			name: "empty akad_type",
			data: map[string]any{
				produk_akad.FieldName:     "Tabungan Wadiah",
				produk_akad.FieldCode:     "TW-001",
				produk_akad.FieldType:     produk_akad.TypeSimpanan,
				produk_akad.FieldAkadType: "",
			},
			wantErr: "akad_type wajib diisi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &produk_akad.Descriptor{}
			err := d.Validate(tt.data)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
