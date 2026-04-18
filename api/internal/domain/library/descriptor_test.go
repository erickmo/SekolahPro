package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/library"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── BookDescriptor metadata tests ──────────────────────────────────────────────

func TestBookDescriptor_TableName(t *testing.T) {
	d := &library.BookDescriptor{}
	assert.Equal(t, "library_books", d.TableName())
}

func TestBookDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &library.BookDescriptor{}
}

func TestBookDescriptor_DefaultRels_Count(t *testing.T) {
	d := &library.BookDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 0, "library_books should have no relations")
}

// ── BookDescriptor validation tests ────────────────────────────────────────────

func validBookData() map[string]any {
	return map[string]any{
		library.FieldTitle:    "Fisika Dasar Jilid 1",
		library.FieldCategory: library.CatTextbook,
	}
}

func TestBookDescriptor_Validate_Success(t *testing.T) {
	d := &library.BookDescriptor{}
	err := d.Validate(validBookData())
	assert.NoError(t, err)
}

func TestBookDescriptor_Validate_AllCategories(t *testing.T) {
	d := &library.BookDescriptor{}
	categories := []string{
		library.CatFiction, library.CatNonfiction, library.CatTextbook,
		library.CatReference, library.CatKitabKuning, library.CatAlQuran,
		library.CatHadits, library.CatFiqh, library.CatAqidah,
		library.CatScience, library.CatLiterature, library.CatOther,
	}
	for _, cat := range categories {
		data := validBookData()
		data[library.FieldCategory] = cat
		err := d.Validate(data)
		assert.NoError(t, err, "category=%q should be valid", cat)
	}
}

func TestBookDescriptor_Validate_MissingTitle(t *testing.T) {
	d := &library.BookDescriptor{}
	data := validBookData()
	delete(data, library.FieldTitle)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "title wajib diisi")
}

func TestBookDescriptor_Validate_MissingCategory(t *testing.T) {
	d := &library.BookDescriptor{}
	data := validBookData()
	delete(data, library.FieldCategory)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "category wajib diisi")
}

func TestBookDescriptor_Validate_InvalidCategory(t *testing.T) {
	d := &library.BookDescriptor{}
	data := validBookData()
	data[library.FieldCategory] = "encyclopedia"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "category tidak valid")
}

// ── CopyDescriptor metadata tests ─────────────────────────────────────────────

func TestCopyDescriptor_TableName(t *testing.T) {
	d := &library.CopyDescriptor{}
	assert.Equal(t, "library_copies", d.TableName())
}

func TestCopyDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &library.CopyDescriptor{}
}

func TestCopyDescriptor_DefaultRels_Count(t *testing.T) {
	d := &library.CopyDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "library_copies should have 1 BelongsTo relation")
}

func TestCopyDescriptor_DefaultRels_BookRelation(t *testing.T) {
	d := &library.CopyDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[library.RelBook]
	require.True(t, ok, "should have book relation")

	assert.Equal(t, "library_books", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, library.FieldBookID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"title", "isbn"}, rel.Fields)
}

// ── CopyDescriptor validation tests ───────────────────────────────────────────

func validCopyData() map[string]any {
	return map[string]any{
		library.FieldBookID:     "00000000-0000-0000-0000-000000000001",
		library.FieldCopyNumber: "1",
		library.FieldCondition:  library.CondGood,
	}
}

func TestCopyDescriptor_Validate_Success(t *testing.T) {
	d := &library.CopyDescriptor{}
	err := d.Validate(validCopyData())
	assert.NoError(t, err)
}

func TestCopyDescriptor_Validate_AllConditions(t *testing.T) {
	d := &library.CopyDescriptor{}
	conditions := []string{
		library.CondNew, library.CondGood, library.CondFair,
		library.CondDamaged, library.CondLost,
	}
	for _, c := range conditions {
		data := validCopyData()
		data[library.FieldCondition] = c
		err := d.Validate(data)
		assert.NoError(t, err, "condition=%q should be valid", c)
	}
}

func TestCopyDescriptor_Validate_MissingBookID(t *testing.T) {
	d := &library.CopyDescriptor{}
	data := validCopyData()
	delete(data, library.FieldBookID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "book_id wajib diisi")
}

func TestCopyDescriptor_Validate_MissingCopyNumber(t *testing.T) {
	d := &library.CopyDescriptor{}
	data := validCopyData()
	delete(data, library.FieldCopyNumber)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "copy_number wajib diisi")
}

func TestCopyDescriptor_Validate_MissingCondition(t *testing.T) {
	d := &library.CopyDescriptor{}
	data := validCopyData()
	delete(data, library.FieldCondition)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "condition wajib diisi")
}

func TestCopyDescriptor_Validate_InvalidCondition(t *testing.T) {
	d := &library.CopyDescriptor{}
	data := validCopyData()
	data[library.FieldCondition] = "worn"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "condition tidak valid")
}

// ── BorrowDescriptor metadata tests ───────────────────────────────────────────

func TestBorrowDescriptor_TableName(t *testing.T) {
	d := &library.BorrowDescriptor{}
	assert.Equal(t, "library_borrows", d.TableName())
}

func TestBorrowDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &library.BorrowDescriptor{}
}

func TestBorrowDescriptor_DefaultRels_Count(t *testing.T) {
	d := &library.BorrowDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "library_borrows should have 3 BelongsTo relations")
}

func TestBorrowDescriptor_DefaultRels_CopyRelation(t *testing.T) {
	d := &library.BorrowDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[library.RelCopy]
	require.True(t, ok, "should have copy relation")

	assert.Equal(t, "library_copies", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, library.FieldCopyID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"barcode", "condition"}, rel.Fields)
}

func TestBorrowDescriptor_DefaultRels_BookRelation(t *testing.T) {
	d := &library.BorrowDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[library.RelBook]
	require.True(t, ok, "should have book relation")

	assert.Equal(t, "library_books", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, "book_id", rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"title", "isbn"}, rel.Fields)
}

func TestBorrowDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &library.BorrowDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[library.RelAcademicYear]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, library.FieldAcademicYearID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name"}, rel.Fields)
}

// ── BorrowDescriptor validation tests ─────────────────────────────────────────

func validBorrowData() map[string]any {
	return map[string]any{
		library.FieldCopyID:         "00000000-0000-0000-0000-000000000001",
		"book_id":                   "00000000-0000-0000-0000-000000000002",
		library.FieldBorrowerType:   library.BorrowerStudent,
		library.FieldBorrowerID:     "00000000-0000-0000-0000-000000000003",
		library.FieldAcademicYearID: "00000000-0000-0000-0000-000000000004",
		library.FieldBorrowDate:     "2026-04-01",
	}
}

func TestBorrowDescriptor_Validate_Success(t *testing.T) {
	d := &library.BorrowDescriptor{}
	err := d.Validate(validBorrowData())
	assert.NoError(t, err)
}

func TestBorrowDescriptor_Validate_AllBorrowerTypes(t *testing.T) {
	d := &library.BorrowDescriptor{}
	types := []string{library.BorrowerStudent, library.BorrowerTeacher}
	for _, bt := range types {
		data := validBorrowData()
		data[library.FieldBorrowerType] = bt
		err := d.Validate(data)
		assert.NoError(t, err, "borrower_type=%q should be valid", bt)
	}
}

func TestBorrowDescriptor_Validate_MissingCopyID(t *testing.T) {
	d := &library.BorrowDescriptor{}
	data := validBorrowData()
	delete(data, library.FieldCopyID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "copy_id wajib diisi")
}

func TestBorrowDescriptor_Validate_MissingBookID(t *testing.T) {
	d := &library.BorrowDescriptor{}
	data := validBorrowData()
	delete(data, "book_id")
	err := d.Validate(data)
	assert.ErrorContains(t, err, "book_id wajib diisi")
}

func TestBorrowDescriptor_Validate_MissingBorrowerType(t *testing.T) {
	d := &library.BorrowDescriptor{}
	data := validBorrowData()
	delete(data, library.FieldBorrowerType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "borrower_type wajib diisi")
}

func TestBorrowDescriptor_Validate_InvalidBorrowerType(t *testing.T) {
	d := &library.BorrowDescriptor{}
	data := validBorrowData()
	data[library.FieldBorrowerType] = "parent"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "borrower_type tidak valid")
}

func TestBorrowDescriptor_Validate_MissingBorrowerID(t *testing.T) {
	d := &library.BorrowDescriptor{}
	data := validBorrowData()
	delete(data, library.FieldBorrowerID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "borrower_id wajib diisi")
}

func TestBorrowDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &library.BorrowDescriptor{}
	data := validBorrowData()
	delete(data, library.FieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestBorrowDescriptor_Validate_MissingBorrowDate(t *testing.T) {
	d := &library.BorrowDescriptor{}
	data := validBorrowData()
	delete(data, library.FieldBorrowDate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "borrow_date wajib diisi")
}
