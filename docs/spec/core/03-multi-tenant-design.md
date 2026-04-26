# 03 — Desain Multi-Tenant SekolahPro

Dokumen ini menjabarkan arsitektur multi-tenant SekolahPro: 4-level hierarki organisasi, isolasi data antar tenant, sistem autentikasi two-phase JWT, dual-mode institution type (Umum vs Islam), serta entitas-entitas fondasi domain sekolah (Tahun Ajaran, Kelas, Guru, User & Roles).

---

## ADR References

| ADR | Judul | Relevansi |
|-----|-------|-----------|
| ADR-004 | Multi-Tenant 4-Level Hierarchy + Two-Phase JWT | Fondasi arsitektur multi-tenant |
| ADR-009 | Dual-Mode Institution Type (General / Islamic) | Dua mode institusi yang independent |
| ADR-010 | Academic Years | Tahun ajaran sebagai unit waktu fundamental |
| ADR-011 | Class Rooms | Kelas sebagai unit organisasi sekolah |
| ADR-012 | Teachers & Staff | Guru dan tenaga kependidikan |
| ADR-013 | Users & Roles | Authentication, authorization, RBAC |

---

## 1. Hierarki 4-Level

SekolahPro menggunakan 4 level scope organisasi sebagai standar untuk semua data:

```
Tenant  (Pelanggan SaaS / Holding Company)
  └── Company  (Anak perusahaan / Sekolah spesifik)
        └── Branch  (Cabang fisik — jarang digunakan di konteks sekolah)
              └── Warehouse  (Gudang — digunakan untuk Koperasi/inventori)
```

Dalam konteks SekolahPro (sekolah dan koperasi):

| Level | Entitas Bisnis | Contoh |
|-------|---------------|--------|
| **Tenant** | Yayasan pendidikan / pemilik SaaS | Yayasan Al-Hikmah, PT Sekolah Pintar |
| **Company** | Satu unit sekolah | SMP Al-Hikmah, SMK Al-Hikmah |
| **Branch** | Kampus cabang (opsional) | Kampus Utama, Kampus Selatan |
| **Warehouse** | Lokasi fisik inventori/kasir koperasi | Gudang Koperasi Utama |

### Deployment Mode

SekolahPro mendukung dua mode deployment dengan **satu codebase**:

| Mode | Trigger | Perilaku |
|------|---------|----------|
| `DEPLOYMENT_MODE=single` | Self-hosted (satu institusi) | Skip tenant selection, `tenant_id` dari config |
| `DEPLOYMENT_MODE=saas` | Multi-tenant SaaS | Two-phase JWT, user pilih tenant/company |

---

## 2. Konvensi Database Schema

### Tabel Master

Hanya memiliki `tenant_id` dan `company_id` (scope tertinggi yang relevan):

```sql
CREATE TABLE academic_years (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    company_id  UUID NOT NULL REFERENCES companies(id),
    ...
);
```

### Tabel Operasional

Memiliki hierarki penuh sesuai kebutuhan domain:

```sql
CREATE TABLE koperasi_transactions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id    UUID NOT NULL,
    company_id   UUID NOT NULL,
    branch_id    UUID NOT NULL,
    warehouse_id UUID,   -- nullable jika tidak relevan
    ...
);
```

### Index Wajib per Scope Level

```sql
CREATE INDEX idx_{table}_tenant    ON {table}(tenant_id);
CREATE INDEX idx_{table}_company   ON {table}(tenant_id, company_id);
-- Tambahkan branch/warehouse index jika domain memerlukan
```

### Row-Level Security via Repository

Repository secara eksplisit mengambil scope dari context:

```go
func (r *repo) ListByScope(ctx context.Context, p Pagination) ([]domain.Entity, error) {
    scope := middleware.ScopeFromContext(ctx)
    // scope.TenantID selalu ada
    // scope.CompanyID, BranchID, WarehouseID bisa nullable
}
```

Tidak ada query yang boleh melewatkan filter `tenant_id`.

---

## 3. Two-Phase JWT Authentication

### Phase 1 — Authentication Token

Diterbitkan setelah login berhasil, sebelum user memilih tenant/company:

```json
{
  "sub":     "user-uuid",
  "phase":   1,
  "email":   "user@sekolah.sch.id",
  "tenants": [
    {
      "tenant_id": "...",
      "companies": [
        { "company_id": "...", "name": "SMP Al-Hikmah" }
      ]
    }
  ],
  "exp": 1714000000   // TTL pendek: 5 menit
}
```

### Phase 2 — Operational Token

Diterbitkan setelah user memilih company. Berisi full scope dan permissions:

```json
{
  "sub":          "user-uuid",
  "phase":        2,
  "tenant_id":    "tenant-uuid",
  "company_id":   "company-uuid",
  "roles":        ["homeroom_teacher"],
  "permissions":  ["students:read", "attendance:write", "grades:write", "rapor:generate"],
  "teacher_id":   "teacher-uuid",
  "class_room_ids": ["class-uuid-1", "class-uuid-2"],
  "exp":          1714003600   // TTL: 1 jam
}
```

### Scope Middleware

```go
// Setiap request yang masuk dicek Phase 2 token-nya
// Scope dimasukkan ke context dan tersedia di semua layer
scope := ScopeContext{
    TenantID:    claims.TenantID,
    CompanyID:   claims.CompanyID,   // nullable
    BranchID:    claims.BranchID,    // nullable
    UserID:      claims.Subject,
    Roles:       claims.Roles,
    Permissions: claims.Permissions,
}
ctx := context.WithValue(r.Context(), ScopeContextKey, scope)
```

---

## 4. Dual-Mode Institution Type

SekolahPro melayani dua segmen institusi dengan **dua field type terpisah** per tenant:

```
school_type: "general" | "islamic"
coop_type:   "general" | "islamic"
```

### 4 Kombinasi yang Didukung

| school_type | coop_type | Konteks |
|-------------|-----------|---------|
| `general` | `general` | Sekolah umum + Koperasi konvensional |
| `general` | `islamic` | Sekolah umum + BMT |
| `islamic` | `general` | Pondok Pesantren + Koperasi konvensional |
| `islamic` | `islamic` | Pondok Pesantren + BMT (**target pasar utama**) |

### Immutability

`school_type` dan `coop_type` **tidak dapat diubah** setelah tenant aktif. Alasan:
- Perubahan mode berdampak pada business rule, kalkulasi finansial, dan data historis
- Migrasi data antar mode berisiko tinggi

Field ini di-set **satu kali saat onboarding tenant** dan di-enforce di application layer.

### Terminologi per Mode

Perbedaan terminologi di-handle melalui label mapping, bukan hardcode di UI:

```
school_type = "general"  → { student: "Siswa",  class: "Kelas",   teacher: "Guru" }
school_type = "islamic"  → { student: "Santri", class: "Halaqah", teacher: "Ustadz" }

coop_type = "general"    → { interest: "Bunga",      loan: "Pinjaman" }
coop_type = "islamic"    → { interest: "Bagi Hasil", loan: "Pembiayaan" }
```

### Feature Flag per Mode

Fitur eksklusif diaktifkan berdasarkan type, bukan konfigurasi manual:

```
school_type = "islamic" → enable: [tahfidz_module, diniyah_schedule, kitab_tracking]
coop_type   = "islamic" → enable: [akad_syariah, bagi_hasil_calculator, zakat_module]
```

### Business Rule Dispatch (Strategy Pattern)

Logika bisnis yang berbeda antar mode di-handle via strategy pattern:

```go
type ProfitCalculator interface {
    Calculate(principal, rate, period float64) float64
}

// general: bunga konvensional
type ConventionalCalculator struct{}

// islamic: bagi hasil / margin murabahah
type IslamicCalculator struct{}
```

Service memilih strategy berdasarkan `coop_type` dari tenant context.

---

## 5. Entitas Fondasi Domain Sekolah

### Academic Years (Tahun Ajaran)

Tahun ajaran adalah **unit waktu fundamental** — hampir semua domain student memiliki FK ke `academic_year_id`. Tanpa tabel ini, tidak ada domain student yang bisa diimplementasi.

```
Lifecycle: planning → active → closed
```

Aturan penting:
- Hanya **satu tahun ajaran aktif** per company pada satu waktu (enforced oleh partial unique index di DB)
- Tahun ajaran `closed` bersifat immutable — tidak bisa diedit
- Saat aktivasi tahun baru, tahun lama otomatis di-set `is_active = false, status = 'closed'`

Domain yang reference `academic_year_id`:
`students`, `class_rooms`, `student_academics`, `student_attendances`, `student_grades`, `student_invoices`, `student_rapor`, `student_placements`, `extracurriculars`, dan 2+ domain lainnya.

Schema key fields:
```sql
name          VARCHAR(20)  -- "2025/2026"
code          VARCHAR(10)  -- "2526" (untuk invoice numbering)
start_date    DATE         -- awal tahun ajaran (biasanya Juli)
end_date      DATE         -- akhir tahun ajaran (biasanya Juni)
semester1_start/end DATE   -- tanggal eksplisit per semester
semester2_start/end DATE
is_active     BOOLEAN      -- partial unique constraint
status        VARCHAR(20)  -- planning | active | closed
```

### Class Rooms (Kelas)

Kelas adalah **unit organisasi utama** operasional sekolah. Kelas di-scope per tahun ajaran — kelas VII-A 2025 berbeda entitas dari VII-A 2026.

Aturan penting:
- Nama kelas unik per company per tahun ajaran
- `current_count` adalah denormalisasi jumlah siswa saat ini (mencegah COUNT query)
- Fitur **carry-forward**: duplikasi kelas dari tahun lama ke tahun baru dengan reset count dan wali kelas

Schema key fields:
```sql
academic_year_id UUID NOT NULL     -- scope per tahun ajaran
homeroom_teacher_id UUID           -- wali kelas (nullable jika belum diassign)
name             VARCHAR(20)       -- "VII-A", "Halaqah VII-A" (Islamic)
grade_level      VARCHAR(5)        -- "1"-"12" (CHECK constraint)
capacity         INT DEFAULT 36    -- batas siswa
current_count    INT DEFAULT 0     -- denormalized count
major            VARCHAR(30)       -- untuk SMA: ipa, ips, bahasa, agama
```

Terminologi Islamic: Kelas → Halaqah, Grade Level → Marhalah, Wali Kelas → Musyrif. Data structure sama, terminologi di-resolve di presentation layer.

### Teachers & Staff (Guru dan Tenaga Kependidikan)

Satu tabel `teachers` menampung semua jenis personel sekolah, dibedakan oleh `role`.

Role yang tersedia:
```
kepala_sekolah, wakasek_kurikulum, wakasek_kesiswaan, wakasek_sarana, wakasek_humas,
guru_mapel, guru_bk, guru_piket, admin_tu, bendahara, pustakawan, staff_umum
```

Employee type (status kepegawaian Indonesia):
```
pns, p3k, honorer, yayasan, kontrak
```

Field penting:
```sql
user_id       UUID     -- link ke tabel users (nullable — tidak semua staff punya akun)
nip           VARCHAR(18) -- Nomor Induk Pegawai (PNS/P3K), partial unique
nuptk         VARCHAR(16) -- Nomor Unik PTK dari Kemendikbud
signature_url TEXT     -- tanda tangan digital untuk rapor
```

Relasi guru ke mata pelajaran menggunakan junction table `teacher_subject_map` yang di-scope per tahun ajaran.

Terminologi Islamic: Guru → Ustadz/Ustadzah, Kepala Sekolah → Mudir, Guru BK → Murshid.

### Users & Roles (Authentication & Authorization)

Tiga tabel: `users` (akun login), `roles` (definisi role), `user_roles` (assignment).

```
users         — identitas login (email, password_hash)
roles         — definisi role per company + permission set (JSONB)
user_roles    — assignment user ke role (bisa multi-company)
```

Permission model: permission sebagai string array di JSONB (`roles.permissions`):
```json
["students:read", "attendance:write", "grades:write", "rapor:generate"]
```

Default system roles (di-seed otomatis saat company dibuat):
`admin`, `principal`, `vice_principal`, `homeroom_teacher`, `teacher`, `counselor`, `finance`, `staff`, `parent`

Scope-based filtering (selain role permissions):
```
Superadmin      → semua tenant
Admin           → tenant_id + company_id
Kepala Sekolah  → company_id (semua kelas)
Guru/Wali Kelas → company_id + class_room_ids (kelas sendiri)
Orang Tua       → company_id + student_ids (anak sendiri)
```

---

## 6. Entity Relationship Overview (Fondasi)

```
tenants
  └── companies
        ├── academic_years  (1 aktif per company)
        ├── teachers
        │     └── teacher_subject_map
        ├── class_rooms  ──FK──→ academic_years
        │                   └──FK──→ teachers (wali kelas)
        ├── users
        │     └── user_roles ──FK──→ roles
        │                     └──FK──→ teachers (opsional)
        └── [domain siswa — menggunakan semua entitas di atas]
```

---

## Key Decisions

1. **4-level hierarki sebagai scope universal** — setiap tabel di-filter minimal oleh `tenant_id`; tabel operasional menggunakan hierarki penuh sesuai kebutuhan domain

2. **Two-phase JWT** — Phase 1 untuk autentikasi, Phase 2 untuk operational scope. Phase 1 token berumur pendek (5 menit) untuk meminimalkan risiko bocor

3. **Dua field type terpisah** — `school_type` dan `coop_type` tidak terikat satu sama lain, menghasilkan 4 kombinasi valid

4. **Immutability institution type** — tidak bisa diubah setelah tenant aktif untuk mencegah inkonsistensi data historis

5. **Satu tabel teachers** — guru dan staff dalam satu tabel dengan `role` sebagai pembeda; `user_id` nullable karena tidak semua staff punya akun

6. **System roles di-seed otomatis** — sekolah tidak perlu setup roles dari nol; custom roles bisa ditambah di atas system roles

7. **Islamic adalah primary market** — onboarding flow dan demo materials harus mengutamakan kasus pesantren; fitur Islamic wajib ada di Phase 1

---

## Constraints & Implications

### Constraints

- Setiap query ke tabel apapun wajib menyertakan filter `tenant_id` — tidak ada pengecualian
- `school_type` dan `coop_type` tidak boleh diubah setelah tenant aktif — enforce di application layer dengan validasi eksplisit
- Hanya satu `academic_year` aktif per company — enforced oleh partial unique index di DB (tidak bisa diedit di application layer saja)
- Phase 1 JWT tidak boleh digunakan untuk mengakses resource apapun selain endpoint tenant selection

### Implications untuk Domain Baru

- Setiap tabel wajib memiliki `tenant_id` dan scope yang sesuai
- Setiap fitur yang berbeda antar mode (general vs islamic) wajib menggunakan strategy pattern, bukan `if mode == "islamic"` di business logic
- Terminologi UI harus diambil dari label mapping berbasis `school_type`/`coop_type`, bukan hardcoded
- Setiap domain yang reference `academic_year_id` atau `class_room_id` wajib mempertimbangkan SyncEngine registration

### Implications untuk Testing

- Semua test harus menggunakan tenant yang diisolasi — tidak boleh ada shared state antar test
- Test harus mencakup 4 kombinasi mode institution type untuk fitur yang terpengaruh
- Test autentikasi harus mencakup Phase 1 dan Phase 2 token secara terpisah
- Test tenant isolation: pastikan user tenant A tidak bisa mengakses data tenant B
