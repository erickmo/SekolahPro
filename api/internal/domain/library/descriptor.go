// Package library adalah domain Vernon untuk manajemen perpustakaan.
//
// Terdiri dari 3 tabel: library_books, library_copies, library_borrows.
// library_books tidak memiliki BelongsTo (root master).
// library_copies memiliki 1 BelongsTo autoload: book.
// library_borrows memiliki 3 BelongsTo autoload: copy, book, academic_year.
// borrower adalah polymorphic (student/teacher) dan TIDAK di-autoload.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package library

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Book field constants ──────────────────────────────────────────────────────

const (
	FieldTitle             = "title"
	FieldAuthor            = "author"
	FieldIsbn              = "isbn"
	FieldPublisher         = "publisher"
	FieldPublicationYear   = "publication_year"
	FieldCategory          = "category"
	FieldLanguage          = "language"
	FieldDdcClassification = "ddc_classification"
	FieldDescription       = "description"
	FieldCoverUrl          = "cover_url"
)

// ── Copy field constants ──────────────────────────────────────────────────────

const (
	FieldBookID           = "book_id"
	FieldCopyNumber       = "copy_number"
	FieldBarcode          = "barcode"
	FieldCondition        = "condition"
	FieldLocationShelf    = "location_shelf"
	FieldAcquisitionDate  = "acquisition_date"
	FieldAcquisitionSource = "acquisition_source"
)

// ── Borrow field constants ────────────────────────────────────────────────────

const (
	FieldCopyID          = "copy_id"
	FieldBorrowerType    = "borrower_type"
	FieldBorrowerID      = "borrower_id"
	FieldAcademicYearID  = "academic_year_id"
	FieldBorrowDate      = "borrow_date"
	FieldDueDate         = "due_date"
	FieldReturnDate      = "return_date"
	FieldActualReturn    = "actual_return_date"
	FieldFineAmount      = "fine_amount"
	FieldFineStatus      = "fine_status"
	FieldNote            = "note"
)

// ── Book enum constants ───────────────────────────────────────────────────────

const (
	CatFiction     = "fiction"
	CatNonfiction  = "nonfiction"
	CatTextbook    = "textbook"
	CatReference   = "reference"
	CatKitabKuning = "kitab_kuning"
	CatAlQuran     = "al_quran"
	CatHadits      = "hadits"
	CatFiqh        = "fiqh"
	CatAqidah      = "aqidah"
	CatScience     = "science"
	CatLiterature  = "literature"
	CatOther       = "other"
)

const (
	LangID     = "id"
	LangEN     = "en"
	LangAR     = "ar"
	LangJV     = "jv"
	LangOther  = "other"
)

// ── Copy enum constants ───────────────────────────────────────────────────────

const (
	CondNew     = "new"
	CondGood    = "good"
	CondFair    = "fair"
	CondDamaged = "damaged"
	CondLost    = "lost"
)

// ── Borrow enum constants ─────────────────────────────────────────────────────

const (
	BorrowerStudent = "student"
	BorrowerTeacher = "teacher"
)

const (
	FineNone    = "none"
	FinePending = "pending"
	FinePaid    = "paid"
	FineWaived  = "waived"
)

// ── Relation name constants ───────────────────────────────────────────────────

const (
	RelBook         = "book"
	RelCopy         = "copy"
	RelAcademicYear = "academic_year"
)

var validCategories = map[string]bool{
	CatFiction: true, CatNonfiction: true, CatTextbook: true,
	CatReference: true, CatKitabKuning: true, CatAlQuran: true,
	CatHadits: true, CatFiqh: true, CatAqidah: true,
	CatScience: true, CatLiterature: true, CatOther: true,
}

var validLanguages = map[string]bool{
	LangID: true, LangEN: true, LangAR: true, LangJV: true, LangOther: true,
}

var validConditions = map[string]bool{
	CondNew: true, CondGood: true, CondFair: true, CondDamaged: true, CondLost: true,
}

var validBorrowerTypes = map[string]bool{
	BorrowerStudent: true, BorrowerTeacher: true,
}

var validFineStatuses = map[string]bool{
	FineNone: true, FinePending: true, FinePaid: true, FineWaived: true,
}

// ── BookDescriptor ────────────────────────────────────────────────────────────

// BookDescriptor mengimplementasi vernon.DomainDescriptor untuk library_books.
type BookDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL untuk buku.
func (d *BookDescriptor) TableName() string { return "library_books" }

// DefaultRels — library_books tidak memiliki relasi (root master).
func (d *BookDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant library_books.
func (d *BookDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldTitle, "title"); err != nil {
		return err
	}
	return requireEnum(data, FieldCategory, "category", validCategories)
}

// ── CopyDescriptor ────────────────────────────────────────────────────────────

// CopyDescriptor mengimplementasi vernon.DomainDescriptor untuk library_copies.
type CopyDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL untuk salinan buku.
func (d *CopyDescriptor) TableName() string { return "library_copies" }

// DefaultRels mendefinisikan relasi library_copies: book.
func (d *CopyDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelBook: {
			Domain:     "library_books",
			Type:       vernon.RelBelongsTo,
			FK:         FieldBookID,
			LocalKey:   FieldBookID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"title", "isbn"},
		},
	}
}

// Validate memvalidasi invariant library_copies.
func (d *CopyDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldBookID, "book_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldCopyNumber, "copy_number"); err != nil {
		return err
	}
	return requireEnum(data, FieldCondition, "condition", validConditions)
}

// ── BorrowDescriptor ──────────────────────────────────────────────────────────

// BorrowDescriptor mengimplementasi vernon.DomainDescriptor untuk library_borrows.
type BorrowDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL untuk peminjaman buku.
func (d *BorrowDescriptor) TableName() string { return "library_borrows" }

// DefaultRels mendefinisikan relasi library_borrows: copy + book + academic_year.
// borrower adalah polymorphic dan TIDAK di-autoload.
func (d *BorrowDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelCopy: {
			Domain:     "library_copies",
			Type:       vernon.RelBelongsTo,
			FK:         FieldCopyID,
			LocalKey:   FieldCopyID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"barcode", "condition"},
		},
		RelBook: {
			Domain:     "library_books",
			Type:       vernon.RelBelongsTo,
			FK:         "book_id",
			LocalKey:   "book_id",
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"title", "isbn"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearID,
			LocalKey:   FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant library_borrows.
func (d *BorrowDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldCopyID, "copy_id"); err != nil {
		return err
	}
	if err := requireString(data, "book_id", "book_id"); err != nil {
		return err
	}
	if err := requireEnum(data, FieldBorrowerType, "borrower_type", validBorrowerTypes); err != nil {
		return err
	}
	if err := requireString(data, FieldBorrowerID, "borrower_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}
	return requireString(data, FieldBorrowDate, "borrow_date")
}

// ── shared validation helpers ─────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}

func requireEnum(data map[string]any, field, label string, valid map[string]bool) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	if !valid[val] {
		return fmt.Errorf("%s tidak valid: %q", label, val)
	}
	return nil
}
