-- 056: library — Library Books, Copies, Borrows
-- Vernon pattern: _rels/_data JSONB columns.

--------------------------------------------------------------------------------
-- library_books
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS library_books (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Domain data
    title           VARCHAR(255) NOT NULL,
    author          VARCHAR(255),
    isbn            VARCHAR(20),
    publisher       VARCHAR(255),
    publication_year INT,
    category        VARCHAR(30) NOT NULL,
    language        VARCHAR(20) NOT NULL DEFAULT 'id',
    ddc_classification VARCHAR(20),
    description     TEXT,
    cover_url       TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_lb_category CHECK (category IN (
        'fiction', 'nonfiction', 'textbook', 'reference',
        'kitab_kuning', 'al_quran', 'hadits', 'fiqh', 'aqidah',
        'science', 'literature', 'other'
    )),
    CONSTRAINT chk_lb_language CHECK (language IN (
        'id', 'en', 'ar', 'jv', 'other'
    )),
    CONSTRAINT uq_lb_isbn UNIQUE (tenant_id, company_id, isbn) WHERE isbn IS NOT NULL
);

CREATE INDEX idx_library_books_tenant_company ON library_books (tenant_id, company_id);
CREATE INDEX idx_library_books_category ON library_books (category);
CREATE INDEX idx_library_books_language ON library_books (language);
CREATE INDEX idx_library_books_title ON library_books (title);
CREATE INDEX idx_library_books_author ON library_books (author);
CREATE INDEX idx_library_books_publication_year ON library_books (publication_year);
CREATE INDEX idx_library_books_rels ON library_books USING GIN (_rels);
CREATE INDEX idx_library_books_data ON library_books USING GIN (_data);

--------------------------------------------------------------------------------
-- library_copies
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS library_copies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    book_id         UUID NOT NULL,

    -- Domain data
    copy_number     VARCHAR(20) NOT NULL,
    barcode         VARCHAR(50),
    condition       VARCHAR(15) NOT NULL DEFAULT 'good',
    location_shelf  VARCHAR(50),
    acquisition_date DATE,
    acquisition_source VARCHAR(100),

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_lc_condition CHECK (condition IN (
        'new', 'good', 'fair', 'damaged', 'lost'
    )),
    CONSTRAINT uq_lc_copy_number UNIQUE (tenant_id, company_id, copy_number)
);

CREATE INDEX idx_library_copies_tenant_company ON library_copies (tenant_id, company_id);
CREATE INDEX idx_library_copies_book ON library_copies (book_id);
CREATE INDEX idx_library_copies_barcode ON library_copies (barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_library_copies_condition ON library_copies (condition);
CREATE INDEX idx_library_copies_rels ON library_copies USING GIN (_rels);
CREATE INDEX idx_library_copies_data ON library_copies USING GIN (_data);

--------------------------------------------------------------------------------
-- library_borrows
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS library_borrows (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    copy_id         UUID NOT NULL,
    book_id         UUID NOT NULL,
    borrower_type   VARCHAR(10) NOT NULL,
    borrower_id     UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Domain data
    borrow_date     DATE NOT NULL,
    due_date        DATE NOT NULL,
    return_date     DATE,
    actual_return_date DATE,
    fine_amount     NUMERIC(12,2) NOT NULL DEFAULT 0,
    fine_status     VARCHAR(15) DEFAULT 'none',
    note            TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_lbr_borrower_type CHECK (borrower_type IN (
        'student', 'teacher'
    )),
    CONSTRAINT chk_lbr_fine_status CHECK (fine_status IN (
        'none', 'pending', 'paid', 'waived'
    )),
    CONSTRAINT chk_lbr_fine_amount CHECK (fine_amount >= 0),
    CONSTRAINT chk_lbr_dates CHECK (due_date >= borrow_date)
);

CREATE INDEX idx_library_borrows_tenant_company ON library_borrows (tenant_id, company_id);
CREATE INDEX idx_library_borrows_copy ON library_borrows (copy_id);
CREATE INDEX idx_library_borrows_book ON library_borrows (book_id);
CREATE INDEX idx_library_borrows_academic_year ON library_borrows (academic_year_id);
CREATE INDEX idx_lb_borrower ON library_borrows (borrower_type, borrower_id);
CREATE INDEX idx_lb_dates ON library_borrows (borrow_date, due_date);
CREATE INDEX idx_library_borrows_fine_status ON library_borrows (fine_status) WHERE fine_status != 'none';
CREATE INDEX idx_library_borrows_rels ON library_borrows USING GIN (_rels);
CREATE INDEX idx_library_borrows_data ON library_borrows USING GIN (_data);
