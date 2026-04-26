# ADR-013: Users & Roles (Authentication & Authorization)

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Sistem memerlukan **authentication** (siapa kamu?) dan **authorization** (apa yang boleh kamu lakukan?) untuk setiap pengguna. Berdasarkan ADR-004 (Multi-Tenant) dan existing domains, pengguna SekolahPro terdiri dari:

| Persona | Akses | Konteks |
|---------|-------|---------|
| **Superadmin** | Semua tenant | Platform operator (SekolahPro team) |
| **Admin Sekolah** | Satu tenant + company | Manajemen data sekolah penuh |
| **Kepala Sekolah** | Satu company | Approval, rapor finalize, laporan |
| **Wakasek** | Satu company (per bidang) | Kurikulum, kesiswaan, sarana, humas |
| **Guru** | Kelas yang diajar | Absensi, nilai, ekskul |
| **Wali Kelas** | Satu kelas | Rapor, absensi kelas, profil siswa |
| **Guru BK** | Kasus sendiri | Counseling records (privasi tinggi) |
| **Bendahara** | Finance module | SPP, tagihan, pembayaran |
| **Staff TU** | Admin module | Dokumen, surat, data master |
| **Orang Tua** | Data anak sendiri | View-only: rapor, absensi, SPP (future) |

Authorization harus mendukung:
- **Role-based**: Permission ditentukan per role.
- **Scope-based**: Data dibatasi per tenant + company (ADR-004).
- **Domain-level**: Beberapa domain (BK) punya access control lebih ketat.
- **Two-phase JWT**: Login → pilih tenant/company (ADR-004).

### Mengapa di Core ADR?

Users & Roles bukan domain student — ini **infrastructure** yang dibutuhkan semua domain. Setiap API endpoint memerlukan auth context: siapa user-nya, role-nya apa, tenant/company scope-nya apa.

## Decision

Menggunakan 3 tabel: `users` (akun login), `roles` (definisi role), dan `user_roles` (assignment role ke user). Permission disimpan di role sebagai JSONB permission set.

### Table Schema

```sql
-- Akun user (auth)
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- Identitas login
    email           VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    password_hash   TEXT NOT NULL,
    full_name       VARCHAR(255) NOT NULL,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    is_superadmin   BOOLEAN NOT NULL DEFAULT false,
    email_verified  BOOLEAN NOT NULL DEFAULT false,
    last_login_at   TIMESTAMPTZ,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_user_email UNIQUE (email)
);

-- Indexes
CREATE INDEX idx_user_email ON users (email);
CREATE INDEX idx_user_active ON users (is_active) WHERE is_active = true;

-- Definisi role per company
CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(50) NOT NULL,
    code            VARCHAR(30) NOT NULL,
    description     TEXT,

    -- Role type
    role_type       VARCHAR(20) NOT NULL,
    is_system       BOOLEAN NOT NULL DEFAULT false,

    -- Permissions (JSONB set)
    permissions     JSONB NOT NULL DEFAULT '[]',

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_role_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_role_type CHECK (role_type IN (
        'admin', 'principal', 'vice_principal',
        'teacher', 'homeroom_teacher', 'counselor',
        'finance', 'staff', 'parent'
    ))
);

-- Indexes
CREATE INDEX idx_role_tenant_company ON roles (tenant_id, company_id);

-- Assignment: user ↔ role (per company)
CREATE TABLE user_roles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id         UUID NOT NULL,
    role_id         UUID NOT NULL,
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Optional: link ke teacher/guardian
    teacher_id      UUID,
    guardian_id     UUID,

    -- Status
    is_active       BOOLEAN NOT NULL DEFAULT true,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_user_role UNIQUE (user_id, role_id, company_id)
);

-- Indexes
CREATE INDEX idx_user_role_user ON user_roles (user_id);
CREATE INDEX idx_user_role_role ON user_roles (role_id);
CREATE INDEX idx_user_role_company ON user_roles (tenant_id, company_id);
CREATE INDEX idx_user_role_teacher ON user_roles (teacher_id) WHERE teacher_id IS NOT NULL;
```

### Permission Model

Permissions disimpan di `roles.permissions` sebagai JSONB array:

```json
{
  "permissions": [
    "students:read",
    "students:write",
    "attendance:read",
    "attendance:write",
    "grades:read",
    "grades:write",
    "finance:read",
    "rapor:read",
    "rapor:generate",
    "rapor:finalize",
    "counseling:read",
    "counseling:write",
    "settings:read",
    "settings:write"
  ]
}
```

### Default System Roles

Saat company dibuat, system roles di-seed otomatis (`is_system = true`, tidak bisa dihapus):

| Role Code | Permissions | Catatan |
|-----------|-------------|---------|
| `admin` | `*` (semua) | Admin sekolah |
| `principal` | Semua read + approve + rapor:finalize | Kepala sekolah |
| `vice_principal` | Semua read + write di bidangnya | Per bidang |
| `homeroom_teacher` | students:read, attendance:*, grades:*, rapor:read+generate | Wali kelas |
| `teacher` | Scoped: grades:write (mapel sendiri), attendance:write (kelas sendiri) | Guru mapel |
| `counselor` | counseling:*, students:read (limited) | Guru BK |
| `finance` | finance:*, students:read (limited) | Bendahara |
| `staff` | students:read, documents:* | Staff TU |
| `parent` | Scoped: students:read (anak sendiri), rapor:read, finance:read | Orang tua |

### JWT Structure (Two-Phase, ADR-004)

**Phase 1 JWT** (setelah login, sebelum pilih tenant):
```json
{
  "sub": "user-uuid",
  "email": "guru@sekolah.sch.id",
  "type": "auth",
  "tenants": [
    { "tenant_id": "...", "companies": [{ "company_id": "...", "name": "SMP Al-Hikmah" }] }
  ]
}
```

**Phase 2 JWT** (setelah pilih company):
```json
{
  "sub": "user-uuid",
  "email": "guru@sekolah.sch.id",
  "type": "session",
  "tenant_id": "...",
  "company_id": "...",
  "roles": ["homeroom_teacher"],
  "permissions": ["students:read", "attendance:write", "grades:write", "rapor:generate"],
  "teacher_id": "...",
  "class_room_ids": ["..."]
}
```

### Scope-Based Data Filtering

Selain role-based permissions, data difilter berdasarkan **scope**:

```
Superadmin      → semua tenant
Admin           → tenant_id + company_id
Kepala Sekolah  → company_id (semua kelas)
Guru/Wali Kelas → company_id + class_room_ids (kelas sendiri)
Guru BK         → company_id + counselor_id (kasus sendiri)
Bendahara       → company_id + finance domain only
Orang Tua       → company_id + student_ids (anak sendiri)
```

### API Endpoints

```
# Auth
POST   /api/v1/auth/login                           — Login (returns Phase 1 JWT)
POST   /api/v1/auth/select-company                   — Pilih company (returns Phase 2 JWT)
POST   /api/v1/auth/refresh                          — Refresh token
POST   /api/v1/auth/logout                           — Invalidate token
POST   /api/v1/auth/change-password                  — Ganti password

# Users (Admin)
GET    /api/v1/users                                 — List users per company
POST   /api/v1/users                                 — Buat user + assign role
PUT    /api/v1/users/{id}                            — Update user
PUT    /api/v1/users/{id}/deactivate                 — Nonaktifkan user

# Roles
GET    /api/v1/roles                                 — List roles per company
POST   /api/v1/roles                                 — Buat custom role
PUT    /api/v1/roles/{id}                            — Update permissions
POST   /api/v1/users/{id}/roles                      — Assign role ke user
DELETE /api/v1/users/{id}/roles/{role_id}            — Remove role

# Profile
GET    /api/v1/me                                    — Current user profile
PUT    /api/v1/me                                    — Update own profile
```

## Consequences

### Positive

- **Flexible RBAC**: Role + permissions + scope memberi kontrol granular.
- **Multi-company**: Satu user bisa punya role berbeda di company berbeda.
- **System roles**: Default roles di-seed — sekolah tidak perlu setup dari nol.
- **Custom roles**: Admin bisa buat role custom dengan permission subset.
- **Teacher/Guardian link**: `user_roles.teacher_id` dan `guardian_id` menghubungkan auth ke domain data.
- **Two-phase JWT**: User pilih company context setelah login — mendukung multi-sekolah.

### Negative / Trade-offs

- **Permission string-based**: Permission sebagai string array di JSONB — tidak ada foreign key validation.
- **JWT size**: Phase 2 JWT bisa besar jika user punya banyak roles/permissions — perlu pruning.
- **No permission inheritance**: Parent role tidak cascade ke child — setiap role harus definisikan sendiri. Bisa ditambah nanti.
- **Password only**: MVP hanya email+password. OAuth (Google), SSO belum di-implement.
- **Orang tua login**: Perlu mekanisme invite/registration khusus (beda dari guru/staff).

## Alternatives Considered

### 1. Permission tabel terpisah (normalized RBAC)
- Deferred: JSONB permissions lebih sederhana untuk MVP. Full normalized RBAC bisa ditambah jika perlu >50 unique permissions.

### 2. External auth provider (Auth0, Supabase Auth)
- Ditolak: menambah dependency dan cost. Two-phase JWT custom sesuai ADR-004 tidak didukung out-of-box.

### 3. Satu tabel users+teachers (gabung)
- Ditolak: tidak semua user adalah teacher (orang tua, superadmin). Dan tidak semua teacher punya user account.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | Validate rejects empty email | `""` | Error |
| U02 | Validate rejects invalid email format | `"notanemail"` | Error |
| U03 | Validate rejects invalid `role_type` | `"janitor"` | Error |
| U04 | Permission check: user has permission | `permissions: ["students:read"]`, check `students:read` | true |
| U05 | Permission check: user lacks permission | `permissions: ["students:read"]`, check `students:write` | false |
| U06 | Permission check: wildcard `*` matches all | `permissions: ["*"]`, check `anything` | true |

### Integration Tests — Auth

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Login with valid credentials | POST /auth/login | 200, Phase 1 JWT |
| I02 | Login with wrong password | POST /auth/login | 401 |
| I03 | Select company | POST /auth/select-company | 200, Phase 2 JWT with roles+permissions |
| I04 | Select company not assigned | POST with wrong company | 403 |
| I05 | Refresh token | POST /auth/refresh | 200, new token |
| I06 | Access with expired token | GET /students | 401 |

### Integration Tests — Users & Roles

| # | Test Case | Action | Expected |
|---|---|---|---|
| I07 | Create user | POST /users | 201 |
| I08 | Unique email | Create 2 users same email | 409/422 |
| I09 | Assign role | POST /users/{id}/roles | 201 |
| I10 | Unique user+role+company | Assign same role twice | 409/422 |
| I11 | Create custom role | POST /roles with permissions | 201 |
| I12 | Cannot delete system role | DELETE system role | 422 |
| I13 | Deactivate user | PUT /deactivate | 200, is_active = false |
| I14 | Deactivated user cannot login | POST /auth/login | 403 |

### Integration Tests — Authorization

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | Admin can access all students | GET /students as admin | 200 |
| I16 | Teacher can only see own class | GET /students as teacher | 200, scoped to class_room_ids |
| I17 | Parent can only see own child | GET /students as parent | 200, scoped to student_ids |
| I18 | Guru BK can access counseling | GET /counseling-cases as counselor | 200 |
| I19 | Admin cannot access counseling | GET /counseling-cases as admin | 403 |
| I20 | Only principal can finalize rapor | POST /rapor/finalize as teacher | 403 |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I21 | User in tenant A cannot access tenant B | GET /students with wrong tenant | 404/403 |
| I22 | Superadmin can access all tenants | GET /students as superadmin | 200 |
