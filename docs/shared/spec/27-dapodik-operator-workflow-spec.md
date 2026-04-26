# 27 — Dapodik Operator Workflow Spec

**Status**: Draft  
**Perspective**: School Ops + Analyst + Engineering  
**Primary Inputs**: `ADR-S055`, `ADR-S054`, `09-onboarding-migration-spec.md`, `15-school-analytics-reporting-spec.md`

## Purpose

Mendefinisikan workflow operator Dapodik dari validasi data, batch preparation, export, import manual ke aplikasi Dapodik, sampai tracking hasil dan correction loop.

## Workflow Principles

- phase awal fokus `manual_export`, bukan ilusi full API sync
- operator harus bisa melihat error sebelum export
- correction loop harus cepat dan terarah
- status batch harus dapat ditelusuri per semester, per entity, dan per record bermasalah

## Operator Stages

### Stage 1 — Data Readiness Review
- cek NISN, NUPTK, NPSN, school profile, dan field wajib
- review mismatch dan missing data
- jalankan report/validation dashboard pre-export

### Stage 2 — Batch Preparation
- pilih academic year dan semester
- pilih entity scope
- generate / review batch draft
- validasi dan approval internal bila diperlukan

### Stage 3 — Manual Export
- generate file export sesuai mode dan format
- catat export artifact dan timestamp
- operator unduh dan import ke aplikasi Dapodik lokal

### Stage 4 — Import Result Tracking
- catat success, warning, invalid records, dan operator notes
- tandai records yang perlu correction
- rerun batch parsial atau ulang sesuai kebutuhan

### Stage 5 — Period Close
- simpan evidence export/import
- freeze reference snapshot bila perlu untuk audit semester
- update status kesiapan pelaporan sekolah

## Key Views Needed

- pre-validation dashboard
- per-record validation error list
- batch status summary
- entity mismatch report
- export history and artifact list

## Acceptance Criteria

- `AC-FUNC`: operator dapat menyelesaikan proses Dapodik tanpa double entry massal
- `AC-DATA`: field mapping, validation, dan record status dapat ditelusuri per batch
- `AC-AUTH`: hanya role operator/authorized admin yang dapat membuat atau approve batch
- `AC-AUDIT`: export artifact, approval, correction, dan completion note tercatat
- `AC-INT`: reporting views dan Dapodik batch records saling mendukung correction loop
- `AC-NFR`: batch validation dan export cukup cepat untuk deadline operasional semester

## Open Questions

- kapan partial re-export per entity/record diizinkan resmi
- evidence minimum apa yang wajib disimpan setelah import manual berhasil
