# 32 — Report Refresh & Data Freshness Policy

**Status**: Draft  
**Perspective**: Data/Engineering + Ops + Product  
**Primary Inputs**: `15-school-analytics-reporting-spec.md`, `19-reporting-export-compliance-spec.md`, `11-observability-deployment-spec.md`

## Purpose

Mendefinisikan kebijakan refresh report, freshness indicator, dan ekspektasi stale data agar pengguna memahami kualitas data yang sedang dilihat dan tim operasional tahu kapan harus bertindak.

## Freshness Principles

- tidak semua report butuh near-real-time
- freshness harus terlihat oleh consumer report
- stale yang diketahui lebih baik daripada angka tanpa konteks waktu
- refresh policy harus mempertimbangkan cost, load, dan urgensi bisnis

## Freshness Classes

### Real-Time / Near-Real-Time
- dashboard kritikal operasional
- status pembayaran / settlement support tertentu

### Scheduled Short Interval
- summary operasional harian
- alert-oriented aggregates

### Periodic Batch
- laporan semester, executive packs, Dapodik prep packs

## Policy Requirements

- setiap report/export menampilkan `data_as_of` atau indikator setara
- materialized view refresh cadence harus terdokumentasi
- stale threshold breach harus punya alert atau flag operasional
- manual refresh/forced rebuild hanya untuk role tertentu dan harus diaudit

## Acceptance Criteria

- `AC-FUNC`: pengguna dapat memahami seberapa segar data report yang mereka lihat
- `AC-DATA`: freshness metadata dapat ditelusuri ke source refresh process
- `AC-AUDIT`: manual refresh dan stale-threshold override tercatat
- `AC-INT`: export dan scheduled report menghormati freshness policy yang sama
- `AC-NFR`: refresh strategy menjaga keseimbangan antara freshness dan sistem load

## Open Questions

- target freshness final per report family
- kapan forced refresh diizinkan pada jam operasional
