package shu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/shu"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// SHUPeriode Descriptor tests
// ============================================================================

func TestSHUPeriodeDescriptor_TableName(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	assert.Equal(t, "shu_periode", d.TableName())
}

func TestSHUPeriodeDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &shu.SHUPeriodeDescriptor{}
}

func TestSHUPeriodeDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	rels := d.DefaultRels()
	assert.Empty(t, rels, "shu_periode should have no BelongsTo relations")
}

func validSPData() map[string]any {
	return map[string]any{
		shu.SPFieldTahunBuku:       float64(2026),
		shu.SPFieldPeriodStart:     "2026-01-01",
		shu.SPFieldPeriodEnd:       "2026-12-31",
		shu.SPFieldTotalPendapatan: float64(50000000),
		shu.SPFieldTotalBeban:      float64(30000000),
		shu.SPFieldSHUBruto:        float64(20000000),
		shu.SPFieldSHUNeto:         float64(20000000),
	}
}

func TestSHUPeriodeDescriptor_Validate_Success(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	err := d.Validate(validSPData())
	assert.NoError(t, err)
}

func TestSHUPeriodeDescriptor_Validate_MissingTahunBuku(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	delete(data, shu.SPFieldTahunBuku)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tahun_buku wajib diisi")
}

func TestSHUPeriodeDescriptor_Validate_TahunBukuTooSmall(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	data[shu.SPFieldTahunBuku] = float64(1999)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tahun_buku wajib diisi")
}

func TestSHUPeriodeDescriptor_Validate_MissingPeriodStart(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	delete(data, shu.SPFieldPeriodStart)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_start wajib diisi")
}

func TestSHUPeriodeDescriptor_Validate_MissingPeriodEnd(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	delete(data, shu.SPFieldPeriodEnd)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_end wajib diisi")
}

func TestSHUPeriodeDescriptor_Validate_MissingTotalPendapatan(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	delete(data, shu.SPFieldTotalPendapatan)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_pendapatan wajib diisi")
}

func TestSHUPeriodeDescriptor_Validate_NegativeTotalPendapatan(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	data[shu.SPFieldTotalPendapatan] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_pendapatan tidak boleh negatif")
}

func TestSHUPeriodeDescriptor_Validate_NegativeTotalBeban(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	data[shu.SPFieldTotalBeban] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_beban tidak boleh negatif")
}

func TestSHUPeriodeDescriptor_Validate_NegativeSHUBruto(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	data[shu.SPFieldSHUBruto] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "shu_bruto tidak boleh negatif")
}

func TestSHUPeriodeDescriptor_Validate_NegativeSHUNeto(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	data[shu.SPFieldSHUNeto] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "shu_neto tidak boleh negatif")
}

func TestSHUPeriodeDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	data := validSPData()
	data[shu.SPFieldStatus] = "pending"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestSHUPeriodeDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	statuses := []string{
		shu.SPStatusCalculated, shu.SPStatusReviewed,
		shu.SPStatusApproved, shu.SPStatusDistributed,
	}
	for _, status := range statuses {
		data := validSPData()
		data[shu.SPFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

// ============================================================================
// SHUAnggota Descriptor tests
// ============================================================================

func TestSHUAnggotaDescriptor_TableName(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	assert.Equal(t, "shu_anggota", d.TableName())
}

func TestSHUAnggotaDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &shu.SHUAnggotaDescriptor{}
}

func TestSHUAnggotaDescriptor_DefaultRels_Count(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "shu_anggota should have 3 BelongsTo relations")
}

func TestSHUAnggotaDescriptor_DefaultRels_SHUPeriodeRelation(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[shu.RelSHUPeriode]
	require.True(t, ok, "should have shu_periode relation")

	assert.Equal(t, "shu_periode", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, shu.SAFieldSHUPeriodeID, rel.FK)
	assert.True(t, rel.IsAutoload, "shu_periode should be autoloaded")
}

func TestSHUAnggotaDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[shu.RelNasabah]
	require.True(t, ok, "should have nasabah relation")

	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, shu.SAFieldNasabahID, rel.FK)
	assert.True(t, rel.IsAutoload, "nasabah should be autoloaded")
}

func TestSHUAnggotaDescriptor_DefaultRels_TargetRekeningRelation(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[shu.RelTargetRekening]
	require.True(t, ok, "should have target_rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, shu.SAFieldTargetRekeningID, rel.FK)
	assert.True(t, rel.IsAutoload, "target_rekening should be autoloaded")
}

func validSAData() map[string]any {
	return map[string]any{
		shu.SAFieldSHUPeriodeID:   "00000000-0000-0000-0000-000000000001",
		shu.SAFieldNasabahID:      "00000000-0000-0000-0000-000000000002",
		shu.SAFieldAvgSimpanan:    float64(5000000),
		shu.SAFieldTotalTransaksi: float64(10000000),
		shu.SAFieldActiveDays:     float64(365),
		shu.SAFieldJasaModal:      float64(500000),
		shu.SAFieldJasaUsaha:      float64(300000),
		shu.SAFieldTotalSHU:       float64(800000),
	}
}

func TestSHUAnggotaDescriptor_Validate_Success(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	err := d.Validate(validSAData())
	assert.NoError(t, err)
}

func TestSHUAnggotaDescriptor_Validate_MissingSHUPeriodeID(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	delete(data, shu.SAFieldSHUPeriodeID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "shu_periode_id wajib diisi")
}

func TestSHUAnggotaDescriptor_Validate_MissingNasabahID(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	delete(data, shu.SAFieldNasabahID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestSHUAnggotaDescriptor_Validate_MissingAvgSimpanan(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	delete(data, shu.SAFieldAvgSimpanan)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "avg_simpanan wajib diisi")
}

func TestSHUAnggotaDescriptor_Validate_NegativeAvgSimpanan(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	data[shu.SAFieldAvgSimpanan] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "avg_simpanan tidak boleh negatif")
}

func TestSHUAnggotaDescriptor_Validate_NegativeTotalTransaksi(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	data[shu.SAFieldTotalTransaksi] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_transaksi tidak boleh negatif")
}

func TestSHUAnggotaDescriptor_Validate_NegativeActiveDays(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	data[shu.SAFieldActiveDays] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "active_days tidak boleh negatif")
}

func TestSHUAnggotaDescriptor_Validate_NegativeJasaModal(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	data[shu.SAFieldJasaModal] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "jasa_modal tidak boleh negatif")
}

func TestSHUAnggotaDescriptor_Validate_NegativeJasaUsaha(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	data[shu.SAFieldJasaUsaha] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "jasa_usaha tidak boleh negatif")
}

func TestSHUAnggotaDescriptor_Validate_NegativeTotalSHU(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	data[shu.SAFieldTotalSHU] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_shu tidak boleh negatif")
}

func TestSHUAnggotaDescriptor_Validate_InvalidDistMethod(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	data := validSAData()
	data[shu.SAFieldDistMethod] = "cheque"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "distribution_method tidak valid")
}

func TestSHUAnggotaDescriptor_Validate_AllDistMethods(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	methods := []string{
		shu.DistMethodCreditTabungan, shu.DistMethodSeparatePayout,
		shu.DistMethodPending,
	}
	for _, m := range methods {
		data := validSAData()
		data[shu.SAFieldDistMethod] = m
		err := d.Validate(data)
		assert.NoError(t, err, "distribution_method=%q should be valid", m)
	}
}
