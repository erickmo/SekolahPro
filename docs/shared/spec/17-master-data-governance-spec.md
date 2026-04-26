# 17 — Master Data Governance Spec

**Status**: Draft  
**Perspective**: System Analyst + Data Governance + Engineering  
**Primary Inputs**: `ADR-010`, `ADR-011`, `ADR-012`, `ADR-013`, `02-capability-map.md`

## Purpose

Mendefinisikan tata kelola master data lintas platform agar academic years, class rooms, teachers/staff, users/roles, dan entitas fondasional lain tetap konsisten, terkontrol, dan dapat diandalkan oleh domain turunan.

## Governance Principles

- satu sumber kebenaran per master domain
- perubahan master data berdampak lintas domain harus dikelola eksplisit
- active/inactive lifecycle harus jelas
- historical accuracy lebih penting daripada overwrite yang menghapus konteks lama
- semua reference domain harus mengikuti ownership master data

## Core Master Domains

### Academic Years
- satu active academic year per company
- closed year bersifat immutable
- activation/closure berdampak ke operasional domain turunan

### Class Rooms
- scoped per academic year
- capacity, homeroom assignment, dan carry-forward harus tervalidasi
- perubahan kelas harus mempertahankan historical separation antar tahun ajaran

### Teachers & Staff
- HR master terpisah dari account auth, walau bisa terhubung
- role fungsional dan account permission tidak boleh disamakan secara buta
- lifecycle join/resign/leave mempengaruhi assignment operasional

### Users & Roles
- identity, permission, dan organizational scope dikelola terpisah namun terkait
- assignment role harus dapat diaudit dan dicabut dengan aman

## Governance Controls

- naming convention dan uniqueness rule per domain
- activation/deactivation rule
- carry-forward / copy-forward workflow yang aman
- reconciliation job untuk data denormalisasi seperti current_count atau assignment snapshot
- approval atau elevated permission untuk perubahan master data kritikal

## Change Impact Rules

- perubahan pada master data yang direferensikan banyak domain harus menghasilkan impact note atau event yang jelas
- sync lag dan affected domain harus dapat dipantau
- rename dan status change tidak boleh memutus referential understanding pada laporan/history

## Acceptance Criteria

- `AC-FUNC`: admin dapat mengelola master data inti tanpa menyebabkan state domain turunan menjadi ambigu
- `AC-DATA`: uniqueness, active-state, lifecycle, dan historical context terjaga
- `AC-AUTH`: hanya peran berwenang yang dapat mengubah master data kritikal
- `AC-AUDIT`: perubahan penting dan impact terhadap domain lain tercatat
- `AC-INT`: carry-forward, activation, dan sync propagation dapat diverifikasi hasilnya
- `AC-NFR`: perubahan master data fondasional tidak merusak operasional harian secara luas

## Open Questions

- domain master tambahan mana yang perlu dikelompokkan sebagai governed master set tahap berikutnya
- kapan approval wajib diterapkan untuk perubahan tertentu seperti closing academic year
