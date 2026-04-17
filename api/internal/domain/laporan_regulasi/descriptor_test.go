package laporan_regulasi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/laporan_regulasi"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── ConfigDescriptor metadata tests ─────────────────────────────────────────

func TestConfigDescriptor_TableName(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	assert.Equal(t, "laporan_config", d.TableName())
}

func TestConfigDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &laporan_regulasi.ConfigDescriptor{}
}

func TestConfigDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	rels := d.DefaultRels()
	assert.Empty(t, rels, "laporan_config should have no relations")
}

// ── ConfigDescriptor validation tests ────────────────────────────────────────

func validConfigData() map[string]any {
	return map[string]any{
		laporan_regulasi.CfgFieldReportType:     "neraca",
		laporan_regulasi.CfgFieldReportName:     "Neraca / Balance Sheet",
		laporan_regulasi.CfgFieldReportCategory: laporan_regulasi.CategoryDinasKoperasi,
		laporan_regulasi.CfgFieldFrequency:      laporan_regulasi.FrequencyAnnually,
	}
}

func TestConfigDescriptor_Validate_Success(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	err := d.Validate(validConfigData())
	assert.NoError(t, err)
}

func TestConfigDescriptor_Validate_MissingReportType(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	data := validConfigData()
	delete(data, laporan_regulasi.CfgFieldReportType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "report_type wajib diisi")
}

func TestConfigDescriptor_Validate_MissingReportName(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	data := validConfigData()
	delete(data, laporan_regulasi.CfgFieldReportName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "report_name wajib diisi")
}

func TestConfigDescriptor_Validate_MissingReportCategory(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	data := validConfigData()
	delete(data, laporan_regulasi.CfgFieldReportCategory)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "report_category wajib diisi")
}

func TestConfigDescriptor_Validate_InvalidReportCategory(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	data := validConfigData()
	data[laporan_regulasi.CfgFieldReportCategory] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "report_category tidak valid")
}

func TestConfigDescriptor_Validate_MissingFrequency(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	data := validConfigData()
	delete(data, laporan_regulasi.CfgFieldFrequency)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "frequency wajib diisi")
}

func TestConfigDescriptor_Validate_InvalidFrequency(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	data := validConfigData()
	data[laporan_regulasi.CfgFieldFrequency] = "hourly"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "frequency tidak valid")
}

func TestConfigDescriptor_Validate_AllCategories(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	categories := []string{
		laporan_regulasi.CategoryDinasKoperasi,
		laporan_regulasi.CategoryOJK,
		laporan_regulasi.CategoryIslamic,
		laporan_regulasi.CategoryInternal,
	}
	for _, cat := range categories {
		data := validConfigData()
		data[laporan_regulasi.CfgFieldReportCategory] = cat
		err := d.Validate(data)
		assert.NoError(t, err, "category=%q should be valid", cat)
	}
}

func TestConfigDescriptor_Validate_AllFrequencies(t *testing.T) {
	d := &laporan_regulasi.ConfigDescriptor{}
	frequencies := []string{
		laporan_regulasi.FrequencyDaily,
		laporan_regulasi.FrequencyWeekly,
		laporan_regulasi.FrequencyMonthly,
		laporan_regulasi.FrequencyQuarterly,
		laporan_regulasi.FrequencyAnnually,
	}
	for _, freq := range frequencies {
		data := validConfigData()
		data[laporan_regulasi.CfgFieldFrequency] = freq
		err := d.Validate(data)
		assert.NoError(t, err, "frequency=%q should be valid", freq)
	}
}

// ── LaporanDescriptor metadata tests ────────────────────────────────────────

func TestLaporanDescriptor_TableName(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	assert.Equal(t, "laporan", d.TableName())
}

func TestLaporanDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &laporan_regulasi.LaporanDescriptor{}
}

func TestLaporanDescriptor_DefaultRels_Count(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "laporan should have 2 BelongsTo relations")
}

func TestLaporanDescriptor_DefaultRels_BranchRelation(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[laporan_regulasi.RelBranch]
	require.True(t, ok, "should have branch relation")
	assert.Equal(t, "branches", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, laporan_regulasi.LapFieldBranchID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

func TestLaporanDescriptor_DefaultRels_ConfigRelation(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[laporan_regulasi.RelLaporanConfig]
	require.True(t, ok, "should have laporan_config relation")
	assert.Equal(t, "laporan_config", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, laporan_regulasi.LapFieldLaporanConfigID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

// ── LaporanDescriptor validation tests ───────────────────────────────────────

func validLaporanData() map[string]any {
	return map[string]any{
		laporan_regulasi.LapFieldLaporanConfigID:    "00000000-0000-0000-0000-000000000001",
		laporan_regulasi.LapFieldReportType:         "neraca",
		laporan_regulasi.LapFieldReportCategory:     laporan_regulasi.CategoryDinasKoperasi,
		laporan_regulasi.LapFieldPeriodType:         laporan_regulasi.FrequencyAnnually,
		laporan_regulasi.LapFieldPeriodStart:        "2026-01-01",
		laporan_regulasi.LapFieldPeriodEnd:          "2026-12-31",
		laporan_regulasi.LapFieldPeriodLabel:        "Tahun 2026",
		laporan_regulasi.LapFieldReportData:         map[string]any{},
		laporan_regulasi.LapFieldGeneratedAt:        "2026-01-15T00:00:00Z",
		laporan_regulasi.LapFieldGeneratedBy:        "00000000-0000-0000-0000-000000000099",
		laporan_regulasi.LapFieldStatus:             laporan_regulasi.LapStatusDraft,
		laporan_regulasi.LapFieldConsolidationLevel: laporan_regulasi.ConsolidationCompany,
	}
}

func TestLaporanDescriptor_Validate_Success(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	err := d.Validate(validLaporanData())
	assert.NoError(t, err)
}

func TestLaporanDescriptor_Validate_MissingConfigID(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	data := validLaporanData()
	delete(data, laporan_regulasi.LapFieldLaporanConfigID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "laporan_config_id wajib diisi")
}

func TestLaporanDescriptor_Validate_MissingReportType(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	data := validLaporanData()
	delete(data, laporan_regulasi.LapFieldReportType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "report_type wajib diisi")
}

func TestLaporanDescriptor_Validate_MissingPeriodStart(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	data := validLaporanData()
	delete(data, laporan_regulasi.LapFieldPeriodStart)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_start wajib diisi")
}

func TestLaporanDescriptor_Validate_MissingReportData(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	data := validLaporanData()
	delete(data, laporan_regulasi.LapFieldReportData)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "report_data wajib diisi")
}

func TestLaporanDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	data := validLaporanData()
	data[laporan_regulasi.LapFieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestLaporanDescriptor_Validate_InvalidPeriodType(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	data := validLaporanData()
	data[laporan_regulasi.LapFieldPeriodType] = "hourly"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_type tidak valid")
}

func TestLaporanDescriptor_Validate_InvalidConsolidationLevel(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	data := validLaporanData()
	data[laporan_regulasi.LapFieldConsolidationLevel] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "consolidation_level tidak valid")
}

func TestLaporanDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &laporan_regulasi.LaporanDescriptor{}
	statuses := []string{
		laporan_regulasi.LapStatusDraft,
		laporan_regulasi.LapStatusReviewed,
		laporan_regulasi.LapStatusFinal,
		laporan_regulasi.LapStatusSubmitted,
		laporan_regulasi.LapStatusOverdue,
	}
	for _, status := range statuses {
		data := validLaporanData()
		data[laporan_regulasi.LapFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

// ── VersiDescriptor metadata tests ──────────────────────────────────────────

func TestVersiDescriptor_TableName(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	assert.Equal(t, "laporan_versi", d.TableName())
}

func TestVersiDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &laporan_regulasi.VersiDescriptor{}
}

func TestVersiDescriptor_DefaultRels_Count(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "laporan_versi should have 1 BelongsTo relation")
}

func TestVersiDescriptor_DefaultRels_LaporanRelation(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[laporan_regulasi.RelLaporan]
	require.True(t, ok, "should have laporan relation")
	assert.Equal(t, "laporan", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, laporan_regulasi.VersiFieldLaporanID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

// ── VersiDescriptor validation tests ────────────────────────────────────────

func validVersiData() map[string]any {
	return map[string]any{
		laporan_regulasi.VersiFieldLaporanID:     "00000000-0000-0000-0000-000000000001",
		laporan_regulasi.VersiFieldVersionNo:     float64(1),
		laporan_regulasi.VersiFieldReportData:    map[string]any{},
		laporan_regulasi.VersiFieldCreatedReason: laporan_regulasi.ReasonAutoGenerate,
	}
}

func TestVersiDescriptor_Validate_Success(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	err := d.Validate(validVersiData())
	assert.NoError(t, err)
}

func TestVersiDescriptor_Validate_MissingLaporanID(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	data := validVersiData()
	delete(data, laporan_regulasi.VersiFieldLaporanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "laporan_id wajib diisi")
}

func TestVersiDescriptor_Validate_MissingVersionNo(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	data := validVersiData()
	delete(data, laporan_regulasi.VersiFieldVersionNo)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "version_number wajib diisi")
}

func TestVersiDescriptor_Validate_ZeroVersionNo(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	data := validVersiData()
	data[laporan_regulasi.VersiFieldVersionNo] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "version_number wajib diisi")
}

func TestVersiDescriptor_Validate_MissingReportData(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	data := validVersiData()
	delete(data, laporan_regulasi.VersiFieldReportData)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "report_data wajib diisi")
}

func TestVersiDescriptor_Validate_MissingCreatedReason(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	data := validVersiData()
	delete(data, laporan_regulasi.VersiFieldCreatedReason)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "created_reason wajib diisi")
}

func TestVersiDescriptor_Validate_InvalidCreatedReason(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	data := validVersiData()
	data[laporan_regulasi.VersiFieldCreatedReason] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "created_reason tidak valid")
}

func TestVersiDescriptor_Validate_AllReasons(t *testing.T) {
	d := &laporan_regulasi.VersiDescriptor{}
	reasons := []string{
		laporan_regulasi.ReasonAutoGenerate,
		laporan_regulasi.ReasonManualTrigger,
		laporan_regulasi.ReasonRegenerate,
		laporan_regulasi.ReasonRevision,
	}
	for _, reason := range reasons {
		data := validVersiData()
		data[laporan_regulasi.VersiFieldCreatedReason] = reason
		err := d.Validate(data)
		assert.NoError(t, err, "reason=%q should be valid", reason)
	}
}
