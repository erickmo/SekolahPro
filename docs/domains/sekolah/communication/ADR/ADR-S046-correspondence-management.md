# ADR-S046: Correspondence Management (Manajemen Surat Menyurat TU)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Tata Usaha (TU) adalah unit administrasi sekolah yang bertanggung jawab atas seluruh pengelolaan surat menyurat. Ini adalah **core activity** staf TU sehari-hari — mulai dari mencatat surat masuk, membuat surat keluar, mengurus disposisi, hingga pengarsipan.

Permasalahan saat ini:

1. **Buku agenda manual**: Surat masuk/keluar dicatat di buku tulis — sulit dicari, rentan hilang.
2. **Nomor surat tidak konsisten**: Format penomoran berbeda antar staf TU, sering terjadi nomor ganda.
3. **Disposisi lambat**: Surat masuk menumpuk di meja kepsek karena tidak ada tracking siapa yang sedang proses.
4. **Template tidak standar**: Setiap kali membuat surat keterangan atau surat tugas, staf TU menulis ulang dari nol.
5. **Arsip fisik**: Surat disimpan di lemari — pencarian memerlukan waktu lama, risiko rusak/hilang.
6. **Tanda tangan basah**: Kepsek harus hadir fisik untuk menandatangani surat — menghambat proses jika kepsek dinas luar.

Ruang lingkup surat menyurat sekolah:

| Jenis | Contoh | Volume |
|-------|--------|--------|
| **Surat Dinas** | Surat ke Dinas Pendidikan, BOS, Dapodik | Medium |
| **Surat Keterangan** | Keterangan aktif siswa, keterangan pindah, keterangan lulus | Tinggi |
| **Surat Tugas** | Penugasan guru untuk diklat, workshop, penilaian | Medium |
| **Surat Undangan** | Undangan rapat komite, acara sekolah | Medium |
| **Surat Edaran** | Edaran libur, edaran kegiatan | Low-Medium |
| **SK Kepala Sekolah** | SK penugasan, SK pembagian tugas mengajar | Medium |

Domain ini berinteraksi dengan:
- **S047** (Approval Workflow): disposisi dan persetujuan surat menggunakan workflow engine.
- **ADR-012** (Teachers & Staff): pengirim/penerima surat internal.
- **ADR-013** (Users & Roles): role-based access (admin_tu, kepsek, wakasek).
- **S043** (Communication): notifikasi disposisi baru ke penerima.

### Mengapa Vernon Pattern (dengan catatan)?

- Read-heavy: arsip surat dibaca jauh lebih sering daripada ditulis.
- Relasi ke users, teachers, academic_year.
- Search & filter intensif: cari surat berdasarkan nomor, tanggal, pengirim, klasifikasi.
- Template rendering memerlukan denormalisasi data sekolah dan penerima.
- Disposition chain = nested read — cocok untuk _data cache.

> **C-Suite CTO Review Note (2026-04-15):**
> Tabel `correspondences` dan `correspondence_templates` tepat menggunakan Vernon (read-heavy, arsip).
> Namun `correspondence_dispositions` bersifat **write-heavy** (state transitions: pending → read → in_progress → completed → forwarded).
> Setiap perubahan status memicu Vernon sync yang overhead-nya mungkin tidak justified.
> **Rekomendasi:** Evaluasi saat implementasi — jika write frequency disposisi > 10x read, pertimbangkan
> migrasi `correspondence_dispositions` ke CQRS murni dan hapus `_rels`/`_data` dari tabel tersebut.
> Tabel utama `correspondences` tetap Vernon.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `correspondences` (surat masuk & keluar), `correspondence_templates` (template surat), dan `correspondence_dispositions` (disposisi/routing surat masuk).

### Table Schema

```sql
-- Surat masuk & keluar
CREATE TABLE correspondences (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas surat
    letter_number   VARCHAR(100),
    direction       VARCHAR(10) NOT NULL,
    classification  VARCHAR(30) NOT NULL,
    subject         VARCHAR(500) NOT NULL,
    letter_date     DATE NOT NULL,
    received_date   DATE,

    -- Auto-numbering
    sequence_number INT,
    fiscal_year     INT NOT NULL,

    -- Pengirim / penerima
    sender_name     VARCHAR(255) NOT NULL,
    sender_institution VARCHAR(255),
    recipient_name  VARCHAR(255) NOT NULL,
    recipient_institution VARCHAR(255),

    -- Konten
    body            TEXT,
    template_id     UUID,
    priority        VARCHAR(10) NOT NULL DEFAULT 'normal',

    -- Tanda tangan
    signer_id       UUID,
    signed_at       TIMESTAMPTZ,
    signature_type  VARCHAR(20),

    -- File attachment
    attachment_urls  JSONB NOT NULL DEFAULT '[]',

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',

    -- Metadata
    registered_by   UUID NOT NULL,
    academic_year_id UUID,
    notes           TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_correspondence_direction CHECK (direction IN ('incoming', 'outgoing')),
    CONSTRAINT chk_correspondence_classification CHECK (classification IN (
        'surat_dinas', 'surat_keterangan', 'surat_tugas',
        'surat_undangan', 'surat_edaran', 'sk_kepala_sekolah'
    )),
    CONSTRAINT chk_correspondence_priority CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    CONSTRAINT chk_correspondence_status CHECK (status IN (
        'draft', 'registered', 'in_disposition', 'completed', 'archived'
    )),
    CONSTRAINT chk_correspondence_signature_type CHECK (
        signature_type IS NULL OR signature_type IN ('wet', 'digital', 'stamped')
    )
);

-- Indexes
CREATE INDEX idx_correspondence_tenant_company ON correspondences (tenant_id, company_id);
CREATE INDEX idx_correspondence_direction ON correspondences (direction);
CREATE INDEX idx_correspondence_classification ON correspondences (classification);
CREATE INDEX idx_correspondence_letter_date ON correspondences (letter_date);
CREATE INDEX idx_correspondence_letter_number ON correspondences (letter_number);
CREATE INDEX idx_correspondence_status ON correspondences (status);
CREATE INDEX idx_correspondence_fiscal_year_seq ON correspondences (fiscal_year, sequence_number) WHERE direction = 'outgoing';
CREATE INDEX idx_correspondence_signer ON correspondences (signer_id) WHERE signer_id IS NOT NULL;
CREATE INDEX idx_correspondence_academic_year ON correspondences (academic_year_id) WHERE academic_year_id IS NOT NULL;
CREATE INDEX idx_correspondence_rels ON correspondences USING GIN (_rels);
CREATE INDEX idx_correspondence_data ON correspondences USING GIN (_data);
CREATE INDEX idx_correspondence_subject_trgm ON correspondences USING GIN (subject gin_trgm_ops);


-- Template surat
CREATE TABLE correspondence_templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Template metadata
    name            VARCHAR(255) NOT NULL,
    classification  VARCHAR(30) NOT NULL,
    description     TEXT,

    -- Template content (Go template / Handlebars syntax)
    header_template TEXT,
    body_template   TEXT NOT NULL,
    footer_template TEXT,

    -- Variabel yang dibutuhkan
    required_variables JSONB NOT NULL DEFAULT '[]',

    -- Numbering format
    number_format   VARCHAR(100),

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    version         INT NOT NULL DEFAULT 1,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_template_classification CHECK (classification IN (
        'surat_dinas', 'surat_keterangan', 'surat_tugas',
        'surat_undangan', 'surat_edaran', 'sk_kepala_sekolah'
    )),
    CONSTRAINT uq_template_name_tenant UNIQUE (tenant_id, company_id, name)
);

-- Indexes
CREATE INDEX idx_template_tenant_company ON correspondence_templates (tenant_id, company_id);
CREATE INDEX idx_template_classification ON correspondence_templates (classification);
CREATE INDEX idx_template_active ON correspondence_templates (is_active) WHERE is_active = true;
CREATE INDEX idx_template_rels ON correspondence_templates USING GIN (_rels);
CREATE INDEX idx_template_data ON correspondence_templates USING GIN (_data);


-- Disposisi surat masuk
CREATE TABLE correspondence_dispositions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Parent
    correspondence_id UUID NOT NULL,

    -- Routing
    from_user_id    UUID NOT NULL,
    to_user_id      UUID NOT NULL,
    disposition_type VARCHAR(30) NOT NULL,
    instruction     TEXT,
    due_date        DATE,

    -- Response
    response        TEXT,
    responded_at    TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- Chain level (kepsek=1 → wakasek=2 → guru=3)
    level           INT NOT NULL DEFAULT 1,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_disposition_type CHECK (disposition_type IN (
        'for_action', 'for_review', 'for_information', 'for_signature', 'for_filing'
    )),
    CONSTRAINT chk_disposition_status CHECK (status IN (
        'pending', 'read', 'in_progress', 'completed', 'forwarded'
    ))
);

-- Indexes
CREATE INDEX idx_disposition_tenant_company ON correspondence_dispositions (tenant_id, company_id);
CREATE INDEX idx_disposition_correspondence ON correspondence_dispositions (correspondence_id);
CREATE INDEX idx_disposition_to_user ON correspondence_dispositions (to_user_id);
CREATE INDEX idx_disposition_from_user ON correspondence_dispositions (from_user_id);
CREATE INDEX idx_disposition_status ON correspondence_dispositions (status);
CREATE INDEX idx_disposition_due_date ON correspondence_dispositions (due_date) WHERE status = 'pending';
CREATE INDEX idx_disposition_rels ON correspondence_dispositions USING GIN (_rels);
CREATE INDEX idx_disposition_data ON correspondence_dispositions USING GIN (_data);
```

### Field Design Rationale

**correspondences:**

| Field | Keputusan | Alasan |
|---|---|---|
| `letter_number` | VARCHAR(100), nullable | Nullable karena surat keluar baru mendapat nomor saat di-finalize; surat masuk menggunakan nomor asli |
| `direction` | VARCHAR(10), CHECK | Incoming vs outgoing — menentukan workflow yang berbeda |
| `classification` | VARCHAR(30), CHECK | 6 klasifikasi standar surat sekolah — cukup untuk MVP |
| `sequence_number` | INT, nullable | Auto-increment per fiscal_year per tenant — untuk generate nomor surat keluar |
| `fiscal_year` | INT, NOT NULL | Tahun anggaran untuk penomoran surat (bisa beda dari tahun ajaran) |
| `sender_name` / `recipient_name` | VARCHAR(255) | Denormalisasi — surat eksternal tidak punya user_id di sistem |
| `body` | TEXT, nullable | Nullable karena surat masuk hanya discan, tidak perlu body text |
| `template_id` | UUID, nullable | Referensi ke template yang digunakan (jika dari template) |
| `attachment_urls` | JSONB array | Multiple file attachment (scan surat, lampiran) |
| `signer_id` | UUID, nullable | User yang menandatangani — nullable untuk surat masuk |
| `signature_type` | VARCHAR(20) | wet (basah), digital, stamped (cap) |

**correspondence_templates:**

| Field | Keputusan | Alasan |
|---|---|---|
| `body_template` | TEXT, NOT NULL | Template body menggunakan Go template syntax dengan variabel |
| `required_variables` | JSONB array | List variabel yang harus diisi user saat generate dari template |
| `number_format` | VARCHAR(100) | Format penomoran: `{kode_sekolah}/{nomor_urut}/{bulan_romawi}/{tahun}` |
| `version` | INT | Template versioning — saat template di-update, versi lama tetap tersimpan |

**correspondence_dispositions:**

| Field | Keputusan | Alasan |
|---|---|---|
| `disposition_type` | VARCHAR(30), CHECK | 5 jenis disposisi standar kedinasan |
| `level` | INT | Chain level: kepsek=1, wakasek=2, guru/staf=3 — mendukung multi-level |
| `due_date` | DATE, nullable | Deadline penyelesaian disposisi |
| `response` | TEXT, nullable | Catatan balasan dari penerima disposisi |

### Vernon Relationships

**correspondences:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `registered_by_user` | belongs_to | **Ya** | Staf TU yang mencatat surat |
| `signer` | belongs_to | Tidak | Hanya dimuat saat detail view |
| `template` | belongs_to | Tidak | Hanya referensi, jarang ditampilkan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |
| `dispositions` | has_many | Tidak | Dimuat terpisah di detail view |

**correspondence_dispositions:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `correspondence` | belongs_to | **Ya** | Selalu perlu tahu surat induknya |
| `from_user` | belongs_to | **Ya** | Siapa yang mendisposisi |
| `to_user` | belongs_to | **Ya** | Siapa penerima disposisi |

### _rels / _data Structure

**correspondences:**

```json
{
  "_rels": {
    "registered_by": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "registered_by_user": {
      "id": "018f...",
      "full_name": "Ibu Siti (TU)",
      "role": "admin_tu"
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    },
    "disposition_summary": {
      "total": 3,
      "pending": 1,
      "completed": 2,
      "current_holder": "Pak Budi (Wakasek Kurikulum)"
    }
  }
}
```

**correspondence_dispositions:**

```json
{
  "_rels": {
    "correspondence_id": "018f...",
    "from_user_id": "018f...",
    "to_user_id": "018f..."
  },
  "_data": {
    "correspondence": {
      "id": "018f...",
      "letter_number": "SDN01/045/IV/2026",
      "subject": "Undangan Rapat Dinas Pendidikan",
      "direction": "incoming"
    },
    "from_user": {
      "id": "018f...",
      "full_name": "Pak Ahmad (Kepala Sekolah)",
      "role": "kepsek"
    },
    "to_user": {
      "id": "018f...",
      "full_name": "Pak Budi (Wakasek Kurikulum)",
      "role": "wakasek"
    }
  }
}
```

### Auto-Numbering Logic

Format penomoran surat keluar bersifat konfigurabel per tenant melalui `correspondence_templates.number_format`:

```
Default format: {kode_sekolah}/{nomor_urut}/{bulan_romawi}/{tahun}
Contoh output:  SDN01/045/IV/2026
```

| Variabel | Sumber | Contoh |
|----------|--------|--------|
| `{kode_sekolah}` | School profile (S048) | SDN01 |
| `{nomor_urut}` | Auto-increment per fiscal_year per classification | 045 |
| `{bulan_romawi}` | Bulan saat surat dibuat (I-XII) | IV |
| `{tahun}` | Tahun fiskal | 2026 |
| `{klasifikasi}` | Kode klasifikasi (opsional) | KET, TGS, UND |

Nomor urut di-generate atomically menggunakan:

```sql
SELECT COALESCE(MAX(sequence_number), 0) + 1
FROM correspondences
WHERE tenant_id = $1 AND company_id = $2
  AND fiscal_year = $3 AND classification = $4
  AND direction = 'outgoing'
FOR UPDATE;
```

### API Endpoints

```
# Surat (CRUD)
POST   /api/v1/correspondences                         — Registrasi surat baru (masuk/keluar)
GET    /api/v1/correspondences                         — List surat (filter: direction, classification, status, date range)
GET    /api/v1/correspondences/{id}                    — Detail surat + disposisi chain
PUT    /api/v1/correspondences/{id}                    — Update surat (sebelum finalized)
DELETE /api/v1/correspondences/{id}                    — Soft delete

# Penomoran
POST   /api/v1/correspondences/{id}/generate-number    — Generate nomor surat otomatis

# Tanda tangan
POST   /api/v1/correspondences/{id}/sign               — Tandatangani surat (digital/wet)

# Disposisi
POST   /api/v1/correspondences/{id}/dispositions        — Buat disposisi baru
GET    /api/v1/correspondences/{id}/dispositions        — List disposisi chain
PUT    /api/v1/correspondence-dispositions/{id}         — Update disposisi (response, status)
GET    /api/v1/correspondence-dispositions/inbox        — Inbox disposisi untuk user saat ini

# Template
POST   /api/v1/correspondence-templates                 — Buat template baru
GET    /api/v1/correspondence-templates                 — List template aktif
GET    /api/v1/correspondence-templates/{id}            — Detail template
PUT    /api/v1/correspondence-templates/{id}            — Update template
POST   /api/v1/correspondence-templates/{id}/generate   — Generate surat dari template

# Arsip & Pencarian
GET    /api/v1/correspondences/search                   — Full-text search arsip surat
GET    /api/v1/correspondences/statistics                — Statistik surat per bulan/klasifikasi
```

## Consequences

### Positive

- **Digitalisasi TU**: Menggantikan buku agenda manual — semua surat tercatat digital dengan pencarian cepat.
- **Nomor surat otomatis**: Format konfigurabel, sequence atomic — tidak ada lagi nomor ganda.
- **Disposisi terlacak**: Setiap surat masuk terlacak siapa yang sedang proses, sudah berapa lama.
- **Template reusable**: Surat keterangan, surat tugas bisa digenerate dalam hitungan detik.
- **Arsip digital**: Pencarian full-text, tidak ada risiko dokumen fisik hilang.
- **Tanda tangan digital**: Kepsek bisa menandatangani surat dari mana saja.
- **Integrasi S047**: Persetujuan surat keluar penting bisa melalui approval workflow.

### Negative / Trade-offs

- **Learning curve TU**: Staf TU yang terbiasa manual perlu adaptasi ke sistem digital.
- **Dual system**: Masa transisi masih perlu mencetak surat fisik — duplikasi kerja.
- **Template complexity**: Go template syntax mungkin terlalu teknis untuk admin TU — butuh UI builder.
- **File storage**: Scan surat memerlukan object storage terpisah (S3/MinIO) — belum di-scope di ADR ini.
- **Digital signature**: Integrasi tanda tangan digital (BSrE / BSSN) memerlukan infrastruktur tambahan.

## Alternatives Considered

### 1. Surat masuk dan keluar di tabel terpisah
- Ditolak: banyak field yang sama — satu tabel dengan `direction` lebih DRY dan query lebih sederhana.

### 2. Disposisi sebagai bagian dari S047 (Approval Workflow)
- Ditolak untuk disposisi routing: disposisi TU bukan approval (setuju/tolak), melainkan routing/assignment (teruskan ke siapa). S047 digunakan untuk persetujuan surat keluar penting, bukan disposisi rutin.

### 3. Menggunakan document management system terpisah (DMS)
- Ditolak: over-engineering untuk MVP. Surat sekolah volumenya kecil (ratusan per tahun), cukup dengan tabel + file attachment.

### 4. Template menggunakan DOCX/ODT generation
- Deferred: bisa ditambahkan nanti. MVP menggunakan HTML template yang di-render ke PDF.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"correspondences"` |
| U02 | `DefaultRels()` returns autoloaded rels | — | `registered_by_user`, `academic_year` |
| U03 | Validate rejects invalid direction | `{ "direction": "internal" }` | Error: direction must be incoming/outgoing |
| U04 | Validate rejects invalid classification | `{ "classification": "memo" }` | Error: classification not in allowed list |
| U05 | Validate rejects missing subject | `{ "subject": "" }` | Error: subject required |
| U06 | Validate rejects missing letter_date | `{ "letter_date": null }` | Error: letter_date required |
| U07 | Generate letter number with format | `{ "format": "{kode}/{seq}/{bulan}/{tahun}" }` | `"SDN01/001/I/2026"` |
| U08 | Roman numeral month conversion | month = 4 | `"IV"` |
| U09 | Validate disposition type | `{ "type": "for_approval" }` | Error: invalid disposition_type |
| U10 | Template variable extraction | `"Nama: {{.nama_siswa}}"` | `["nama_siswa"]` |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Register incoming letter | POST with direction=incoming | 201, letter_number from original letter preserved |
| I02 | Register outgoing letter + auto-number | POST with direction=outgoing, then generate-number | 201, sequence_number assigned atomically |
| I03 | Concurrent numbering is atomic | 10 concurrent generate-number requests | All get unique sequential numbers |
| I04 | Create disposition chain | POST disposition kepsek→wakasek→guru | 201, 3 records with level 1,2,3 |
| I05 | Disposition inbox shows pending | GET inbox for wakasek | 200, only pending dispositions for that user |
| I06 | Update disposition response | PUT with response text | 200, responded_at auto-set |
| I07 | Generate letter from template | POST generate with template_id + variables | 201, body populated from template |
| I08 | Search archive by subject | GET search?q=rapat dinas | 200, trigram match results |
| I09 | Filter by classification + date | GET ?classification=surat_tugas&from=2026-01-01 | 200, filtered results |
| I10 | Sign correspondence | POST sign with signature_type=digital | 200, signed_at + signer_id set |
| I11 | Cannot edit signed letter | PUT on signed correspondence | 403, immutable after signing |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I12 | UserUpdated syncs to correspondence | Update registered_by user name | `_data.registered_by_user.full_name` updated |
| I13 | UserUpdated syncs to dispositions | Update to_user name | `_data.to_user.full_name` updated |
| I14 | CorrespondenceUpdated syncs to dispositions | Update correspondence subject | `_data.correspondence.subject` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | Cannot access other tenant's letters | GET with wrong tenant scope | 404 |
| I16 | Cannot create disposition on other tenant's letter | POST disposition with cross-tenant correspondence_id | 403/404, scope violation |
| I17 | Letter numbering is tenant-scoped | Two tenants generate numbers simultaneously | Each has own sequence starting from 1 |
