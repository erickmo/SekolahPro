# ADR-S038: Library Management / Perpustakaan

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Perpustakaan adalah fasilitas wajib di setiap sekolah Indonesia sesuai Peraturan Pemerintah No. 24/2014. Manajemen perpustakaan digital diperlukan untuk:

1. **Katalog buku**: Master data buku dengan ISBN, pengarang, kategori, lokasi rak.
2. **Copy management**: Satu judul bisa punya beberapa salinan (exemplar) — tracking per salinan.
3. **Peminjaman/pengembalian**: Transaksi borrow/return per salinan buku.
4. **Denda keterlambatan**: Kalkulasi dan tracking denda otomatis.
5. **Peminjam**: Siswa dan guru bisa meminjam (ADR-S001, ADR-012).
6. **Laporan**: Statistik peminjaman, buku populer, siswa aktif membaca.
7. **Pesantren**: Koleksi kitab kuning, Al-Quran, dan buku Islam (ADR-009).

Konteks sekolah Indonesia:
- Rata-rata perpustakaan sekolah memiliki 1.000-10.000 judul buku.
- Batas peminjaman biasanya 2-3 buku per siswa, durasi 7-14 hari.
- Denda keterlambatan Rp 500 - Rp 1.000 per hari per buku.
- Pesantren memiliki koleksi **kitab kuning** dan **Al-Quran** yang perlu dikategorikan terpisah.
- Beberapa sekolah sudah menggunakan barcode scanner untuk tracking.

### Mengapa Vernon Pattern?

- Master data (buku, copy) = read-heavy, jarang berubah.
- Transaksi peminjaman = has_many dari student, moderate volume.
- Relasi ke student, teacher, academic_year.
- Business logic moderate: due date calculation, fine computation, availability check.
- Eventually consistent acceptable.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `library_books` (master judul), `library_copies` (salinan per judul), dan `library_borrows` (transaksi peminjaman).

### Table Schema

```sql
-- Master judul buku
CREATE TABLE library_books (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    title           VARCHAR(500) NOT NULL,
    isbn            VARCHAR(20),
    author          VARCHAR(300) NOT NULL,
    publisher       VARCHAR(200),
    publish_year    INT,
    edition         VARCHAR(50),

    -- Klasifikasi
    category        VARCHAR(30) NOT NULL,
    subcategory     VARCHAR(50),
    language        VARCHAR(20) NOT NULL DEFAULT 'indonesian',
    ddc_code        VARCHAR(20),

    -- Lokasi
    shelf_location  VARCHAR(50),

    -- Statistik (denormalisasi)
    total_copies    INT NOT NULL DEFAULT 0,
    available_copies INT NOT NULL DEFAULT 0,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_library_book_isbn UNIQUE (tenant_id, company_id, isbn),
    CONSTRAINT chk_library_category CHECK (category IN (
        'textbook', 'reference', 'fiction', 'non_fiction',
        'science', 'social', 'religion', 'kitab_kuning',
        'al_quran', 'magazine', 'journal', 'other'
    )),
    CONSTRAINT chk_library_language CHECK (language IN (
        'indonesian', 'english', 'arabic', 'javanese', 'other'
    ))
);

-- Indexes
CREATE INDEX idx_library_book_tenant_company ON library_books (tenant_id, company_id);
CREATE INDEX idx_library_book_isbn ON library_books (isbn) WHERE isbn IS NOT NULL;
CREATE INDEX idx_library_book_category ON library_books (category);
CREATE INDEX idx_library_book_author ON library_books (author);
CREATE INDEX idx_library_book_title ON library_books USING gin (to_tsvector('indonesian', title));
CREATE INDEX idx_library_book_rels ON library_books USING GIN (_rels);
CREATE INDEX idx_library_book_data ON library_books USING GIN (_data);

-- Salinan buku (exemplar)
CREATE TABLE library_copies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    book_id         UUID NOT NULL,

    -- Identitas salinan
    copy_number     INT NOT NULL,
    barcode         VARCHAR(50),

    -- Kondisi
    condition       VARCHAR(20) NOT NULL DEFAULT 'good',
    acquisition_date DATE,
    acquisition_source VARCHAR(50),

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'available',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_library_copy_number UNIQUE (book_id, copy_number),
    CONSTRAINT uq_library_copy_barcode UNIQUE (tenant_id, company_id, barcode),
    CONSTRAINT chk_library_copy_condition CHECK (condition IN ('new', 'good', 'fair', 'damaged', 'lost')),
    CONSTRAINT chk_library_copy_status CHECK (status IN ('available', 'borrowed', 'reserved', 'maintenance', 'lost', 'retired')),
    CONSTRAINT chk_library_copy_number CHECK (copy_number >= 1)
);

-- Indexes
CREATE INDEX idx_library_copy_tenant_company ON library_copies (tenant_id, company_id);
CREATE INDEX idx_library_copy_book ON library_copies (book_id);
CREATE INDEX idx_library_copy_barcode ON library_copies (barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_library_copy_status ON library_copies (status);
CREATE INDEX idx_library_copy_available ON library_copies (book_id, status) WHERE status = 'available';
CREATE INDEX idx_library_copy_rels ON library_copies USING GIN (_rels);
CREATE INDEX idx_library_copy_data ON library_copies USING GIN (_data);

-- Transaksi peminjaman
CREATE TABLE library_borrows (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    copy_id         UUID NOT NULL,
    book_id         UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Peminjam (bisa siswa atau guru)
    borrower_type   VARCHAR(10) NOT NULL,
    borrower_id     UUID NOT NULL,

    -- Tanggal
    borrow_date     DATE NOT NULL,
    due_date        DATE NOT NULL,
    return_date     DATE,

    -- Denda
    is_overdue      BOOLEAN NOT NULL DEFAULT false,
    overdue_days    INT NOT NULL DEFAULT 0,
    fine_amount     BIGINT NOT NULL DEFAULT 0,
    fine_paid       BOOLEAN NOT NULL DEFAULT false,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'borrowed',

    -- Pencatat
    processed_by    UUID NOT NULL,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_library_borrower_type CHECK (borrower_type IN ('student', 'teacher')),
    CONSTRAINT chk_library_borrow_status CHECK (status IN ('borrowed', 'returned', 'overdue', 'lost')),
    CONSTRAINT chk_library_fine CHECK (fine_amount >= 0),
    CONSTRAINT chk_library_overdue_days CHECK (overdue_days >= 0)
);

-- Indexes
CREATE INDEX idx_library_borrow_tenant_company ON library_borrows (tenant_id, company_id);
CREATE INDEX idx_library_borrow_copy ON library_borrows (copy_id);
CREATE INDEX idx_library_borrow_book ON library_borrows (book_id);
CREATE INDEX idx_library_borrow_borrower ON library_borrows (borrower_type, borrower_id);
CREATE INDEX idx_library_borrow_status ON library_borrows (status) WHERE status IN ('borrowed', 'overdue');
CREATE INDEX idx_library_borrow_due ON library_borrows (due_date) WHERE status = 'borrowed';
CREATE INDEX idx_library_borrow_date ON library_borrows (borrow_date);
CREATE INDEX idx_library_borrow_rels ON library_borrows USING GIN (_rels);
CREATE INDEX idx_library_borrow_data ON library_borrows USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `category` | 12 kategori | Termasuk `kitab_kuning` dan `al_quran` untuk pesantren (ADR-009) |
| `language` | 5 bahasa | Indonesia, English, Arabic (pesantren), Javanese (kitab salaf), other |
| `ddc_code` | VARCHAR(20), nullable | Dewey Decimal Classification — standar perpustakaan, nullable karena tidak semua sekolah menggunakan |
| `total_copies/available_copies` | Denormalisasi | Counter cepat — diupdate setiap copy ditambah/dipinjam/dikembalikan |
| `barcode` | VARCHAR(50), nullable | Untuk sekolah yang menggunakan barcode scanner — nullable karena tidak wajib |
| `condition` | 5 status | Lifecycle kondisi: new → good → fair → damaged → lost |
| `borrower_type` | student/teacher | Polymorphic borrower — bisa siswa atau guru |
| `due_date` | DATE, computed | Dihitung dari borrow_date + loan_period (configurable per sekolah, default 14 hari) |
| `fine_amount` | BIGINT | Dihitung dari overdue_days × fine_per_day (configurable, default Rp 500/hari) |

### Fine Calculation (Application Logic)

```
overdue_days = MAX(0, return_date - due_date)  // atau today - due_date jika belum dikembalikan
fine_amount  = overdue_days × fine_per_day

Default config:
- loan_period  = 14 hari (siswa), 30 hari (guru)
- max_borrows  = 3 (siswa), 5 (guru)
- fine_per_day  = 500 (Rupiah)
- max_fine      = 50000 (Rupiah) per buku
```

Config ini disimpan per tenant, bukan hardcoded.

### Vernon Relationships

**library_copies:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `book` | belongs_to | **Ya** | Judul buku selalu ditampilkan bersama salinan |

**library_borrows:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `copy` | belongs_to | **Ya** | Nomor salinan + barcode |
| `book` | belongs_to | **Ya** | Judul buku |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

### _rels / _data Structure

```json
// library_copies
{
  "_rels": {
    "book_id": "018f..."
  },
  "_data": {
    "book": { "id": "018f...", "title": "Fisika Dasar Jilid 1", "author": "Halliday", "isbn": "978-xxx", "category": "textbook" }
  }
}

// library_borrows
{
  "_rels": {
    "copy_id": "018f...",
    "book_id": "018f...",
    "academic_year_id": "018f...",
    "borrower_id": "018f..."
  },
  "_data": {
    "copy": { "id": "018f...", "copy_number": 1, "barcode": "LIB-00123" },
    "book": { "id": "018f...", "title": "Fisika Dasar Jilid 1", "author": "Halliday" },
    "borrower": { "id": "018f...", "full_name": "Ahmad Fauzi", "nis": "12345", "type": "student" },
    "academic_year": { "id": "018f...", "name": "2025/2026" }
  }
}
```

### API Endpoints

```
# Books (Master)
GET    /api/v1/library-books                           — List/search buku
POST   /api/v1/library-books                           — Buat judul buku
PUT    /api/v1/library-books/{id}                      — Update buku
GET    /api/v1/library-books/{id}                      — Detail buku + copies

# Copies
GET    /api/v1/library-books/{id}/copies               — List salinan per judul
POST   /api/v1/library-copies                          — Tambah salinan
PUT    /api/v1/library-copies/{id}                     — Update salinan (kondisi, status)

# Borrows
GET    /api/v1/library-borrows                         — List peminjaman aktif
POST   /api/v1/library-borrows                         — Pinjam buku
POST   /api/v1/library-borrows/{id}/return             — Kembalikan buku
GET    /api/v1/students/{id}/library-borrows           — Riwayat peminjaman siswa
GET    /api/v1/teachers/{id}/library-borrows           — Riwayat peminjaman guru
GET    /api/v1/library-borrows/overdue                 — Daftar peminjaman terlambat
POST   /api/v1/library-borrows/{id}/pay-fine           — Bayar denda

# Reports
GET    /api/v1/library/statistics                      — Statistik perpustakaan
GET    /api/v1/library/popular-books                   — Buku terpopuler
```

## Consequences

### Positive

- **Copy-level tracking**: Setiap salinan buku terlacak individu — penting untuk audit inventaris.
- **Pesantren support**: Kategori `kitab_kuning` dan `al_quran` + bahasa Arabic (ADR-009).
- **Polymorphic borrower**: Siswa dan guru bisa meminjam dengan satu tabel.
- **Auto fine calculation**: Denda dihitung otomatis berdasarkan konfigurasi per sekolah.
- **Barcode ready**: Support barcode scanning untuk sekolah yang sudah menggunakan.
- **Full-text search**: Index tsvector pada title untuk pencarian buku.

### Negative / Trade-offs

- **Availability denormalisasi**: `total_copies` dan `available_copies` perlu dijaga konsisten — risiko drift.
- **No reservation queue**: Belum ada antrian reservasi jika buku habis — hanya status `reserved` di copy.
- **No e-book**: Belum mendukung koleksi digital/e-book. Enhancement di masa depan.
- **Fine tidak otomatis collected**: Denda hanya dihitung dan dicatat — collection tetap manual.
- **No integration with S009**: Denda perpustakaan belum terintegrasi dengan student finance.

## Alternatives Considered

### 1. Tanpa copy management (langsung book → borrow)
- Ditolak: tidak bisa track kondisi per salinan, tidak bisa bedakan salinan yang hilang vs yang dipinjam.

### 2. Borrower sebagai 2 tabel terpisah (student_borrows, teacher_borrows)
- Ditolak: duplikasi struktur dan logic — polymorphic `borrower_type` + `borrower_id` lebih DRY.

### 3. Fine sebagai tabel terpisah
- Ditolak untuk MVP: fine tightly coupled dengan borrow transaction — kolom di borrow record cukup.

### 4. JSONB untuk book metadata
- Ditolak: author, publisher, category perlu queryable untuk search dan filter — kolom terpisah lebih efisien.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `BookDescriptor.TableName()` | — | `"library_books"` |
| U02 | `CopyDescriptor.TableName()` | — | `"library_copies"` |
| U03 | `BorrowDescriptor.TableName()` | — | `"library_borrows"` |
| U04 | Validate rejects invalid `category` | `"comic"` | Error: invalid category |
| U05 | Validate rejects invalid `language` | `"french"` | Error: invalid language |
| U06 | Validate rejects invalid `condition` | `"broken"` | Error: must be new/good/fair/damaged/lost |
| U07 | Validate rejects invalid `borrower_type` | `"parent"` | Error: must be student/teacher |
| U08 | Fine calculation: 5 days overdue | due=Apr 10, return=Apr 15, rate=500 | fine=2500 |
| U09 | Fine calculation: max cap | 200 days overdue, rate=500, max=50000 | fine=50000 |
| U10 | Validate accepts valid borrow | All fields valid | No error |

### Integration Tests — Books & Copies

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create book | POST with title + author + category | 201 |
| I02 | Unique ISBN per company | Create 2 books same ISBN | 409/422 |
| I03 | Add copy | POST copy for book | 201, `total_copies` incremented |
| I04 | Search by title | GET /library-books?q=fisika | 200, full-text search results |
| I05 | Category CHECK | INSERT with `category = 'comic'` | DB error |
| I06 | Language CHECK | INSERT with `language = 'french'` | DB error |

### Integration Tests — Borrows

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Borrow book | POST borrow | 201, copy status → borrowed, available_copies decremented |
| I08 | Borrow unavailable copy | Borrow already borrowed copy | 422, not available |
| I09 | Exceed max borrows | Student borrows 4th book (max=3) | 422, limit exceeded |
| I10 | Return book on time | POST /return within due_date | 200, fine=0, copy status → available |
| I11 | Return book late | POST /return after due_date | 200, fine calculated, is_overdue=true |
| I12 | Get overdue list | GET /library-borrows/overdue | 200, only overdue records |
| I13 | Pay fine | POST /pay-fine | 200, fine_paid=true |
| I14 | Borrow status CHECK | INSERT with `status = 'expired'` | DB error |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | BookUpdated syncs to copies | Update book title | `_data.book.title` updated |
| I16 | CopyUpdated syncs to borrows | Update copy barcode | `_data.copy.barcode` updated |
| I17 | BookUpdated syncs to borrows | Update book title | `_data.book.title` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I18 | Cannot access other tenant's books | GET with wrong tenant | 404 |
| I19 | Cannot borrow cross-tenant | POST borrow for other tenant's copy | Error |
