// Package jurnal_coa adalah domain Vernon untuk akuntansi koperasi.
//
// Terdiri dari 4 Vernon descriptor:
//   - COA (Chart of Accounts): standalone, tanpa BelongsTo.
//   - Jurnal (header): BelongsTo branch dan accounting_period.
//   - JournalMapping: standalone, tanpa BelongsTo.
//   - AccountingPeriod: standalone, tanpa BelongsTo.
//
// jurnal_line bukan Vernon domain — hanya regular detail table.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package jurnal_coa

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// COA — Chart of Accounts
// ============================================================================

// COA field name constants.
const (
	COAFieldAccountCode       = "account_code"
	COAFieldAccountName       = "account_name"
	COAFieldParentID          = "parent_id"
	COAFieldLevel             = "level"
	COAFieldAccountType       = "account_type"
	COAFieldNormalBalance     = "normal_balance"
	COAFieldDescription       = "description"
	COAFieldIsSystem          = "is_system"
	COAFieldIsActive          = "is_active"
	COAFieldCoopTypeRequired  = "coop_type_required"
)

// COA account type constants.
const (
	AccountTypeAsset     = "asset"
	AccountTypeLiability = "liability"
	AccountTypeEquity    = "equity"
	AccountTypeRevenue   = "revenue"
	AccountTypeExpense   = "expense"
	AccountTypeZakat     = "zakat"
	AccountTypeKebajikan = "kebajikan"
	AccountTypeTazir     = "tazir"
)

// COA normal balance constants.
const (
	NormalBalanceDebit  = "debit"
	NormalBalanceCredit = "credit"
)

// COA coop type required constants.
const (
	CoopTypeGeneral = "general"
	CoopTypeIslamic = "islamic"
	CoopTypeBoth    = "both"
)

var validAccountTypes = map[string]bool{
	AccountTypeAsset: true, AccountTypeLiability: true,
	AccountTypeEquity: true, AccountTypeRevenue: true,
	AccountTypeExpense: true, AccountTypeZakat: true,
	AccountTypeKebajikan: true, AccountTypeTazir: true,
}

var validNormalBalances = map[string]bool{
	NormalBalanceDebit: true, NormalBalanceCredit: true,
}

var validCoopTypes = map[string]bool{
	CoopTypeGeneral: true, CoopTypeIslamic: true, CoopTypeBoth: true,
}

// COADescriptor mengimplementasi vernon.DomainDescriptor untuk coa.
type COADescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *COADescriptor) TableName() string { return "coa" }

// DefaultRels — COA tidak memiliki BelongsTo (standalone config).
func (d *COADescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain COA sebelum write.
func (d *COADescriptor) Validate(data map[string]any) error {
	if err := validateCOARequired(data); err != nil {
		return err
	}
	return validateCOAEnums(data)
}

func validateCOARequired(data map[string]any) error {
	code, _ := data[COAFieldAccountCode].(string)
	if code == "" {
		return fmt.Errorf("account_code wajib diisi")
	}

	name, _ := data[COAFieldAccountName].(string)
	if name == "" {
		return fmt.Errorf("account_name wajib diisi")
	}

	level, ok := data[COAFieldLevel].(float64)
	if !ok || level < 1 {
		return fmt.Errorf("level wajib diisi dan harus >= 1")
	}

	at, _ := data[COAFieldAccountType].(string)
	if at == "" {
		return fmt.Errorf("account_type wajib diisi")
	}

	nb, _ := data[COAFieldNormalBalance].(string)
	if nb == "" {
		return fmt.Errorf("normal_balance wajib diisi")
	}
	return nil
}

func validateCOAEnums(data map[string]any) error {
	at, _ := data[COAFieldAccountType].(string)
	if at != "" && !validAccountTypes[at] {
		return fmt.Errorf("account_type tidak valid: %q", at)
	}

	nb, _ := data[COAFieldNormalBalance].(string)
	if nb != "" && !validNormalBalances[nb] {
		return fmt.Errorf("normal_balance tidak valid: %q", nb)
	}

	ct, _ := data[COAFieldCoopTypeRequired].(string)
	if ct != "" && !validCoopTypes[ct] {
		return fmt.Errorf("coop_type_required tidak valid: %q", ct)
	}
	return nil
}

// ============================================================================
// Jurnal (header)
// ============================================================================

// Jurnal field name constants.
const (
	JurnalFieldJournalNumber = "journal_number"
	JurnalFieldJournalDate   = "journal_date"
	JurnalFieldDescription   = "description"
	JurnalFieldSourceType    = "source_type"
	JurnalFieldSourceID      = "source_id"
	JurnalFieldReference     = "reference"
	JurnalFieldPeriodID      = "period_id"
	JurnalFieldBranchID      = "branch_id"
	JurnalFieldTotalDebit    = "total_debit"
	JurnalFieldTotalCredit   = "total_credit"
	JurnalFieldStatus        = "status"
	JurnalFieldIsAutoPost    = "is_auto_post"
)

// Jurnal source type constants.
const (
	SourceTypeTransaction    = "transaction"
	SourceTypeManual         = "manual"
	SourceTypeClosing        = "closing"
	SourceTypeAdjustment     = "adjustment"
	SourceTypeSHUDistribution = "shu_distribution"
)

// Jurnal status constants.
const (
	JurnalStatusUnposted = "unposted"
	JurnalStatusPosted   = "posted"
	JurnalStatusReversed = "reversed"
)

// Jurnal relation name constants.
const (
	RelBranch            = "branch"
	RelAccountingPeriod  = "accounting_period"
)

var validSourceTypes = map[string]bool{
	SourceTypeTransaction: true, SourceTypeManual: true,
	SourceTypeClosing: true, SourceTypeAdjustment: true,
	SourceTypeSHUDistribution: true,
}

var validJurnalStatuses = map[string]bool{
	JurnalStatusUnposted: true, JurnalStatusPosted: true,
	JurnalStatusReversed: true,
}

// JurnalDescriptor mengimplementasi vernon.DomainDescriptor untuk jurnal.
type JurnalDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *JurnalDescriptor) TableName() string { return "jurnal" }

// DefaultRels mendefinisikan relasi domain jurnal.
// BelongsTo branch dan accounting_period (keduanya autoload).
func (d *JurnalDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelBranch: {
			Domain:     "branches",
			Type:       vernon.RelBelongsTo,
			FK:         JurnalFieldBranchID,
			LocalKey:   JurnalFieldBranchID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
		RelAccountingPeriod: {
			Domain:     "accounting_period",
			Type:       vernon.RelBelongsTo,
			FK:         JurnalFieldPeriodID,
			LocalKey:   JurnalFieldPeriodID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"period_name", "year", "month"},
		},
	}
}

// Validate memvalidasi invariant domain jurnal sebelum write.
func (d *JurnalDescriptor) Validate(data map[string]any) error {
	if err := validateJurnalRequired(data); err != nil {
		return err
	}
	return validateJurnalEnums(data)
}

func validateJurnalRequired(data map[string]any) error {
	jn, _ := data[JurnalFieldJournalNumber].(string)
	if jn == "" {
		return fmt.Errorf("journal_number wajib diisi")
	}

	jd, _ := data[JurnalFieldJournalDate].(string)
	if jd == "" {
		return fmt.Errorf("journal_date wajib diisi")
	}

	desc, _ := data[JurnalFieldDescription].(string)
	if desc == "" {
		return fmt.Errorf("description wajib diisi")
	}

	st, _ := data[JurnalFieldSourceType].(string)
	if st == "" {
		return fmt.Errorf("source_type wajib diisi")
	}

	pi, _ := data[JurnalFieldPeriodID].(string)
	if pi == "" {
		return fmt.Errorf("period_id wajib diisi")
	}

	td, ok := data[JurnalFieldTotalDebit].(float64)
	if !ok {
		return fmt.Errorf("total_debit wajib diisi dan harus berupa angka")
	}
	if td < 0 {
		return fmt.Errorf("total_debit tidak boleh negatif")
	}

	tc, ok := data[JurnalFieldTotalCredit].(float64)
	if !ok {
		return fmt.Errorf("total_credit wajib diisi dan harus berupa angka")
	}
	if tc < 0 {
		return fmt.Errorf("total_credit tidak boleh negatif")
	}
	return nil
}

func validateJurnalEnums(data map[string]any) error {
	st, _ := data[JurnalFieldSourceType].(string)
	if st != "" && !validSourceTypes[st] {
		return fmt.Errorf("source_type tidak valid: %q", st)
	}

	status, _ := data[JurnalFieldStatus].(string)
	if status != "" && !validJurnalStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}

// ============================================================================
// JournalMapping
// ============================================================================

// JournalMapping field name constants.
const (
	JMapFieldTransactionType = "transaction_type"
	JMapFieldCoopType        = "coop_type"
	JMapFieldRules           = "rules"
	JMapFieldDescription     = "description"
	JMapFieldIsActive        = "is_active"
)

// JournalMappingDescriptor mengimplementasi vernon.DomainDescriptor untuk journal_mapping.
type JournalMappingDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *JournalMappingDescriptor) TableName() string { return "journal_mapping" }

// DefaultRels — JournalMapping tidak memiliki BelongsTo (standalone config).
func (d *JournalMappingDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain journal_mapping sebelum write.
func (d *JournalMappingDescriptor) Validate(data map[string]any) error {
	tt, _ := data[JMapFieldTransactionType].(string)
	if tt == "" {
		return fmt.Errorf("transaction_type wajib diisi")
	}

	rules, ok := data[JMapFieldRules]
	if !ok || rules == nil {
		return fmt.Errorf("rules wajib diisi")
	}

	ct, _ := data[JMapFieldCoopType].(string)
	if ct != "" && !validCoopTypes[ct] {
		return fmt.Errorf("coop_type tidak valid: %q", ct)
	}
	return nil
}

// ============================================================================
// AccountingPeriod
// ============================================================================

// AccountingPeriod field name constants.
const (
	APFieldPeriodType = "period_type"
	APFieldYear       = "year"
	APFieldMonth      = "month"
	APFieldPeriodName = "period_name"
	APFieldStartDate  = "start_date"
	APFieldEndDate    = "end_date"
	APFieldStatus     = "status"
)

// AccountingPeriod period type constants.
const (
	PeriodTypeMonthly = "monthly"
	PeriodTypeAnnual  = "annual"
)

// AccountingPeriod status constants.
const (
	APStatusOpen   = "open"
	APStatusClosed = "closed"
	APStatusLocked = "locked"
)

var validPeriodTypes = map[string]bool{
	PeriodTypeMonthly: true, PeriodTypeAnnual: true,
}

var validAPStatuses = map[string]bool{
	APStatusOpen: true, APStatusClosed: true, APStatusLocked: true,
}

// AccountingPeriodDescriptor mengimplementasi vernon.DomainDescriptor untuk accounting_period.
type AccountingPeriodDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *AccountingPeriodDescriptor) TableName() string { return "accounting_period" }

// DefaultRels — AccountingPeriod tidak memiliki BelongsTo (standalone config).
func (d *AccountingPeriodDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain accounting_period sebelum write.
func (d *AccountingPeriodDescriptor) Validate(data map[string]any) error {
	if err := validateAPRequired(data); err != nil {
		return err
	}
	return validateAPEnums(data)
}

func validateAPRequired(data map[string]any) error {
	pt, _ := data[APFieldPeriodType].(string)
	if pt == "" {
		return fmt.Errorf("period_type wajib diisi")
	}

	year, ok := data[APFieldYear].(float64)
	if !ok || year < 2000 {
		return fmt.Errorf("year wajib diisi dan harus >= 2000")
	}

	sd, _ := data[APFieldStartDate].(string)
	if sd == "" {
		return fmt.Errorf("start_date wajib diisi")
	}

	ed, _ := data[APFieldEndDate].(string)
	if ed == "" {
		return fmt.Errorf("end_date wajib diisi")
	}
	return nil
}

func validateAPEnums(data map[string]any) error {
	pt, _ := data[APFieldPeriodType].(string)
	if pt != "" && !validPeriodTypes[pt] {
		return fmt.Errorf("period_type tidak valid: %q", pt)
	}

	status, _ := data[APFieldStatus].(string)
	if status != "" && !validAPStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
