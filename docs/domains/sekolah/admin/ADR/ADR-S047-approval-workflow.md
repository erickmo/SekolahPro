# ADR-S047: Approval Workflow Engine (Mesin Persetujuan Universal)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Banyak domain di SekolahPro memerlukan proses persetujuan bertingkat sebelum suatu entitas menjadi efektif. Saat ini setiap domain memiliki potensi membangun approval logic sendiri — ini menyebabkan duplikasi kode, inkonsistensi UX, dan kesulitan maintenance.

Domain-domain yang memerlukan approval:

| Domain | ADR | Flow | Approvers |
|--------|-----|------|-----------|
| Lesson Plans (RPP) | S024 | guru → wakasek_kurikulum | 2-level |
| Leave Request (Izin) | S030 | guru → kepsek (simple) **atau** guru → wakasek → kepsek → yayasan (multi-level) | 2-4 level |
| Correspondence (Surat Keluar) | S046 | admin_tu → kepsek | 2-level |
| Budget / RKAS | S050 | bendahara → kepsek → komite → dinas | 4-level |
| Facility Booking | S041 | requester → admin/wakasek_sarana | 2-level |
| Student Transfer | S016 | admin → kepsek | 2-level |
| Procurement | (future) | pengaju → bendahara → kepsek | 3-level |

Masalah tanpa engine terpusat:

1. **Duplikasi**: Setiap domain membangun state machine sendiri (draft → pending → approved/rejected).
2. **Inkonsistensi**: Approval di S024 punya 3 state, di S050 punya 5 state — UX membingungkan.
3. **Tidak ada delegasi**: Kepsek sedang dinas luar tidak bisa mendelegasikan approval ke wakasek.
4. **Tidak ada SLA tracking**: Tidak tahu apakah suatu permintaan sudah menunggu terlalu lama.
5. **Audit trail tersebar**: History approval tersebar di banyak tabel — sulit di-audit.

### Mengapa Vernon Pattern (dengan catatan)?

- Read-heavy: dashboard approval, inbox pending items, audit trail.
- Generic entity: approval_requests dan approval_steps adalah entity tersendiri yang di-query lintas domain.
- Relasi ke requestor, approver, dan entity asal — cocok untuk _data denormalisasi.
- Write terjadi saat submit/approve/reject — bukan transaksional tinggi.
- Eventually consistent acceptable — domain asal bisa diupdate async via event.

> **C-Suite CTO Review Note (2026-04-15):**
> Vernon `_data` memberikan benefit untuk inbox display (denormalisasi requestor, entity title,
> current approver) tanpa JOIN. Namun `approval_steps` mengalami **state transition yang cepat**
> (waiting → active → approved/rejected/delegated), dan setiap transisi memicu Vernon sync ke `_data`.
> Jika volume approval tinggi (>100 requests/hari), sync overhead bisa menjadi bottleneck.
> **Rekomendasi:** Gunakan Vernon untuk `approval_requests` (inbox read-heavy, benefit `_data` jelas).
> Untuk `approval_steps`, evaluasi saat implementasi — jika bottleneck terjadi, migrasi ke CQRS murni
> dengan JOIN saat read (steps per request selalu sedikit, JOIN murah).
> `approval_chains` (config, rarely written) tetap Vernon.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `approval_chains` (konfigurasi chain per entity_type), `approval_requests` (request persetujuan), dan `approval_steps` (langkah per-level dalam request).

Engine ini bersifat **generic** — tidak mengandung business logic domain manapun. Domain menggunakan engine ini via event-driven integration:

1. Domain membuat `approval_request` dengan `entity_type` + `entity_id`.
2. Engine mengelola state machine dan routing ke approver yang tepat berdasarkan `approval_chain` config.
3. Saat approved/rejected, engine mengirim event — domain mendengarkan dan mengupdate statusnya sendiri.

### Table Schema

```sql
-- Konfigurasi approval chain per entity_type
CREATE TABLE approval_chains (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas chain
    name            VARCHAR(255) NOT NULL,
    entity_type     VARCHAR(50) NOT NULL,
    description     TEXT,

    -- Konfigurasi
    steps           JSONB NOT NULL DEFAULT '[]',
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Kondisi (opsional — chain berbeda berdasarkan kondisi)
    condition_rules JSONB NOT NULL DEFAULT '{}',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_chain_entity_type CHECK (entity_type IN (
        'lesson_plan', 'leave_request', 'correspondence',
        'budget_plan', 'facility_booking', 'student_transfer',
        'procurement', 'custom'
    )),
    CONSTRAINT uq_chain_tenant_entity UNIQUE (tenant_id, company_id, entity_type, name)
);

-- Indexes
CREATE INDEX idx_chain_tenant_company ON approval_chains (tenant_id, company_id);
CREATE INDEX idx_chain_entity_type ON approval_chains (entity_type);
CREATE INDEX idx_chain_active ON approval_chains (is_active) WHERE is_active = true;
CREATE INDEX idx_chain_rels ON approval_chains USING GIN (_rels);
CREATE INDEX idx_chain_data ON approval_chains USING GIN (_data);


-- Request persetujuan (1 per entity submission)
CREATE TABLE approval_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Referensi ke entity asal (polymorphic)
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       UUID NOT NULL,
    entity_title    VARCHAR(500) NOT NULL,

    -- Chain yang digunakan
    chain_id        UUID NOT NULL,

    -- Requestor
    requestor_id    UUID NOT NULL,
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Status keseluruhan
    status          VARCHAR(25) NOT NULL DEFAULT 'pending',
    current_step    INT NOT NULL DEFAULT 1,
    total_steps     INT NOT NULL,

    -- Completion
    completed_at    TIMESTAMPTZ,
    completed_by    UUID,
    final_remarks   TEXT,

    -- SLA tracking
    sla_deadline    TIMESTAMPTZ,
    is_overdue      BOOLEAN NOT NULL DEFAULT false,

    -- Cancellation
    cancelled_at    TIMESTAMPTZ,
    cancelled_by    UUID,
    cancel_reason   TEXT,

    -- Priority
    priority        VARCHAR(10) NOT NULL DEFAULT 'normal',

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_request_entity_type CHECK (entity_type IN (
        'lesson_plan', 'leave_request', 'correspondence',
        'budget_plan', 'facility_booking', 'student_transfer',
        'procurement', 'custom'
    )),
    CONSTRAINT chk_request_status CHECK (status IN (
        'pending', 'in_review', 'approved', 'rejected',
        'revision_requested', 'cancelled', 'expired'
    )),
    CONSTRAINT chk_request_priority CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    CONSTRAINT uq_request_entity UNIQUE (entity_type, entity_id, status)
);

-- Indexes
CREATE INDEX idx_request_tenant_company ON approval_requests (tenant_id, company_id);
CREATE INDEX idx_request_entity ON approval_requests (entity_type, entity_id);
CREATE INDEX idx_request_chain ON approval_requests (chain_id);
CREATE INDEX idx_request_requestor ON approval_requests (requestor_id);
CREATE INDEX idx_request_status ON approval_requests (status);
CREATE INDEX idx_request_current_step ON approval_requests (current_step) WHERE status IN ('pending', 'in_review');
CREATE INDEX idx_request_sla ON approval_requests (sla_deadline) WHERE is_overdue = false AND status IN ('pending', 'in_review');
CREATE INDEX idx_request_submitted ON approval_requests (submitted_at);
CREATE INDEX idx_request_rels ON approval_requests USING GIN (_rels);
CREATE INDEX idx_request_data ON approval_requests USING GIN (_data);


-- Langkah approval per request (1 row per step per request)
CREATE TABLE approval_steps (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Parent
    request_id      UUID NOT NULL,

    -- Step identity
    step_order      INT NOT NULL,
    step_name       VARCHAR(100) NOT NULL,

    -- Approver assignment
    approver_role   VARCHAR(50) NOT NULL,
    approver_id     UUID,
    delegated_to    UUID,
    delegated_by    UUID,
    delegated_at    TIMESTAMPTZ,

    -- Decision
    action          VARCHAR(25),
    remarks         TEXT,
    decided_at      TIMESTAMPTZ,

    -- Status
    status          VARCHAR(25) NOT NULL DEFAULT 'waiting',

    -- SLA per step
    sla_hours       INT,
    sla_deadline    TIMESTAMPTZ,
    is_overdue      BOOLEAN NOT NULL DEFAULT false,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_step_action CHECK (
        action IS NULL OR action IN ('approve', 'reject', 'revision_requested', 'delegate', 'skip')
    ),
    CONSTRAINT chk_step_status CHECK (status IN (
        'waiting', 'active', 'approved', 'rejected',
        'revision_requested', 'delegated', 'skipped'
    )),
    CONSTRAINT uq_step_request_order UNIQUE (request_id, step_order)
);

-- Indexes
CREATE INDEX idx_step_tenant_company ON approval_steps (tenant_id, company_id);
CREATE INDEX idx_step_request ON approval_steps (request_id);
CREATE INDEX idx_step_approver ON approval_steps (approver_id) WHERE status = 'active';
CREATE INDEX idx_step_delegated ON approval_steps (delegated_to) WHERE delegated_to IS NOT NULL;
CREATE INDEX idx_step_status ON approval_steps (status);
CREATE INDEX idx_step_sla ON approval_steps (sla_deadline) WHERE is_overdue = false AND status = 'active';
CREATE INDEX idx_step_rels ON approval_steps USING GIN (_rels);
CREATE INDEX idx_step_data ON approval_steps USING GIN (_data);
```

### Field Design Rationale

**approval_chains:**

| Field | Keputusan | Alasan |
|---|---|---|
| `entity_type` | VARCHAR(50), CHECK | Polymorphic reference — mendukung semua domain yang butuh approval |
| `steps` | JSONB array | Konfigurasi langkah: `[{"order":1,"role":"wakasek_kurikulum","sla_hours":48},{"order":2,"role":"kepsek","sla_hours":72}]` |
| `condition_rules` | JSONB | Rule engine sederhana: chain berbeda berdasarkan kondisi (misal: cuti >3 hari = 3-level, cuti <=3 hari = 2-level) |
| `is_active` | BOOLEAN | Soft-disable chain tanpa delete — bisa switch chain config |

**approval_requests:**

| Field | Keputusan | Alasan |
|---|---|---|
| `entity_type` + `entity_id` | Polymorphic reference | Bisa merujuk ke entity apapun — lesson_plan, leave_request, budget_plan, dst |
| `entity_title` | VARCHAR(500) | Denormalisasi judul entity — agar inbox approval bisa ditampilkan tanpa JOIN ke tabel asal |
| `current_step` | INT | Pointer ke step yang sedang aktif — memudahkan query "siapa yang perlu approve sekarang" |
| `total_steps` | INT | Total langkah — di-set saat request dibuat berdasarkan chain config |
| `sla_deadline` | TIMESTAMPTZ | Deadline keseluruhan request — jika terlewat, auto-flag `is_overdue` |
| `priority` | VARCHAR(10) | Urgent request bisa di-highlight di inbox approver |

**approval_steps:**

| Field | Keputusan | Alasan |
|---|---|---|
| `step_order` | INT | Urutan sequential: 1, 2, 3... — menentukan siapa approve duluan |
| `approver_role` | VARCHAR(50) | Role yang harus approve (misal: `kepsek`) — resolved ke user_id saat step menjadi active |
| `approver_id` | UUID, nullable | NULL saat waiting, di-resolve saat step menjadi active berdasarkan role + tenant context |
| `delegated_to` | UUID, nullable | Kepsek bisa delegate ke wakasek — delegasi tercatat untuk audit |
| `action` | VARCHAR(25), nullable | NULL sampai approver mengambil keputusan |
| `sla_hours` | INT, nullable | SLA per step — dari chain config. Step 1 mungkin 48 jam, step 2 mungkin 72 jam |

### Vernon Relationships

**approval_requests:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `chain` | belongs_to | Tidak | Chain config jarang dimuat di list view |
| `requestor` | belongs_to | **Ya** | Selalu perlu tahu siapa yang mengajukan |
| `steps` | has_many | Tidak | Dimuat di detail view, bukan list |

**approval_steps:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `request` | belongs_to | **Ya** | Selalu perlu konteks request induk |
| `approver` | belongs_to | **Ya** | Selalu perlu tahu siapa approver |
| `delegated_to_user` | belongs_to | Tidak | Hanya jika ada delegasi |

### _rels / _data Structure

**approval_chains:**

```json
{
  "_rels": {},
  "_data": {
    "steps_summary": [
      { "order": 1, "role": "wakasek_kurikulum", "sla_hours": 48 },
      { "order": 2, "role": "kepsek", "sla_hours": 72 }
    ],
    "usage_count": 145
  }
}
```

**approval_requests:**

```json
{
  "_rels": {
    "chain_id": "018f...",
    "requestor_id": "018f..."
  },
  "_data": {
    "requestor": {
      "id": "018f...",
      "full_name": "Ibu Dewi",
      "role": "guru"
    },
    "chain": {
      "id": "018f...",
      "name": "Lesson Plan Approval"
    },
    "current_approver": {
      "id": "018f...",
      "full_name": "Pak Budi",
      "role": "wakasek_kurikulum"
    },
    "steps_progress": [
      { "order": 1, "role": "wakasek_kurikulum", "status": "approved", "decided_at": "2026-04-10T09:00:00Z" },
      { "order": 2, "role": "kepsek", "status": "active", "decided_at": null }
    ]
  }
}
```

**approval_steps:**

```json
{
  "_rels": {
    "request_id": "018f...",
    "approver_id": "018f..."
  },
  "_data": {
    "request": {
      "id": "018f...",
      "entity_type": "lesson_plan",
      "entity_title": "RPP Matematika Kelas VII - Semester Genap",
      "status": "in_review"
    },
    "approver": {
      "id": "018f...",
      "full_name": "Pak Budi",
      "role": "wakasek_kurikulum"
    }
  }
}
```

### Chain Configuration (steps JSONB)

```json
// Contoh: Leave Request — multi-level berdasarkan durasi
{
  "name": "Leave Request (>3 days)",
  "entity_type": "leave_request",
  "condition_rules": {
    "field": "duration_days",
    "operator": "gt",
    "value": 3
  },
  "steps": [
    {
      "order": 1,
      "name": "Wakasek Review",
      "role": "wakasek",
      "sla_hours": 24,
      "can_delegate": true
    },
    {
      "order": 2,
      "name": "Kepala Sekolah Approval",
      "role": "kepsek",
      "sla_hours": 48,
      "can_delegate": true
    },
    {
      "order": 3,
      "name": "Yayasan Approval",
      "role": "yayasan",
      "sla_hours": 72,
      "can_delegate": false
    }
  ]
}

// Contoh: Budget Plan (RKAS) — 4-level
{
  "name": "RKAS Approval",
  "entity_type": "budget_plan",
  "steps": [
    { "order": 1, "name": "Bendahara Review", "role": "bendahara", "sla_hours": 72 },
    { "order": 2, "name": "Kepala Sekolah Approval", "role": "kepsek", "sla_hours": 120 },
    { "order": 3, "name": "Komite Sekolah Review", "role": "komite", "sla_hours": 168 },
    { "order": 4, "name": "Dinas Pendidikan Approval", "role": "dinas", "sla_hours": 336 }
  ]
}
```

### State Machine

```
                ┌──────────────────────────────────────────────────┐
                │                                                  │
    ┌───────┐   │  ┌───────────┐   ┌──────────┐   ┌───────────┐  │
    │ draft │──>│  │  pending   │──>│in_review │──>│ approved  │  │
    └───────┘   │  └───────────┘   └──────────┘   └───────────┘  │
                │       │              │  │                        │
                │       │              │  └──>┌───────────┐       │
                │       │              │      │ rejected  │       │
                │       │              │      └───────────┘       │
                │       │              │                           │
                │       │              └──>┌───────────────────┐  │
                │       │                  │revision_requested │──┘
                │       │                  └───────────────────┘
                │       │
                │       └──>┌───────────┐
                │           │ cancelled │
                │           └───────────┘
                │
                └──>┌───────────┐
                    │  expired  │  (SLA terlewat tanpa action)
                    └───────────┘
```

Transisi status:

| From | To | Trigger |
|------|----|---------|
| draft | pending | Requestor submit |
| pending | in_review | Step 1 approver mulai review |
| in_review | approved | Semua step approved (current_step > total_steps) |
| in_review | rejected | Salah satu step reject |
| in_review | revision_requested | Approver minta revisi — kembali ke requestor |
| revision_requested | pending | Requestor re-submit setelah revisi |
| pending/in_review | cancelled | Requestor cancel request |
| pending/in_review | expired | SLA terlewat — cron job menandai |

### Event-Driven Integration

Domain menggunakan approval engine via events:

```
Domain → ApprovalEngine:
  ApprovalRequested { entity_type, entity_id, chain_id, requestor_id }

ApprovalEngine → Domain:
  ApprovalCompleted { entity_type, entity_id, status: "approved", approvals: [...] }
  ApprovalRejected  { entity_type, entity_id, status: "rejected", remarks: "..." }
  RevisionRequested { entity_type, entity_id, remarks: "..." }

ApprovalEngine → Notification (S044):
  StepAssigned      { approver_id, entity_title, sla_deadline }
  StepOverdue       { approver_id, entity_title, overdue_hours }
  RequestCompleted  { requestor_id, entity_title, final_status }
```

### API Endpoints

```
# Chain Configuration (admin)
POST   /api/v1/approval-chains                              — Buat chain baru
GET    /api/v1/approval-chains                              — List chains (filter: entity_type)
GET    /api/v1/approval-chains/{id}                         — Detail chain
PUT    /api/v1/approval-chains/{id}                         — Update chain config
DELETE /api/v1/approval-chains/{id}                         — Soft delete chain

# Approval Requests (domain submits)
POST   /api/v1/approval-requests                            — Submit request baru
GET    /api/v1/approval-requests                            — List requests (filter: entity_type, status, requestor)
GET    /api/v1/approval-requests/{id}                       — Detail request + steps
PUT    /api/v1/approval-requests/{id}/cancel                — Cancel request (by requestor)
PUT    /api/v1/approval-requests/{id}/resubmit              — Resubmit after revision

# Approval Actions (approver)
GET    /api/v1/approval-steps/inbox                         — Inbox: pending steps for current user
GET    /api/v1/approval-steps/inbox/count                   — Badge count for UI
PUT    /api/v1/approval-steps/{id}/approve                  — Approve step
PUT    /api/v1/approval-steps/{id}/reject                   — Reject step (with remarks)
PUT    /api/v1/approval-steps/{id}/request-revision         — Request revision (with remarks)
PUT    /api/v1/approval-steps/{id}/delegate                 — Delegate to another user

# Audit & Reporting
GET    /api/v1/approval-requests/{id}/history               — Full audit trail
GET    /api/v1/approval-requests/statistics                  — Statistics: avg approval time, overdue count, etc.
GET    /api/v1/approval-requests/overdue                     — List overdue requests (for admin dashboard)
```

### Delegation Flow

```
1. Kepsek mendapat step "active" di inbox-nya.
2. Kepsek sedang dinas luar → PUT /approval-steps/{id}/delegate { "delegated_to": "{wakasek_id}" }
3. Step di-update: delegated_to = wakasek_id, delegated_by = kepsek_id, status tetap "active".
4. Wakasek sekarang melihat step ini di inbox-nya.
5. Wakasek approve → action = "approve", decided_at = now.
6. Audit trail mencatat: "Approved by Wakasek (delegated from Kepsek)".
```

## Consequences

### Positive

- **DRY**: Satu engine melayani semua domain — tidak ada duplikasi approval logic.
- **Konsisten**: Semua approval punya UX yang sama: inbox, approve/reject button, history.
- **Konfigurabel**: Chain config di JSONB — admin bisa menambah/mengubah step tanpa deploy ulang.
- **Delegasi**: Kepsek bisa mendelegasikan approval tanpa menunggu kembali dari dinas luar.
- **SLA tracking**: Manajemen sekolah bisa melihat bottleneck: "Siapa yang paling sering terlambat approve?"
- **Audit trail lengkap**: Setiap action tercatat — siapa, kapan, keputusan apa, komentar apa.
- **Event-driven**: Domain tidak tightly-coupled ke engine — komunikasi via event.
- **Skalabel**: Menambah domain baru (misal: procurement) hanya perlu tambah chain config + entity_type.

### Negative / Trade-offs

- **Complexity**: Generic engine lebih kompleks dari approval logic per-domain sederhana.
- **Polymorphic reference**: `entity_type` + `entity_id` tidak punya FK constraint — perlu application-level validation.
- **Eventual consistency**: Domain baru tahu status approval setelah event diterima — ada jeda sebelum entity berubah status.
- **Chain config management**: Admin perlu UI untuk mengkonfigurasi chain — perlu dibangun.
- **Condition rules**: Rule engine sederhana (field + operator + value) mungkin tidak cukup untuk kondisi kompleks — perlu dievaluasi.
- **Performance**: Inbox query lintas entity_type bisa lambat jika volume tinggi — mitigasi dengan proper indexing dan pagination.

## Alternatives Considered

### 1. Approval logic per domain (setiap domain punya state machine sendiri)
- Ditolak: duplikasi code, inkonsisten UX, tidak ada delegasi terpusat, audit trail tersebar. Untuk 7+ domain ini akan menjadi maintenance nightmare.

### 2. Menggunakan library workflow engine (Temporal, Cadence)
- Ditolak: over-engineering untuk approval use case sekolah. Workflow engine general-purpose terlalu kompleks dan memerlukan infrastruktur tambahan. Approval sekolah cukup sequential (bukan parallel/conditional branching yang kompleks).

### 3. Simple 2-column (approved_by, approved_at) di setiap tabel
- Ditolak: tidak mendukung multi-level approval, delegasi, SLA tracking, atau audit trail per step.

### 4. BPMN-based workflow engine
- Deferred: BPMN terlalu general dan visual editor-nya mahal. Jika di masa depan approval menjadi sangat kompleks (parallel approval, conditional branching), baru dipertimbangkan.

### 5. Approval steps sebagai JSONB array di approval_requests (bukan tabel terpisah)
- Ditolak: sulit query "semua step yang pending untuk user X" (inbox query). Tabel terpisah memungkinkan indexing dan query efisien per approver.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `Descriptor.TableName()` returns correct name | — | `"approval_requests"` |
| U02 | `DefaultRels()` returns autoloaded rels | — | `requestor` |
| U03 | Validate rejects invalid entity_type | `{ "entity_type": "unknown" }` | Error: entity_type not in allowed list |
| U04 | Validate rejects invalid status | `{ "status": "half_approved" }` | Error: status not in allowed list |
| U05 | Resolve chain for entity_type | `{ "entity_type": "lesson_plan", "tenant_id": "..." }` | Returns matching active chain |
| U06 | Resolve chain with condition | `{ "entity_type": "leave_request", "duration_days": 5 }` | Returns chain with condition `gt 3` |
| U07 | State transition: pending → in_review | Valid transition | No error |
| U08 | State transition: approved → pending | Invalid transition | Error: invalid state transition |
| U09 | Calculate SLA deadline | `{ "sla_hours": 48, "submitted_at": "2026-04-15T09:00:00Z" }` | `"2026-04-17T09:00:00Z"` |
| U10 | Validate step_order is sequential | `[1, 2, 4]` (gap) | Error: step_order must be sequential |
| U11 | Delegation validation | Delegate to self | Error: cannot delegate to self |
| U12 | Delegation validation | Delegate when can_delegate=false | Error: delegation not allowed for this step |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Submit approval request | POST with valid entity_type + chain_id | 201, steps created based on chain config |
| I02 | Inbox shows active step | GET inbox for step-1 approver | 200, includes the new request |
| I03 | Approve step 1 → step 2 activates | PUT approve step 1 | 200, step 1 = approved, step 2 = active, request.current_step = 2 |
| I04 | Approve final step → request approved | PUT approve last step | 200, request.status = approved, completed_at set |
| I05 | Reject at any step → request rejected | PUT reject step 2 | 200, request.status = rejected, remaining steps skipped |
| I06 | Request revision → back to requestor | PUT request-revision on step 1 | 200, request.status = revision_requested |
| I07 | Resubmit after revision | PUT resubmit | 200, request.status = pending, steps reset to waiting |
| I08 | Cancel request | PUT cancel | 200, request.status = cancelled |
| I09 | Cannot approve non-active step | PUT approve step 2 while step 1 is active | 403, step not active |
| I10 | Delegation flow | PUT delegate to wakasek | 200, delegated_to set, wakasek sees in inbox |
| I11 | Delegated user approves | Wakasek PUT approve on delegated step | 200, approved, audit shows delegation |
| I12 | SLA overdue detection | Create request with sla=1 hour, wait/simulate | is_overdue = true after deadline |
| I13 | Badge count endpoint | GET inbox/count | 200, `{ "pending": 5, "overdue": 1 }` |
| I14 | Statistics endpoint | GET statistics | 200, avg_approval_time, overdue_count per entity_type |

### Integration Tests — Event Integration

| # | Test Case | Action | Expected |
|---|---|---|---|
| I15 | ApprovalCompleted event fired | Approve final step | Event `ApprovalCompleted` published with entity_type + entity_id |
| I16 | ApprovalRejected event fired | Reject at any step | Event `ApprovalRejected` published |
| I17 | StepAssigned notification | New step becomes active | Event `StepAssigned` published to notification system (S044) |
| I18 | Domain receives ApprovalCompleted | Lesson plan request approved | Lesson plan status updated to "approved" (via event handler) |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | UserUpdated syncs to requests | Update requestor name | `_data.requestor.full_name` updated |
| I20 | UserUpdated syncs to steps | Update approver name | `_data.approver.full_name` updated |
| I21 | RequestUpdated syncs to steps | Update request status | `_data.request.status` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's requests | GET with wrong tenant scope | 404 |
| I23 | Cannot approve other tenant's step | PUT approve with cross-tenant step_id | 403/404, scope violation |
| I24 | Chain config is tenant-scoped | Tenant A has 3-level chain, Tenant B has 2-level | Each uses own config independently |
