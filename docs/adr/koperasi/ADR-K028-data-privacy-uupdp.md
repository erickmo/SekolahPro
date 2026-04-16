# ADR-K028: Data Privacy & Personal Data Protection (UU PDP)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

UU No. 27/2022 tentang Perlindungan Data Pribadi (UU PDP) mewajibkan setiap organisasi yang memproses data pribadi untuk mematuhi prinsip-prinsip perlindungan data. Untuk koperasi yang mengelola data keuangan nasabah, kepatuhan ini sangat penting karena:

- Data keuangan termasuk **data pribadi sensitif** (kategori khusus)
- Data nasabah meliputi KTP, NPWP, foto, biometrik, saldo, riwayat transaksi
- Koperasi beroperasi di lingkungan sekolah dengan data **anak di bawah umur** (students)
- Pelanggaran UU PDP: sanksi administratif hingga Rp 5 miliar + pidana penjara

Saat ini tidak ada ADR yang mengatur:
- Consent management untuk pengumpulan dan pemrosesan data
- Hak nasabah atas data mereka (akses, koreksi, hapus, portabilitas)
- Data breach notification procedure
- Data Processing Agreement (DPA) dengan pihak ketiga
- Privacy Impact Assessment untuk fitur baru

## Decision

### 1. Data Classification

```
data_classification:
├── PUBLIC
│   ├── Nama koperasi, alamat, kontak
│   ├── Produk yang ditawarkan
│   └── Informasi umum lainnya
│
├── INTERNAL
│   ├── Kebijakan internal (tidak berisi data pribadi)
│   ├── Laporan aggregat (tanpa identitas)
│   └── Prosedur operasional
│
├── CONFIDENTIAL
│   ├── Data pribadi umum: nama, alamat, telepon, email
│   ├── Data keanggotaan: nomor anggota, status
│   └── Data keuangan non-sensitif: produk yang dimiliki
│
└── HIGHLY_CONFIDENTIAL
    ├── Data pribadi sensitif: KTP, NPWP, foto, biometrik
    ├── Data keuangan sensitif: saldo, riwayat transaksi, pinjaman
    ├── Data kesehatan (jika ada)
    ├── Data anak di bawah umur (student e-wallet data)
    └── Password, PIN, security credentials
```

### 2. Legal Basis for Processing

```
processing_legal_basis:
├── CONSENT                ← Nasabah memberikan persetujuan eksplisit
├── CONTRACTUAL            ← Pelaksanaan perjanjian (pembukaan rekening, pinjaman)
├── LEGAL_OBLIGATION       ← Kewajiban hukum (laporan OJK, PPATK, pajak)
├── VITAL_INTEREST         ← Melindungi kepentingan vital nasabah
├── PUBLIC_INTEREST        ← Kepentingan umum (regulatory reporting)
└── LEGITIMATE_INTEREST    ← Kepentingan bisnis yang sah (credit scoring, fraud prevention)
```

**Mapping legal basis per data use:**

| Data Use | Legal Basis | Consent Required |
|---|---|---|
| Membuka rekening | Contractual | Ya (T&C) |
| Transaksi harian | Contractual | Tidak |
| Laporan ke OJK/PPATK | Legal obligation | Tidak |
| Credit scoring | Legitimate interest | Ya (disclosure) |
| Marketing produk baru | Consent | **Ya (opt-in)** |
| Berbagi data ke orang tua (student) | Contractual + Consent | Ya |
| Audit internal | Legitimate interest | Tidak |
| Backup & DR | Legitimate interest | Tidak |

### 3. Consent Management

```
consent_record:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Consent Info ──
├── consent_type          ENUM (data_collection, data_processing, data_sharing,
│                                marketing, biometric, minor_parental,
│                                cross_border_transfer, profiling)
├── consent_purpose       TEXT            ← Untuk apa data digunakan
├── legal_basis           ENUM (consent, contractual, legal_obligation,
│                                vital_interest, public_interest, legitimate_interest)
│
├── ── Consent Detail ──
├── consent_text          TEXT            ← Teks persetujuan yang ditampilkan
├── consent_version       VARCHAR         ← "v1.0", "v1.1" — versioning
├── consent_given         BOOLEAN
├── consent_method        ENUM (digital_signature, checkbox, verbal_recorded, written)
│
├── ── Withdrawal ──
├── withdrawn             BOOLEAN DEFAULT false
├── withdrawn_at          TIMESTAMPTZ (nullable)
├── withdrawal_reason     TEXT (nullable)
│
├── ── Parental Consent (untuk minor) ──
├── parent_id             UUID (nullable, FK → nasabah)
├── parent_relationship   ENUM (father, mother, guardian, nullable)
├── parent_consent_given  BOOLEAN (nullable)
│
├── ── Audit ──
├── given_at              TIMESTAMPTZ
├── ip_address            VARCHAR (nullable)
├── user_agent            VARCHAR (nullable)
├── witness_id            UUID (nullable, FK → user, staff yang menyaksikan)
└── created_at            TIMESTAMPTZ
```

**Aturan consent:**
- Consent harus **spesifik, informed, dan voluntary** — tidak boleh di-bundle
- Consent untuk data sensitif harus **eksplisit** (opt-in, bukan opt-out)
- Consent bisa dicabut kapan saja tanpa mengurangi layanan yang sudah ada
- Jika consent dicabut → data yang terkait harus di-anonymize atau dihapus (sesuai konteks)
- Untuk anak < 17 tahun → **WAJIB** parental consent

### 4. Data Subject Rights

```
data_subject_request:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Request Info ──
├── request_type          ENUM (access, rectification, erasure, portability,
│                                restriction, objection, withdraw_consent)
├── description           TEXT            ← Detail permintaan
├── data_categories       VARCHAR[]       ← Kategori data yang diminta
│
├── ── Processing ──
├── status                ENUM (received, verified, in_progress, completed,
│                                partially_completed, rejected)
├── assigned_to           UUID (FK → user, nullable)
├── rejection_reason      TEXT (nullable)
│
├── ── Response ──
├── response_data         JSONB (nullable)  ← Data yang diberikan (untuk access/portability)
├── response_summary      TEXT (nullable)
├── completed_at          TIMESTAMPTZ (nullable)
│
├── ── SLA ──
├── received_at           TIMESTAMPTZ
├── deadline_at           TIMESTAMPTZ     ← Max 72 jam untuk acknowledgment, 14 hari untuk completion
│
└── created_at            TIMESTAMPTZ
```

**SLA per request type:**

| Request Type | Acknowledgment | Completion | Keterangan |
|---|---|---|---|
| Access | 72 jam | 14 hari | Berikan salinan data pribadi yang dimiliki |
| Rectification | 72 jam | 7 hari | Perbaiki data yang tidak akurat |
| Erasure | 72 jam | 14 hari | Hapus data (kecuali wajib retensi regulasi) |
| Portability | 72 jam | 14 hari | Export data dalam format terstruktur |
| Restriction | 72 jam | 7 hari | Batasi pemrosesan data tertentu |
| Objection | 72 jam | 14 hari | Hentikan pemrosesan berdasarkan legitimate interest |
| Withdraw consent | 72 jam | 7 hari | Hentikan pemrosesan berdasarkan consent |

**Catatan penting untuk erasure:**
- Data yang wajib disimpan regulasi (transaksi, CDD, PPATK) **TIDAK BOLEH** dihapus
- Solusi: anonymize data (hapus identitas, pertahankan record untuk audit)
- Berikan penjelasan ke nasabah tentang data mana yang tidak bisa dihapus dan alasannya

### 5. Data Breach Notification

```
data_breach:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Breach Info ──
├── breach_type           ENUM (unauthorized_access, data_leak, ransomware,
│                                physical_theft, insider_threat, third_party_breach)
├── detected_at           TIMESTAMPTZ
├── detected_by           UUID (FK → user)
│
├── ── Impact Assessment ──
├── affected_data_types   VARCHAR[]       ← ["ktp", "saldo", "transaksi"]
├── affected_nasabah_count INT
├── severity              ENUM (low, medium, high, critical)
├── risk_to_individuals   ENUM (none, low, medium, high, very_high)
│
├── ── Containment ──
├── containment_actions   TEXT[]
├── root_cause            TEXT
├── remediation_plan      TEXT
│
├── ── Notifications ──
├── authority_notified    BOOLEAN DEFAULT false
├── authority_notified_at TIMESTAMPTZ (nullable)
├── subjects_notified     BOOLEAN DEFAULT false
├── subjects_notified_at  TIMESTAMPTZ (nullable)
├── notification_method   ENUM (email, whatsapp, letter, public_announcement)
│
├── ── Status ──
├── status                ENUM (detected, investigating, contained, resolved, closed)
├── resolved_at           TIMESTAMPTZ (nullable)
├── post_mortem_doc_id    UUID (nullable, FK → document)
│
├── ── Audit ──
└── created_at            TIMESTAMPTZ
```

**Breach notification timeline (UU PDP):**
- **72 jam** setelah mengetahui breach → notifikasi ke otoritas (Kementerian/Komisi PDP)
- **Secara langsung** → notifikasi ke subjek data jika risk tinggi
- Notifikasi harus berisi: jenis breach, data yang terdampak, langkah mitigasi, kontak

### 6. Privacy by Design Principles

**Untuk setiap fitur baru di sistem koperasi:**

1. **Data Minimization**: Hanya kumpulkan data yang benar-benar dibutuhkan
   - Jangan minta foto jika tidak diperlukan
   - Jangan simpan full KTP jika nomor saja cukup untuk use case tertentu

2. **Purpose Limitation**: Data hanya digunakan untuk tujuan yang dinyatakan
   - Data KYC tidak boleh dipakai untuk marketing tanpa consent terpisah

3. **Storage Limitation**: Data tidak disimpan lebih lama dari yang diperlukan
   - Otomatis anonymize data setelah periode retensi berakhir

4. **Access Control**: Akses ke data dibatasi berdasarkan kebutuhan (least privilege)
   - Teller hanya lihat data yang relevan untuk transaksi
   - Manager tidak perlu akses ke semua data nasabah individual

5. **Encryption**: Data sensitif di-encrypt baik at-rest maupun in-transit
   - KTP/NPWP: encrypted at rest, masked di display (cuma tampil 4 digit terakhir)
   - Password/PIN: hashed (bcrypt), tidak pernah plain text
   - Biometric: encrypted, tidak disimpan di server jika bisa (on-device)

### 7. Third-Party Data Processing

```
data_processing_agreement:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Agreement Info ──
├── processor_name        VARCHAR         ← Nama pihak ketiga
├── processor_type        ENUM (cloud_provider, payment_gateway, sms_provider,
│                                whatsapp_api, biometric_vendor, analytics)
├── agreement_doc_id      UUID (FK → document)
│
├── ── Scope ──
├── data_categories       VARCHAR[]       ← Jenis data yang diproses
├── processing_purpose    TEXT
├── processing_location   VARCHAR         ← Lokasi server (country)
├── cross_border          BOOLEAN         ← Apakah data keluar Indonesia
│
├── ── Term ──
├── effective_date        DATE
├── expiry_date           DATE (nullable)
├── is_active             BOOLEAN DEFAULT true
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

**Aturan untuk pihak ketiga:**
- **WAJIB** ada DPA tertulis sebelum sharing data
- Pihak ketiga tidak boleh menggunakan data untuk tujuan lain
- Pihak ketiga wajib melaporkan breach dalam 24 jam
- Penyimpanan data di luar Indonesia perlu persetujuan nasabah (cross-border)

### 8. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Data sensitif | Standar (KTP, NPWP, keuangan) | + data ibadah (zakat, infaq) |
| Consent untuk produk | Standar financial product | + consent untuk produk syariah |
| DPS akses data | Tidak ada | DPS bisa akses data untuk audit syariah |
| Zakat recipient data | Tidak ada | Data mustahik = HIGHLY_CONFIDENTIAL |
| Laporan ke regulator | OJK/Dinas Koperasi | + DPS audit report |

### 9. Vernon _rels dan _data Structure

**Consent _rels:**
```json
{
  "tenant_id":   "018f...",
  "nasabah_id":  "018f..."
}
```

**Consent _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  }
}
```

### 10. Authorization — RBAC

| Permission | Teller | Compliance | Manager | Admin |
|---|---|---|---|---|
| View consent status | v | v | v | v |
| Record consent | v | v | v | v |
| Process data subject request | - | v | v | v |
| View breach logs | - | v | v | v |
| Report breach to authority | - | - | v | v |
| Notify affected nasabah | - | - | v | v |
| Manage DPA | - | - | v | v |
| Configure data classification | - | - | - | v |
| Anonymize/delete data | - | v | v | v |
| Privacy impact assessment | - | v | v | v |
