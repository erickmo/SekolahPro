package jaminan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/jaminan"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ═══════════════════════════════════════════════════════════════════════════════
// JAMINAN DESCRIPTOR TESTS
// ═══════════════════════════════════════════════════════════════════════════════

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestJaminanDescriptor_TableName(t *testing.T) {
	d := &jaminan.Descriptor{}
	assert.Equal(t, "jaminan", d.TableName())
}

func TestJaminanDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jaminan.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestJaminanDescriptor_DefaultRels_Count(t *testing.T) {
	d := &jaminan.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "jaminan should have 2 BelongsTo relations")
}

func TestJaminanDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &jaminan.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jaminan.RelNasabah]
	require.True(t, ok, "should have nasabah relation")

	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, jaminan.FieldNasabahID, rel.FK)
	assert.True(t, rel.IsAutoload, "nasabah should be autoloaded")
	assert.Equal(t, []string{"nama_lengkap", "no_nasabah", "type"}, rel.Fields)
}

func TestJaminanDescriptor_DefaultRels_RekeningRelation(t *testing.T) {
	d := &jaminan.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jaminan.RelRekening]
	require.True(t, ok, "should have rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, jaminan.FieldRekeningID, rel.FK)
	assert.True(t, rel.IsAutoload, "rekening should be autoloaded")
	assert.Equal(t, []string{"no_rekening", "category", "balance"}, rel.Fields)
}

// ── Validation tests — valid data ────────────────────────────────────────────

func validJaminanData() map[string]any {
	return map[string]any{
		jaminan.FieldNasabahID:      "00000000-0000-0000-0000-000000000001",
		jaminan.FieldRekeningID:     "00000000-0000-0000-0000-000000000002",
		jaminan.FieldCollateralType: jaminan.CollateralInternalBalance,
		jaminan.FieldDescription:    "Tabungan siswa sebagai jaminan",
		jaminan.FieldAppraisedValue: float64(5000000),
		jaminan.FieldAcceptanceRate: float64(1.0),
		jaminan.FieldCollateralVal:  float64(5000000),
		jaminan.FieldStatus:         jaminan.StatusRegistered,
	}
}

func TestJaminanDescriptor_Validate_Success(t *testing.T) {
	d := &jaminan.Descriptor{}
	err := d.Validate(validJaminanData())
	assert.NoError(t, err)
}

func TestJaminanDescriptor_Validate_AllCollateralTypes(t *testing.T) {
	d := &jaminan.Descriptor{}
	types := []string{
		jaminan.CollateralInternalBalance, jaminan.CollateralSuratBerharga,
		jaminan.CollateralBarangBergerak, jaminan.CollateralPersonalGuarantee,
	}
	for _, ct := range types {
		data := validJaminanData()
		data[jaminan.FieldCollateralType] = ct
		err := d.Validate(data)
		assert.NoError(t, err, "collateral_type=%q should be valid", ct)
	}
}

func TestJaminanDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &jaminan.Descriptor{}
	statuses := []string{
		jaminan.StatusRegistered, jaminan.StatusPledged,
		jaminan.StatusReleased, jaminan.StatusForeclosed,
	}
	for _, status := range statuses {
		data := validJaminanData()
		data[jaminan.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestJaminanDescriptor_Validate_AcceptanceRateBoundary(t *testing.T) {
	d := &jaminan.Descriptor{}

	// rate = 0 valid
	data := validJaminanData()
	data[jaminan.FieldAcceptanceRate] = float64(0)
	assert.NoError(t, d.Validate(data))

	// rate = 1 valid
	data = validJaminanData()
	data[jaminan.FieldAcceptanceRate] = float64(1)
	assert.NoError(t, d.Validate(data))

	// rate = 0.5 valid
	data = validJaminanData()
	data[jaminan.FieldAcceptanceRate] = float64(0.5)
	assert.NoError(t, d.Validate(data))
}

func TestJaminanDescriptor_Validate_ZeroAmounts(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldAppraisedValue] = float64(0)
	data[jaminan.FieldCollateralVal] = float64(0)
	err := d.Validate(data)
	assert.NoError(t, err, "zero amounts should be valid")
}

// ── Validation tests — missing required fields ───────────────────────────────

func TestJaminanDescriptor_Validate_MissingNasabahID(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	delete(data, jaminan.FieldNasabahID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestJaminanDescriptor_Validate_MissingCollateralType(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	delete(data, jaminan.FieldCollateralType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "collateral_type wajib diisi")
}

func TestJaminanDescriptor_Validate_MissingDescription(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	delete(data, jaminan.FieldDescription)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "description wajib diisi")
}

func TestJaminanDescriptor_Validate_EmptyDescription(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldDescription] = ""
	err := d.Validate(data)
	assert.ErrorContains(t, err, "description wajib diisi")
}

// ── Validation tests — invalid enums ─────────────────────────────────────────

func TestJaminanDescriptor_Validate_InvalidCollateralType(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldCollateralType] = "real_estate"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "collateral_type tidak valid")
}

func TestJaminanDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldStatus] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

// ── Validation tests — range violations ───────────────────────────────────────

func TestJaminanDescriptor_Validate_NegativeAppraisedValue(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldAppraisedValue] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "appraised_value tidak boleh negatif")
}

func TestJaminanDescriptor_Validate_AcceptanceRateOverOne(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldAcceptanceRate] = float64(1.5)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "acceptance_rate harus antara 0 dan 1")
}

func TestJaminanDescriptor_Validate_AcceptanceRateNegative(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldAcceptanceRate] = float64(-0.1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "acceptance_rate harus antara 0 dan 1")
}

func TestJaminanDescriptor_Validate_NegativeCollateralValue(t *testing.T) {
	d := &jaminan.Descriptor{}
	data := validJaminanData()
	data[jaminan.FieldCollateralVal] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "collateral_value tidak boleh negatif")
}

// ── Validation tests — edge cases ────────────────────────────────────────────

func TestJaminanDescriptor_Validate_NilData(t *testing.T) {
	d := &jaminan.Descriptor{}
	err := d.Validate(nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestJaminanDescriptor_Validate_EmptyData(t *testing.T) {
	d := &jaminan.Descriptor{}
	err := d.Validate(map[string]any{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

// ═══════════════════════════════════════════════════════════════════════════════
// JAMINAN_PINJAMAN DESCRIPTOR TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestPinjamanDescriptor_TableName(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	assert.Equal(t, "jaminan_pinjaman", d.TableName())
}

func TestPinjamanDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jaminan.PinjamanDescriptor{}
}

func TestPinjamanDescriptor_DefaultRels_Count(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "jaminan_pinjaman should have 2 BelongsTo relations")
}

func TestPinjamanDescriptor_DefaultRels_JaminanRelation(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jaminan.JPRelJaminan]
	require.True(t, ok, "should have jaminan relation")

	assert.Equal(t, "jaminan", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "jaminan should be autoloaded")
}

func TestPinjamanDescriptor_DefaultRels_PinjamanRelation(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jaminan.JPRelPinjaman]
	require.True(t, ok, "should have pinjaman relation")

	assert.Equal(t, "pinjaman", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "pinjaman should be autoloaded")
}

func validJPData() map[string]any {
	return map[string]any{
		jaminan.JPFieldJaminanID:    "00000000-0000-0000-0000-000000000001",
		jaminan.JPFieldPinjamanID:   "00000000-0000-0000-0000-000000000002",
		jaminan.JPFieldPledgedValue: float64(3000000),
		jaminan.JPFieldStatus:       jaminan.JPStatusActive,
	}
}

func TestPinjamanDescriptor_Validate_Success(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	err := d.Validate(validJPData())
	assert.NoError(t, err)
}

func TestPinjamanDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	statuses := []string{jaminan.JPStatusActive, jaminan.JPStatusReleased}
	for _, status := range statuses {
		data := validJPData()
		data[jaminan.JPFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestPinjamanDescriptor_Validate_MissingJaminanID(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	data := validJPData()
	delete(data, jaminan.JPFieldJaminanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "jaminan_id wajib diisi")
}

func TestPinjamanDescriptor_Validate_MissingPinjamanID(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	data := validJPData()
	delete(data, jaminan.JPFieldPinjamanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "pinjaman_id wajib diisi")
}

func TestPinjamanDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	data := validJPData()
	data[jaminan.JPFieldStatus] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestPinjamanDescriptor_Validate_NegativePledgedValue(t *testing.T) {
	d := &jaminan.PinjamanDescriptor{}
	data := validJPData()
	data[jaminan.JPFieldPledgedValue] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "pledged_value tidak boleh negatif")
}

// ═══════════════════════════════════════════════════════════════════════════════
// JAMINAN_DOKUMEN DESCRIPTOR TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestDokumenDescriptor_TableName(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	assert.Equal(t, "jaminan_dokumen", d.TableName())
}

func TestDokumenDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jaminan.DokumenDescriptor{}
}

func TestDokumenDescriptor_DefaultRels_Count(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "jaminan_dokumen should have 1 BelongsTo relation")
}

func TestDokumenDescriptor_DefaultRels_JaminanRelation(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jaminan.JDRelJaminan]
	require.True(t, ok, "should have jaminan relation")

	assert.Equal(t, "jaminan", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "jaminan should be autoloaded")
}

func validJDData() map[string]any {
	return map[string]any{
		jaminan.JDFieldJaminanID: "00000000-0000-0000-0000-000000000001",
		jaminan.JDFieldDocType:   jaminan.DocFotoBarang,
		jaminan.JDFieldFileURL:   "https://storage.example.com/dokumen/foto.jpg",
		jaminan.JDFieldFileName:  "foto_jaminan.jpg",
		jaminan.JDFieldFileSize:  float64(102400),
		jaminan.JDFieldMimeType:  "image/jpeg",
	}
}

func TestDokumenDescriptor_Validate_Success(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	err := d.Validate(validJDData())
	assert.NoError(t, err)
}

func TestDokumenDescriptor_Validate_AllDocTypes(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	docTypes := []string{
		jaminan.DocFotoBarang, jaminan.DocFotokopiBPKB,
		jaminan.DocSertifikatTanah, jaminan.DocSuratKuasa,
		jaminan.DocBeritaAcara, jaminan.DocFotoIdentitas,
		jaminan.DocSuratPernyataan, jaminan.DocIjazah,
		jaminan.DocLainnya,
	}
	for _, dt := range docTypes {
		data := validJDData()
		data[jaminan.JDFieldDocType] = dt
		err := d.Validate(data)
		assert.NoError(t, err, "document_type=%q should be valid", dt)
	}
}

func TestDokumenDescriptor_Validate_MissingJaminanID(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	delete(data, jaminan.JDFieldJaminanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "jaminan_id wajib diisi")
}

func TestDokumenDescriptor_Validate_MissingDocType(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	delete(data, jaminan.JDFieldDocType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "document_type wajib diisi")
}

func TestDokumenDescriptor_Validate_InvalidDocType(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	data[jaminan.JDFieldDocType] = "passport"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "document_type tidak valid")
}

func TestDokumenDescriptor_Validate_MissingFileURL(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	delete(data, jaminan.JDFieldFileURL)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_url wajib diisi")
}

func TestDokumenDescriptor_Validate_MissingFileName(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	delete(data, jaminan.JDFieldFileName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_name wajib diisi")
}

func TestDokumenDescriptor_Validate_MissingMimeType(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	delete(data, jaminan.JDFieldMimeType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "mime_type wajib diisi")
}

func TestDokumenDescriptor_Validate_FileSizeZero(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	data[jaminan.JDFieldFileSize] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_size_bytes harus lebih dari 0")
}

func TestDokumenDescriptor_Validate_FileSizeNegative(t *testing.T) {
	d := &jaminan.DokumenDescriptor{}
	data := validJDData()
	data[jaminan.JDFieldFileSize] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_size_bytes harus lebih dari 0")
}

// ═══════════════════════════════════════════════════════════════════════════════
// JAMINAN_VALUASI DESCRIPTOR TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestValuasiDescriptor_TableName(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	assert.Equal(t, "jaminan_valuasi", d.TableName())
}

func TestValuasiDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jaminan.ValuasiDescriptor{}
}

func TestValuasiDescriptor_DefaultRels_Count(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "jaminan_valuasi should have 1 BelongsTo relation")
}

func TestValuasiDescriptor_DefaultRels_JaminanRelation(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jaminan.JVRelJaminan]
	require.True(t, ok, "should have jaminan relation")

	assert.Equal(t, "jaminan", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "jaminan should be autoloaded")
}

func validJVData() map[string]any {
	return map[string]any{
		jaminan.JVFieldJaminanID:     "00000000-0000-0000-0000-000000000001",
		jaminan.JVFieldAppraisedVal:  float64(5000000),
		jaminan.JVFieldAcceptanceRate: float64(0.8),
		jaminan.JVFieldCollateralVal:  float64(4000000),
		jaminan.JVFieldValuationType:  jaminan.ValuationInitial,
	}
}

func TestValuasiDescriptor_Validate_Success(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	err := d.Validate(validJVData())
	assert.NoError(t, err)
}

func TestValuasiDescriptor_Validate_AllValuationTypes(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	types := []string{
		jaminan.ValuationInitial, jaminan.ValuationPeriodic, jaminan.ValuationOnDemand,
	}
	for _, vt := range types {
		data := validJVData()
		data[jaminan.JVFieldValuationType] = vt
		err := d.Validate(data)
		assert.NoError(t, err, "valuation_type=%q should be valid", vt)
	}
}

func TestValuasiDescriptor_Validate_MissingJaminanID(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	data := validJVData()
	delete(data, jaminan.JVFieldJaminanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "jaminan_id wajib diisi")
}

func TestValuasiDescriptor_Validate_MissingValuationType(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	data := validJVData()
	delete(data, jaminan.JVFieldValuationType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "valuation_type wajib diisi")
}

func TestValuasiDescriptor_Validate_InvalidValuationType(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	data := validJVData()
	data[jaminan.JVFieldValuationType] = "annual"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "valuation_type tidak valid")
}

func TestValuasiDescriptor_Validate_NegativeAppraisedValue(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	data := validJVData()
	data[jaminan.JVFieldAppraisedVal] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "appraised_value tidak boleh negatif")
}

func TestValuasiDescriptor_Validate_AcceptanceRateOverOne(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	data := validJVData()
	data[jaminan.JVFieldAcceptanceRate] = float64(1.5)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "acceptance_rate harus antara 0 dan 1")
}

func TestValuasiDescriptor_Validate_AcceptanceRateNegative(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	data := validJVData()
	data[jaminan.JVFieldAcceptanceRate] = float64(-0.1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "acceptance_rate harus antara 0 dan 1")
}

func TestValuasiDescriptor_Validate_NegativeCollateralValue(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	data := validJVData()
	data[jaminan.JVFieldCollateralVal] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "collateral_value tidak boleh negatif")
}

func TestValuasiDescriptor_Validate_AcceptanceRateBoundary(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}

	// rate = 0 valid
	data := validJVData()
	data[jaminan.JVFieldAcceptanceRate] = float64(0)
	assert.NoError(t, d.Validate(data))

	// rate = 1 valid
	data = validJVData()
	data[jaminan.JVFieldAcceptanceRate] = float64(1)
	assert.NoError(t, d.Validate(data))
}

func TestValuasiDescriptor_Validate_NilData(t *testing.T) {
	d := &jaminan.ValuasiDescriptor{}
	err := d.Validate(nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "jaminan_id wajib diisi")
}

// ═══════════════════════════════════════════════════════════════════════════════
// CONSTANT CORRECTNESS
// ═══════════════════════════════════════════════════════════════════════════════

func TestConstants_JaminanNotEmpty(t *testing.T) {
	assert.NotEmpty(t, jaminan.CollateralInternalBalance)
	assert.NotEmpty(t, jaminan.CollateralSuratBerharga)
	assert.NotEmpty(t, jaminan.CollateralBarangBergerak)
	assert.NotEmpty(t, jaminan.CollateralPersonalGuarantee)
	assert.NotEmpty(t, jaminan.StatusRegistered)
	assert.NotEmpty(t, jaminan.StatusPledged)
	assert.NotEmpty(t, jaminan.StatusReleased)
	assert.NotEmpty(t, jaminan.StatusForeclosed)
}

func TestConstants_PinjamanNotEmpty(t *testing.T) {
	assert.NotEmpty(t, jaminan.JPStatusActive)
	assert.NotEmpty(t, jaminan.JPStatusReleased)
}

func TestConstants_DokumenNotEmpty(t *testing.T) {
	assert.NotEmpty(t, jaminan.DocFotoBarang)
	assert.NotEmpty(t, jaminan.DocFotokopiBPKB)
	assert.NotEmpty(t, jaminan.DocSertifikatTanah)
	assert.NotEmpty(t, jaminan.DocSuratKuasa)
	assert.NotEmpty(t, jaminan.DocBeritaAcara)
	assert.NotEmpty(t, jaminan.DocFotoIdentitas)
	assert.NotEmpty(t, jaminan.DocSuratPernyataan)
	assert.NotEmpty(t, jaminan.DocIjazah)
	assert.NotEmpty(t, jaminan.DocLainnya)
}

func TestConstants_ValuasiNotEmpty(t *testing.T) {
	assert.NotEmpty(t, jaminan.ValuationInitial)
	assert.NotEmpty(t, jaminan.ValuationPeriodic)
	assert.NotEmpty(t, jaminan.ValuationOnDemand)
}
