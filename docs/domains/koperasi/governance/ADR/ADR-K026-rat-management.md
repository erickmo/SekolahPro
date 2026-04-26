# ADR-K026: RAT (Rapat Anggota Tahunan) Management

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Rapat Anggota Tahunan (RAT) adalah kekuasaan tertinggi dalam koperasi sesuai UU No. 25/1992 Pasal 46-50. RAT memutuskan hal-hal strategis seperti:

- Pengesahan laporan keuangan tahunan
- Pembagian SHU (Sisa Hasil Usaha)
- Pemilihan/pemberhentian Pengurus dan Pengawas
- Perubahan AD/ART
- Rencana kerja dan anggaran tahun berikutnya
- Penggabungan/pecahan/dissolution koperasi

Saat ini K016 (SHU) menyebut RAT sebagai approval untuk distribusi SHU, tetapi tidak ada ADR yang mengatur proses RAT itu sendiri: undangan, quorum, voting, notulensi, dan implementasi keputusan.

Tanpa sistem RAT yang terstruktur:
- Keputusan koperasi bisa dianulir karena prosedur tidak sesuai AD/ART
- Tidak ada audit trail untuk keputusan strategis
- Pemilihan pengurus/pengawas tidak terdokumentasi dengan baik

## Decision

### 1. RAT Types

```
rat_meeting_type:
├── ANNUAL            ← RAT tahunan (wajib, sekali per tahun)
├── SPECIAL           ← Rapat Anggota Luar Biasa (RALB) — urgent matters
├── JOINT             ← Rapat Bersama (untuk penggabungan koperasi)
└── DISSOLUTION       ← Rapat Anggota untuk pembubaran
```

**Kapan RALB bisa dipanggil:**
- Atas permintaan ≥ 25% anggota
- Atas keputusan Pengawas (jika Pengurus melanggar AD/ART)
- Atas keputusan Rapat Pengurus (ada hal urgent)
- Atas perintah pemerintah (Kanwil Koperasi)

### 2. Meeting Data Model

```
rat_meeting:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Meeting Info ──
├── meeting_type          ENUM (annual, special, joint, dissolution)
├── meeting_number        VARCHAR         ← "RAT-2026-001", "RALB-2026-001"
├── title                 VARCHAR
├── description           TEXT
│
├── ── Schedule ──
├── meeting_date          DATE
├── meeting_time          TIME
├── meeting_location      VARCHAR         ← Alamat lengkap
├── meeting_mode          ENUM (offline, online, hybrid)
├── online_link           VARCHAR (nullable, URL video call)
│
├── ── Quorum ──
├── total_eligible_members INT            ← Jumlah anggota berhak hadir
├── quorum_required       INT            ← Minimum hadir (biasanya >50%)
├── actual_attendees      INT            ← Jumlah yang hadir
├── quorum_met            BOOLEAN
│
├── ── Status ──
├── status                ENUM (draft, invitation_sent, in_progress,
│                                completed, cancelled, rescheduled)
├── cancelled_reason      TEXT (nullable)
├── rescheduled_to_id     UUID (nullable, FK → rat_meeting)
│
├── ── Documents ──
├── agenda_doc_id         UUID (nullable, FK → document)
├── minutes_doc_id        UUID (nullable, FK → document)
├── financial_report_id   UUID (nullable, FK → document)
├── shu_proposal_id       UUID (nullable, FK → document)
│
├── ── Audit ──
├── convened_by           UUID (FK → governance_position, Ketua)
├── secretary_id          UUID (FK → governance_position, Sekretaris)
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

### 3. Meeting Agenda & Proposals

```
rat_agenda_item:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── meeting_id            UUID (FK → rat_meeting)
│
├── ── Agenda ──
├── agenda_number         INT            ← Urutan agenda
├── category              ENUM (opening, report, election, shu,
│                                ad_art_change, budget, other, closing)
├── title                 VARCHAR
├── description           TEXT
│
├── ── Proposal ──
├── proposal_type         ENUM (information, discussion, voting, election)
├── proposer_id           UUID (FK → nasabah, nullable)
├── proposal_document_id  UUID (nullable, FK → document)
│
├── ── Discussion ──
├── discussion_notes      TEXT (nullable)
├── discussion_duration   INT (nullable, minutes)
│
├── ── Result ──
├── result                ENUM (approved, rejected, deferred, no_vote)
├── result_notes          TEXT
├── vote_for              INT (nullable)
├── vote_against          INT (nullable)
├── vote_abstain          INT (nullable)
│
├── ── Follow-up ──
├── followup_action       TEXT (nullable)
├── followup_deadline     DATE (nullable)
├── followup_assignee_id  UUID (nullable, FK → governance_position)
├── followup_status       ENUM (pending, in_progress, completed)
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

### 4. Attendance & Voting

```
rat_attendance:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── meeting_id            UUID (FK → rat_meeting)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Attendance ──
├── attendance_type       ENUM (present, proxy, absent_apology, absent_no_info)
├── proxy_for_id          UUID (nullable, FK → nasabah, jika hadir sebagai kuasa)
├── proxy_document_id     UUID (nullable, FK → document, surat kuasa)
│
├── ── Voting Rights ──
├── voting_right          ENUM (full, limited, none)
│   # full: anggota aktif dengan hak suara penuh
│   # limited: anggota tertentu dengan hak suara terbatas (misal staff)
│   # none: tidak berhak voting (hadir sebagai tamu/observer)
│
├── checked_in_at         TIMESTAMPTZ (nullable)
├── checked_in_by         UUID (FK → user, nullable)
│
├── ── Audit ──
└── created_at            TIMESTAMPTZ

UNIQUE(meeting_id, nasabah_id)
```

```
rat_vote:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── meeting_id            UUID (FK → rat_meeting)
├── agenda_item_id        UUID (FK → rat_agenda_item)
├── voter_id              UUID (FK → nasabah)
│
├── ── Vote ──
├── vote                  ENUM (for, against, abstain)
├── vote_method           ENUM (show_of_hands, secret_ballot, electronic)
│
├── ── Audit ──
├── voted_at              TIMESTAMPTZ
└── created_at            TIMESTAMPTZ

UNIQUE(agenda_item_id, voter_id)
```

### 5. Election Management (Pemilihan Pengurus/Pengawas)

```
rat_election:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── meeting_id            UUID (FK → rat_meeting)
├── agenda_item_id        UUID (FK → rat_agenda_item)
│
├── ── Election Info ──
├── position_type         ENUM (ketua, sekretaris, bendahara,
│                                pengawas_ketua, pengawas_anggota,
│                                dps_ketua, dps_anggota)
├── term_start            DATE
├── term_end              DATE
├── vacancies             INT            ← Jumlah kursi kosong
│
├── ── Status ──
├── status                ENUM (nomination, voting, completed, failed)
├── election_method       ENUM (open, secret_ballot, acclamation)
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

```
rat_election_candidate:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── election_id           UUID (FK → rat_election)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Nomination ──
├── nominated_by_id       UUID (FK → nasabah, yang mengusulkan)
├── seconded_by_id        UUID (FK → nasabah, yang mendukung)
├── nomination_statement  TEXT
├── status                ENUM (proposed, accepted, withdrawn, elected, not_elected)
│
├── ── Results ──
├── vote_count            INT (nullable)
├── rank                  INT (nullable)
│
├── ── Audit ──
└── created_at            TIMESTAMPTZ
```

### 6. RAT Workflow

```
RAT Annual Workflow:
┌──────────────────────────────────────────────────┐
│ Phase 1: Preparation (H-60 s/d H-30)            │
│ ├── Pengurus menyusun laporan tahunan           │
│ ├── Pengawas menyusun laporan pemeriksaan        │
│ ├── Bendahara menyiapkan laporan keuangan        │
│ ├── SHU proposal disiapkan (ref K016)            │
│ ├── Rencana kerja & RKAP tahun berikutnya        │
│ └── Agenda draft disusun                         │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Phase 2: Invitation (H-14 s/d H-7)              │
│ ├── Surat undangan dikirim ke semua anggota      │
│ ├── Agenda & dokumen dilampirkan                 │
│ ├── Anggota bisa ajukan tambahan agenda          │
│ └── Deadline konfirmasi kehadiran                 │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Phase 3: Meeting Day                             │
│ ├── Registration & quorum check                  │
│ ├── Pembacaan & pengesahan notulens sebelumnya   │
│ ├── Laporan Pengurus (keuangan, operasional)     │
│ ├── Laporan Pengawas (audit findings)             │
│ ├── Diskusi & voting per agenda                  │
│ ├── SHU approval                                 │
│ ├── Pemilihan Pengurus/Pengawas (jika ada)       │
│ └── Penutupan                                    │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Phase 4: Post-Meeting (H+7 s/d H+30)            │
│ ├── Notulensi final disusun & disahkan           │
│ ├── Keputusan RAT didistribusikan ke anggota      │
│ ├── Follow-up actions assigned                    │
│ ├── Pemilihan results → update governance_position│
│ ├── Laporan ke Dinas Koperasi (jika wajib)       │
│ └── SHU distribution executed (ref K016)          │
└──────────────────────────────────────────────────┘
```

### 7. Quorum Rules

**Quorum requirements (sesuai UU Koperasi):**
- RAT pertama: > 50% anggota berhak hadir
- RAT kedua (jika quorum pertama tidak tercapai): > 33% anggota berhak hadir
- Untuk perubahan AD/ART: > 66% anggota berhak hadir
- Untuk dissolution: > 66% anggota berhak hadir

**Voting passing criteria:**
- Keputusan biasa: > 50% suara yang hadir
- Perubahan AD/ART: > 66% suara yang hadir
- Pemilihan: plurality (suara terbanyak menang)
- Dissolution: > 66% suara yang hadir

### 8. Minutes (Notulensi) Template

```
notulensi_structure:
├── Header
│   ├── Nama koperasi + nomor badan hukum
│   ├── Jenis rapat (RAT/RALB)
│   ├── Hari, tanggal, waktu, tempat
│   ├── Yang menghadiri: jumlah, representasi
│   └── Quorum: tercapai/tidak tercapai
│
├── Agenda Items (per item)
│   ├── Nomor agenda
│   ├── Topik
│   ├── Pembahasan ringkas
│   ├── Pendapat pro/kontra
│   ├── Hasil voting (if voted)
│   └── Keputusan final
│
├── Elections (if any)
│   ├── Posisi yang diperebutkan
│   ├── Calon-calon
│   ├── Hasil voting per kandidat
│   └── Yang terpilih
│
├── Closing
│   ├── Waktu penutupan
│   ├── Tanda tangan Ketua Rapat
│   ├── Tanda tangan Sekretaris
│   └── Tanda tangan Pengawas (sebagai saksi)
│
└── Attachments
    ├── Daftar hadir
    ├── Laporan keuangan
    ├── SHU proposal
    └── Dokumen pendukung lainnya
```

### 9. Vernon _rels dan _data Structure

**Meeting _rels:**
```json
{
  "tenant_id": "018f..."
}
```

**Meeting _data:**
```json
{
  "convened_by": {
    "id":        "018f...",
    "full_name": "Ahmad Fauzi",
    "position":  "Ketua"
  },
  "secretary": {
    "id":        "018f...",
    "full_name": "Siti Aminah",
    "position":  "Sekretaris"
  },
  "stats": {
    "total_eligible": 150,
    "attended":        98,
    "quorum_met":      true
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update nama nasabah di attendance/election records
- `ElectionCompletedEvent` → auto-create `governance_position` (ref K025) untuk yang terpilih
- `MeetingCompletedEvent` → trigger SHU distribution workflow (ref K016) jika SHU approved

### 10. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| DPS report | Tidak ada | **WAJIB** — laporan kepatuhan syariah |
| SHU terminology | Sisa Hasil Usaha | Sisa Hasil Usaha |
| Produk baru approval | RAT + Pengurus | RAT + Pengurus + DPS |
| Laporan keuangan | PSAK/SAK ETAP | PSAK Syariah |
| Pengawas | Standard Pengawas | Standard Pengawas + DPS |
| Pemilihan DPS | Tidak ada | Dipilih oleh RAT |
| Meeting prayer | Opsional | Doa Islami wajib di pembukaan |

### 11. Authorization — RBAC

| Permission | Ketua | Sekretaris | Bendahara | Pengawas | Anggota | Admin |
|---|---|---|---|---|---|---|
| Create RAT meeting | v | v (draft) | - | - | - | v |
| Send invitation | v | v | - | - | - | v |
| Confirm attendance | - | v | - | - | v | v |
| Add agenda item | v | v | - | v | v (propose) | v |
| Conduct voting | v | v (record) | - | v (witness) | v (vote) | - |
| Finalize minutes | v | v | - | v (approve) | - | v |
| View meeting details | v | v | v | v | v | v |
| View voting details | v | v | v | v | v (own) | v |
| Manage elections | v | v | - | v | - | v |
| Cancel/reschedule | v | - | - | v (approve) | - | v |
| View historical RAT | v | v | v | v | v | v |
