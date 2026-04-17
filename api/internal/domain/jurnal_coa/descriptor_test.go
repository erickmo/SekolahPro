package jurnal_coa_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/jurnal_coa"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// COA Descriptor tests
// ============================================================================

func TestCOADescriptor_TableName(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	assert.Equal(t, "coa", d.TableName())
}

func TestCOADescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jurnal_coa.COADescriptor{}
}

func TestCOADescriptor_DefaultRels_Empty(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	rels := d.DefaultRels()
	assert.Empty(t, rels, "coa should have no BelongsTo relations")
}

func validCOAData() map[string]any {
	return map[string]any{
		jurnal_coa.COAFieldAccountCode:   "1101",
		jurnal_coa.COAFieldAccountName:   "Kas Teller",
		jurnal_coa.COAFieldLevel:         float64(2),
		jurnal_coa.COAFieldAccountType:   jurnal_coa.AccountTypeAsset,
		jurnal_coa.COAFieldNormalBalance: jurnal_coa.NormalBalanceDebit,
	}
}

func TestCOADescriptor_Validate_Success(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	err := d.Validate(validCOAData())
	assert.NoError(t, err)
}

func TestCOADescriptor_Validate_MissingAccountCode(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	delete(data, jurnal_coa.COAFieldAccountCode)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "account_code wajib diisi")
}

func TestCOADescriptor_Validate_MissingAccountName(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	delete(data, jurnal_coa.COAFieldAccountName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "account_name wajib diisi")
}

func TestCOADescriptor_Validate_MissingLevel(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	delete(data, jurnal_coa.COAFieldLevel)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "level wajib diisi")
}

func TestCOADescriptor_Validate_MissingAccountType(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	delete(data, jurnal_coa.COAFieldAccountType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "account_type wajib diisi")
}

func TestCOADescriptor_Validate_MissingNormalBalance(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	delete(data, jurnal_coa.COAFieldNormalBalance)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "normal_balance wajib diisi")
}

func TestCOADescriptor_Validate_InvalidAccountType(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	data[jurnal_coa.COAFieldAccountType] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "account_type tidak valid")
}

func TestCOADescriptor_Validate_InvalidNormalBalance(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	data[jurnal_coa.COAFieldNormalBalance] = "mixed"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "normal_balance tidak valid")
}

func TestCOADescriptor_Validate_InvalidCoopType(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	data := validCOAData()
	data[jurnal_coa.COAFieldCoopTypeRequired] = "mixed"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "coop_type_required tidak valid")
}

func TestCOADescriptor_Validate_AllAccountTypes(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	types := []string{
		jurnal_coa.AccountTypeAsset, jurnal_coa.AccountTypeLiability,
		jurnal_coa.AccountTypeEquity, jurnal_coa.AccountTypeRevenue,
		jurnal_coa.AccountTypeExpense, jurnal_coa.AccountTypeZakat,
		jurnal_coa.AccountTypeKebajikan, jurnal_coa.AccountTypeTazir,
	}
	for _, at := range types {
		data := validCOAData()
		data[jurnal_coa.COAFieldAccountType] = at
		err := d.Validate(data)
		assert.NoError(t, err, "account_type=%q should be valid", at)
	}
}

// ============================================================================
// Jurnal Descriptor tests
// ============================================================================

func TestJurnalDescriptor_TableName(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	assert.Equal(t, "jurnal", d.TableName())
}

func TestJurnalDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jurnal_coa.JurnalDescriptor{}
}

func TestJurnalDescriptor_DefaultRels_Count(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "jurnal should have 2 BelongsTo relations")
}

func TestJurnalDescriptor_DefaultRels_BranchRelation(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jurnal_coa.RelBranch]
	require.True(t, ok, "should have branch relation")

	assert.Equal(t, "branches", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, jurnal_coa.JurnalFieldBranchID, rel.FK)
	assert.True(t, rel.IsAutoload, "branch should be autoloaded")
}

func TestJurnalDescriptor_DefaultRels_AccountingPeriodRelation(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[jurnal_coa.RelAccountingPeriod]
	require.True(t, ok, "should have accounting_period relation")

	assert.Equal(t, "accounting_period", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, jurnal_coa.JurnalFieldPeriodID, rel.FK)
	assert.True(t, rel.IsAutoload, "accounting_period should be autoloaded")
}

func validJurnalData() map[string]any {
	return map[string]any{
		jurnal_coa.JurnalFieldJournalNumber: "JRN-2026-01-00000001",
		jurnal_coa.JurnalFieldJournalDate:   "2026-01-15",
		jurnal_coa.JurnalFieldDescription:   "Setoran tunai tabungan",
		jurnal_coa.JurnalFieldSourceType:    jurnal_coa.SourceTypeTransaction,
		jurnal_coa.JurnalFieldPeriodID:      "00000000-0000-0000-0000-000000000001",
		jurnal_coa.JurnalFieldTotalDebit:    float64(100000),
		jurnal_coa.JurnalFieldTotalCredit:   float64(100000),
	}
}

func TestJurnalDescriptor_Validate_Success(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	err := d.Validate(validJurnalData())
	assert.NoError(t, err)
}

func TestJurnalDescriptor_Validate_MissingJournalNumber(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	delete(data, jurnal_coa.JurnalFieldJournalNumber)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "journal_number wajib diisi")
}

func TestJurnalDescriptor_Validate_MissingJournalDate(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	delete(data, jurnal_coa.JurnalFieldJournalDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "journal_date wajib diisi")
}

func TestJurnalDescriptor_Validate_MissingDescription(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	delete(data, jurnal_coa.JurnalFieldDescription)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "description wajib diisi")
}

func TestJurnalDescriptor_Validate_MissingSourceType(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	delete(data, jurnal_coa.JurnalFieldSourceType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "source_type wajib diisi")
}

func TestJurnalDescriptor_Validate_MissingPeriodID(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	delete(data, jurnal_coa.JurnalFieldPeriodID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_id wajib diisi")
}

func TestJurnalDescriptor_Validate_MissingTotalDebit(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	delete(data, jurnal_coa.JurnalFieldTotalDebit)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_debit wajib diisi")
}

func TestJurnalDescriptor_Validate_NegativeTotalDebit(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	data[jurnal_coa.JurnalFieldTotalDebit] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_debit tidak boleh negatif")
}

func TestJurnalDescriptor_Validate_NegativeTotalCredit(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	data[jurnal_coa.JurnalFieldTotalCredit] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_credit tidak boleh negatif")
}

func TestJurnalDescriptor_Validate_InvalidSourceType(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	data[jurnal_coa.JurnalFieldSourceType] = "import"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "source_type tidak valid")
}

func TestJurnalDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	data := validJurnalData()
	data[jurnal_coa.JurnalFieldStatus] = "cancelled"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestJurnalDescriptor_Validate_AllSourceTypes(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	types := []string{
		jurnal_coa.SourceTypeTransaction, jurnal_coa.SourceTypeManual,
		jurnal_coa.SourceTypeClosing, jurnal_coa.SourceTypeAdjustment,
		jurnal_coa.SourceTypeSHUDistribution,
	}
	for _, st := range types {
		data := validJurnalData()
		data[jurnal_coa.JurnalFieldSourceType] = st
		err := d.Validate(data)
		assert.NoError(t, err, "source_type=%q should be valid", st)
	}
}

// ============================================================================
// JournalMapping Descriptor tests
// ============================================================================

func TestJournalMappingDescriptor_TableName(t *testing.T) {
	d := &jurnal_coa.JournalMappingDescriptor{}
	assert.Equal(t, "journal_mapping", d.TableName())
}

func TestJournalMappingDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jurnal_coa.JournalMappingDescriptor{}
}

func TestJournalMappingDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &jurnal_coa.JournalMappingDescriptor{}
	rels := d.DefaultRels()
	assert.Empty(t, rels, "journal_mapping should have no BelongsTo relations")
}

func validJMapData() map[string]any {
	return map[string]any{
		jurnal_coa.JMapFieldTransactionType: "deposit_cash",
		jurnal_coa.JMapFieldRules: []map[string]any{
			{"line": 1, "coa_code": "1101", "side": "debit"},
			{"line": 2, "coa_code": "2101", "side": "credit"},
		},
	}
}

func TestJournalMappingDescriptor_Validate_Success(t *testing.T) {
	d := &jurnal_coa.JournalMappingDescriptor{}
	err := d.Validate(validJMapData())
	assert.NoError(t, err)
}

func TestJournalMappingDescriptor_Validate_MissingTransactionType(t *testing.T) {
	d := &jurnal_coa.JournalMappingDescriptor{}
	data := validJMapData()
	delete(data, jurnal_coa.JMapFieldTransactionType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "transaction_type wajib diisi")
}

func TestJournalMappingDescriptor_Validate_MissingRules(t *testing.T) {
	d := &jurnal_coa.JournalMappingDescriptor{}
	data := validJMapData()
	delete(data, jurnal_coa.JMapFieldRules)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rules wajib diisi")
}

func TestJournalMappingDescriptor_Validate_InvalidCoopType(t *testing.T) {
	d := &jurnal_coa.JournalMappingDescriptor{}
	data := validJMapData()
	data[jurnal_coa.JMapFieldCoopType] = "mixed"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "coop_type tidak valid")
}

// ============================================================================
// AccountingPeriod Descriptor tests
// ============================================================================

func TestAccountingPeriodDescriptor_TableName(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	assert.Equal(t, "accounting_period", d.TableName())
}

func TestAccountingPeriodDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &jurnal_coa.AccountingPeriodDescriptor{}
}

func TestAccountingPeriodDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	rels := d.DefaultRels()
	assert.Empty(t, rels, "accounting_period should have no BelongsTo relations")
}

func validAPData() map[string]any {
	return map[string]any{
		jurnal_coa.APFieldPeriodType: jurnal_coa.PeriodTypeMonthly,
		jurnal_coa.APFieldYear:       float64(2026),
		jurnal_coa.APFieldStartDate:  "2026-01-01",
		jurnal_coa.APFieldEndDate:    "2026-01-31",
	}
}

func TestAccountingPeriodDescriptor_Validate_Success(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	err := d.Validate(validAPData())
	assert.NoError(t, err)
}

func TestAccountingPeriodDescriptor_Validate_MissingPeriodType(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	data := validAPData()
	delete(data, jurnal_coa.APFieldPeriodType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_type wajib diisi")
}

func TestAccountingPeriodDescriptor_Validate_MissingYear(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	data := validAPData()
	delete(data, jurnal_coa.APFieldYear)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "year wajib diisi")
}

func TestAccountingPeriodDescriptor_Validate_YearTooSmall(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	data := validAPData()
	data[jurnal_coa.APFieldYear] = float64(1999)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "year wajib diisi")
}

func TestAccountingPeriodDescriptor_Validate_MissingStartDate(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	data := validAPData()
	delete(data, jurnal_coa.APFieldStartDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "start_date wajib diisi")
}

func TestAccountingPeriodDescriptor_Validate_MissingEndDate(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	data := validAPData()
	delete(data, jurnal_coa.APFieldEndDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "end_date wajib diisi")
}

func TestAccountingPeriodDescriptor_Validate_InvalidPeriodType(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	data := validAPData()
	data[jurnal_coa.APFieldPeriodType] = "weekly"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_type tidak valid")
}

func TestAccountingPeriodDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	data := validAPData()
	data[jurnal_coa.APFieldStatus] = "pending"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestAccountingPeriodDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &jurnal_coa.AccountingPeriodDescriptor{}
	statuses := []string{
		jurnal_coa.APStatusOpen, jurnal_coa.APStatusClosed,
		jurnal_coa.APStatusLocked,
	}
	for _, status := range statuses {
		data := validAPData()
		data[jurnal_coa.APFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}
