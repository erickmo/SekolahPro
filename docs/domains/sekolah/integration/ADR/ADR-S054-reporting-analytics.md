# ADR-S054: Reporting & Analytics Dashboard (Laporan & Analitik)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Reporting adalah kebutuhan inti setiap sekolah — dari kepala sekolah yang memantau KPI sekolah, guru yang melihat performa kelas, hingga Dinas Pendidikan yang meminta laporan periodik. SekolahPro menyimpan data dari puluhan domain (S001-S052), dan semua data tersebut perlu disajikan dalam bentuk laporan dan dashboard yang actionable.

Kebutuhan reporting:

1. **Laporan akademik**: Nilai rata-rata per kelas/mapel, ranking siswa, analisis butir soal, tren nilai per semester.
2. **Laporan kehadiran**: Rekap hadir/sakit/izin/alpha per kelas, per bulan, per semester — dari S008.
3. **Laporan keuangan**: Rekap SPP, tunggakan per kelas, pemasukan per bulan, aging report — dari S009.
4. **Laporan Dapodik**: Format yang sesuai requirement Dapodik — dari S055.
5. **Dashboard real-time**: Widget counter (jumlah siswa, guru, tunggakan), chart (tren kehadiran, pemasukan), dan alert.
6. **Custom report builder**: Admin bisa pilih field, filter, grouping untuk laporan ad-hoc.
7. **Export**: PDF (untuk cetak), Excel/XLSX (untuk olah data), CSV (untuk import ke sistem lain).
8. **Role-based access**: Kepsek melihat semua, guru melihat kelasnya, orang tua melihat anaknya.

### Mengapa CQRS Pattern (bukan Vernon)?

Reporting adalah **100% read operation** — tidak ada write ke entity laporan. Data berasal dari aggregasi lintas banyak domain (students, grades, attendance, finance). Vernon pattern (_rels/_data) tidak cocok karena:

- Tidak ada "entity laporan" yang perlu relationship tracking.
- Data sudah di-aggregate — bukan single entity read.
- Butuh **materialized views** untuk pre-compute aggregasi berat.
- Query pattern sangat berbeda dari CRUD: GROUP BY, SUM, AVG, COUNT, window functions.

## Decision

Menggunakan **CQRS Pattern** dengan **materialized views** untuk pre-computed reports dan **report configuration tables** untuk custom report builder. TIDAK menggunakan Vernon _rels/_data pattern.

### Table Schema

```sql
-- ============================================================
-- MATERIALIZED VIEWS — Pre-computed report data
-- ============================================================

-- Rekap nilai per kelas per mata pelajaran per semester
CREATE MATERIALIZED VIEW mv_grade_summary AS
SELECT
    si.tenant_id,
    si.company_id,
    si.academic_year_id,
    si.semester,
    scp.class_id,
    sgd.subject_id,
    COUNT(DISTINCT sgd.student_id)          AS student_count,
    AVG(sgd.final_score)                    AS avg_score,
    MIN(sgd.final_score)                    AS min_score,
    MAX(sgd.final_score)                    AS max_score,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY sgd.final_score) AS median_score,
    COUNT(CASE WHEN sgd.final_score >= 75 THEN 1 END) AS pass_count,
    COUNT(CASE WHEN sgd.final_score < 75 THEN 1 END)  AS fail_count
FROM subject_grade_details sgd
JOIN student_invoices si ON si.student_id = sgd.student_id
JOIN student_class_placements scp ON scp.student_id = sgd.student_id
    AND scp.academic_year_id = si.academic_year_id
WHERE sgd.deleted_at IS NULL
    AND scp.deleted_at IS NULL
GROUP BY si.tenant_id, si.company_id, si.academic_year_id, si.semester,
         scp.class_id, sgd.subject_id;

CREATE UNIQUE INDEX idx_mv_grade_summary
    ON mv_grade_summary (tenant_id, company_id, academic_year_id, semester, class_id, subject_id);

-- Rekap kehadiran per kelas per bulan
CREATE MATERIALIZED VIEW mv_attendance_summary AS
SELECT
    da.tenant_id,
    da.company_id,
    da.academic_year_id,
    scp.class_id,
    EXTRACT(YEAR FROM da.attendance_date)::INT   AS year,
    EXTRACT(MONTH FROM da.attendance_date)::INT  AS month,
    COUNT(DISTINCT da.student_id)                AS student_count,
    COUNT(*)                                     AS total_records,
    COUNT(CASE WHEN da.status = 'hadir' THEN 1 END)  AS hadir_count,
    COUNT(CASE WHEN da.status = 'sakit' THEN 1 END)  AS sakit_count,
    COUNT(CASE WHEN da.status = 'izin' THEN 1 END)   AS izin_count,
    COUNT(CASE WHEN da.status = 'alpha' THEN 1 END)  AS alpha_count,
    ROUND(COUNT(CASE WHEN da.status = 'hadir' THEN 1 END)::NUMERIC / NULLIF(COUNT(*), 0) * 100, 2) AS attendance_rate
FROM daily_attendances da
JOIN student_class_placements scp ON scp.student_id = da.student_id
    AND scp.academic_year_id = da.academic_year_id
WHERE da.deleted_at IS NULL
    AND scp.deleted_at IS NULL
GROUP BY da.tenant_id, da.company_id, da.academic_year_id, scp.class_id,
         EXTRACT(YEAR FROM da.attendance_date), EXTRACT(MONTH FROM da.attendance_date);

CREATE UNIQUE INDEX idx_mv_attendance_summary
    ON mv_attendance_summary (tenant_id, company_id, academic_year_id, class_id, year, month);

-- Rekap keuangan per bulan per jenis tagihan
CREATE MATERIALIZED VIEW mv_finance_summary AS
SELECT
    si.tenant_id,
    si.company_id,
    si.academic_year_id,
    si.fee_type_id,
    si.period_year,
    si.period_month,
    COUNT(*)                               AS invoice_count,
    SUM(si.total_amount)                   AS total_billed,
    SUM(si.paid_amount)                    AS total_paid,
    SUM(si.total_amount - si.paid_amount)  AS total_outstanding,
    COUNT(CASE WHEN si.status = 'paid' THEN 1 END)    AS paid_count,
    COUNT(CASE WHEN si.status = 'partial' THEN 1 END) AS partial_count,
    COUNT(CASE WHEN si.status IN ('unpaid', 'overdue') THEN 1 END) AS unpaid_count
FROM student_invoices si
WHERE si.deleted_at IS NULL
GROUP BY si.tenant_id, si.company_id, si.academic_year_id, si.fee_type_id,
         si.period_year, si.period_month;

CREATE UNIQUE INDEX idx_mv_finance_summary
    ON mv_finance_summary (tenant_id, company_id, academic_year_id, fee_type_id, period_year, period_month);

-- Dashboard counters (real-time stats per sekolah)
CREATE MATERIALIZED VIEW mv_school_dashboard AS
SELECT
    s.tenant_id,
    s.company_id,
    COUNT(DISTINCT s.id) FILTER (WHERE s.status = 'active')   AS active_students,
    COUNT(DISTINCT t.id) FILTER (WHERE t.status = 'active')   AS active_teachers,
    COUNT(DISTINCT c.id)                                        AS total_classes,
    COALESCE(SUM(inv.total_amount - inv.paid_amount) FILTER (WHERE inv.status IN ('unpaid', 'partial', 'overdue')), 0) AS total_tunggakan
FROM students s
LEFT JOIN teachers t ON t.tenant_id = s.tenant_id AND t.company_id = s.company_id AND t.deleted_at IS NULL
LEFT JOIN classes c ON c.tenant_id = s.tenant_id AND c.company_id = s.company_id AND c.deleted_at IS NULL
LEFT JOIN student_invoices inv ON inv.tenant_id = s.tenant_id AND inv.company_id = s.company_id AND inv.deleted_at IS NULL
WHERE s.deleted_at IS NULL
GROUP BY s.tenant_id, s.company_id;

CREATE UNIQUE INDEX idx_mv_school_dashboard
    ON mv_school_dashboard (tenant_id, company_id);

-- ============================================================
-- REPORT CONFIGURATION — Custom report builder
-- ============================================================

-- Template laporan (pre-built + custom)
CREATE TABLE report_templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(200) NOT NULL,
    code            VARCHAR(50) NOT NULL,
    description     TEXT,
    category        VARCHAR(30) NOT NULL,

    -- Konfigurasi
    source_type     VARCHAR(30) NOT NULL,
    query_config    JSONB NOT NULL DEFAULT '{}',
    column_config   JSONB NOT NULL DEFAULT '[]',
    filter_config   JSONB NOT NULL DEFAULT '[]',
    group_config    JSONB NOT NULL DEFAULT '[]',
    sort_config     JSONB NOT NULL DEFAULT '[]',

    -- Display
    chart_type      VARCHAR(20),
    chart_config    JSONB,

    -- Access
    is_system       BOOLEAN NOT NULL DEFAULT false,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    allowed_roles   JSONB NOT NULL DEFAULT '[]',

    -- Metadata
    created_by      UUID NOT NULL,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_report_template_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_report_category CHECK (category IN ('akademik', 'kehadiran', 'keuangan', 'dapodik', 'kepegawaian', 'sarana', 'custom')),
    CONSTRAINT chk_source_type CHECK (source_type IN ('mv_grade_summary', 'mv_attendance_summary', 'mv_finance_summary', 'mv_school_dashboard', 'custom_query')),
    CONSTRAINT chk_chart_type CHECK (chart_type IS NULL OR chart_type IN ('bar', 'line', 'pie', 'area', 'table', 'counter', 'heatmap'))
);

CREATE INDEX idx_report_template_tenant ON report_templates (tenant_id, company_id);
CREATE INDEX idx_report_template_category ON report_templates (category);
CREATE INDEX idx_report_template_active ON report_templates (is_active) WHERE is_active = true;

-- Dashboard widget configuration
CREATE TABLE dashboard_widgets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    name            VARCHAR(100) NOT NULL,
    widget_type     VARCHAR(20) NOT NULL,
    description     TEXT,

    -- Konfigurasi
    report_template_id UUID,
    data_source     VARCHAR(30) NOT NULL,
    query_config    JSONB NOT NULL DEFAULT '{}',
    display_config  JSONB NOT NULL DEFAULT '{}',

    -- Layout
    position_x      INT NOT NULL DEFAULT 0,
    position_y      INT NOT NULL DEFAULT 0,
    width           INT NOT NULL DEFAULT 4,
    height          INT NOT NULL DEFAULT 3,

    -- Access
    target_role     VARCHAR(30) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_widget_type CHECK (widget_type IN ('counter', 'bar_chart', 'line_chart', 'pie_chart', 'area_chart', 'table', 'trend', 'alert')),
    CONSTRAINT chk_widget_source CHECK (data_source IN ('mv_grade_summary', 'mv_attendance_summary', 'mv_finance_summary', 'mv_school_dashboard', 'custom')),
    CONSTRAINT chk_widget_role CHECK (target_role IN ('super_admin', 'kepsek', 'wakil_kepsek', 'guru', 'wali_kelas', 'tata_usaha', 'bendahara', 'orang_tua', 'siswa'))
);

CREATE INDEX idx_dashboard_widget_tenant ON dashboard_widgets (tenant_id, company_id);
CREATE INDEX idx_dashboard_widget_role ON dashboard_widgets (target_role);
CREATE INDEX idx_dashboard_widget_active ON dashboard_widgets (is_active) WHERE is_active = true;

-- Report generation history / export log
CREATE TABLE report_exports (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    report_template_id UUID NOT NULL,
    requested_by    UUID NOT NULL,

    -- Export config
    format          VARCHAR(10) NOT NULL,
    filters_applied JSONB NOT NULL DEFAULT '{}',
    parameters      JSONB NOT NULL DEFAULT '{}',

    -- Result
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    file_url        TEXT,
    file_size_bytes BIGINT,
    row_count       INT,

    -- Timing
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    error_message   TEXT,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_export_format CHECK (format IN ('pdf', 'xlsx', 'csv')),
    CONSTRAINT chk_export_status CHECK (status IN ('pending', 'generating', 'completed', 'error', 'expired'))
);

CREATE INDEX idx_report_export_tenant ON report_exports (tenant_id, company_id);
CREATE INDEX idx_report_export_template ON report_exports (report_template_id);
CREATE INDEX idx_report_export_user ON report_exports (requested_by);
CREATE INDEX idx_report_export_status ON report_exports (status) WHERE status IN ('pending', 'generating');

-- ============================================================
-- REFRESH FUNCTIONS — Materialized view refresh
-- ============================================================

-- Function to refresh all materialized views
-- Called by cron job (every 15 minutes) or after bulk operations
-- In production: use pg_cron or application-level scheduler

-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_grade_summary;
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_attendance_summary;
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_finance_summary;
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_school_dashboard;
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| Materialized views | Bukan regular views | Aggregasi berat (SUM, AVG, COUNT) di jutaan baris — materialized view di-refresh periodik untuk performa |
| `REFRESH CONCURRENTLY` | Unique index wajib | Agar view bisa di-refresh tanpa locking read queries |
| `query_config` | JSONB di report_templates | Konfigurasi query flexible — field selection, filters, grouping disimpan sebagai JSON |
| `column_config` | JSONB array | Kolom yang ditampilkan + label + format (currency, date, percentage) |
| `allowed_roles` | JSONB array | Role-based access — kepsek sees all, guru sees own class |
| `is_system` | BOOLEAN | Laporan bawaan (akademik, kehadiran, keuangan) tidak bisa dihapus user |
| `format` | VARCHAR(10) | 3 format export: PDF, XLSX, CSV |
| `expires_at` | TIMESTAMPTZ | File export dihapus setelah expired — storage management |
| `position_x/y, width, height` | INT di widgets | Grid layout untuk dashboard — drag-and-drop positioning |

### Materialized View Refresh Strategy

```
┌─────────────────────────────────────────────┐
│ Refresh Strategy                             │
├─────────────────────────────────────────────┤
│ mv_school_dashboard   → every 5 minutes     │
│ mv_attendance_summary → every 15 minutes    │
│ mv_finance_summary    → every 15 minutes    │
│ mv_grade_summary      → every 30 minutes    │
│                                             │
│ + On-demand refresh after bulk operations:  │
│   - Bulk attendance input                   │
│   - Bulk payment recording                  │
│   - Grade finalization per class            │
└─────────────────────────────────────────────┘
```

### Role-Based Report Access Matrix

| Report Category | Kepsek | Wakil Kepsek | Guru/Wali Kelas | Tata Usaha | Bendahara | Orang Tua |
|---|---|---|---|---|---|---|
| Akademik (semua kelas) | Ya | Ya | Hanya kelasnya | Tidak | Tidak | Hanya anaknya |
| Kehadiran (semua) | Ya | Ya | Hanya kelasnya | Ya | Tidak | Hanya anaknya |
| Keuangan (semua) | Ya | Tidak | Tidak | Ya | Ya | Hanya anaknya |
| Dapodik | Ya | Ya | Tidak | Ya | Tidak | Tidak |
| Kepegawaian | Ya | Tidak | Tidak | Ya | Tidak | Tidak |
| Custom | Per konfigurasi | Per konfigurasi | Per konfigurasi | Per konfigurasi | Per konfigurasi | Tidak |

### API Endpoints

```
# Pre-built Reports
GET    /api/v1/reports/academic/grade-summary            — Rekap nilai per kelas/mapel
GET    /api/v1/reports/academic/ranking                  — Ranking siswa per kelas
GET    /api/v1/reports/attendance/summary                — Rekap kehadiran per kelas/bulan
GET    /api/v1/reports/attendance/trend                  — Tren kehadiran per bulan
GET    /api/v1/reports/finance/summary                   — Rekap keuangan per bulan
GET    /api/v1/reports/finance/overdue                   — Daftar tunggakan (aging report)
GET    /api/v1/reports/finance/collection-rate           — Persentase tertagih per bulan

# Dashboard
GET    /api/v1/dashboard/widgets                         — Widget config per role user
GET    /api/v1/dashboard/counters                        — Real-time counters (from mv_school_dashboard)
GET    /api/v1/dashboard/charts/{widget_id}              — Chart data per widget

# Report Templates (Custom Report Builder)
GET    /api/v1/report-templates                          — List templates
POST   /api/v1/report-templates                          — Buat custom report
PUT    /api/v1/report-templates/{id}                     — Update template
DELETE /api/v1/report-templates/{id}                     — Hapus custom template (system templates protected)
POST   /api/v1/report-templates/{id}/preview             — Preview report (max 100 rows)

# Dashboard Widgets
GET    /api/v1/dashboard-widgets                         — List widgets per role
POST   /api/v1/dashboard-widgets                         — Tambah widget
PUT    /api/v1/dashboard-widgets/{id}                    — Update widget (termasuk position)
DELETE /api/v1/dashboard-widgets/{id}                    — Hapus widget

# Export
POST   /api/v1/report-exports                           — Request export (async)
GET    /api/v1/report-exports/{id}                       — Cek status export
GET    /api/v1/report-exports/{id}/download              — Download file
GET    /api/v1/report-exports                            — Riwayat export

# Materialized View Admin
POST   /api/v1/admin/reports/refresh                     — Force refresh materialized views (admin only)
GET    /api/v1/admin/reports/refresh-status               — Status last refresh
```

## Consequences

### Positive

- **High performance**: Materialized views pre-compute aggregasi berat — dashboard load dalam milliseconds, bukan seconds.
- **CQRS clean separation**: Read path (reports) completely decoupled dari write path (CRUD operations).
- **Flexible**: Custom report builder memungkinkan sekolah membuat laporan sesuai kebutuhan spesifik mereka.
- **Role-based**: Setiap role hanya melihat data yang relevan — kepsek sees all, guru sees own class.
- **Export friendly**: PDF untuk cetak, XLSX untuk olah data, CSV untuk import ke sistem lain.
- **Scalable**: Materialized views di-refresh async — tidak impact write performance.

### Negative / Trade-offs

- **Stale data**: Materialized views bukan real-time — data bisa delay 5-30 menit tergantung refresh interval.
- **Refresh cost**: REFRESH MATERIALIZED VIEW CONCURRENTLY memakan resource — perlu scheduling yang bijak.
- **Storage overhead**: Materialized views menduplikasi data — perlu monitoring disk usage.
- **Custom query security**: Report builder yang mengizinkan custom query perlu SQL injection protection yang ketat — hanya allow predefined fields dan operators.
- **No Vernon pattern**: Report tables tidak punya _rels/_data — relationship resolution dilakukan di query layer, bukan denormalized cache.

## Alternatives Considered

### 1. Real-time aggregation tanpa materialized views
- Ditolak: query GROUP BY + SUM di jutaan baris terlalu lambat untuk dashboard — user expects < 1 second load time.

### 2. External BI tool (Metabase, Grafana)
- Ditolak untuk MVP: menambah operational complexity. Materialized views + custom report builder cukup untuk kebutuhan sekolah. BI tool bisa jadi enhancement untuk enterprise plan.

### 3. Vernon pattern untuk report templates
- Ditolak: report templates bukan domain entity dengan rich relationships — simple CRUD table cukup. Tidak perlu _rels/_data overhead.

### 4. Event sourcing untuk analytics
- Ditolak: terlalu complex untuk MVP. Materialized views memberikan 80% benefit dengan 20% complexity. Event sourcing bisa ditambahkan nanti untuk advanced analytics.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `ReportTemplateDescriptor.TableName()` | — | `"report_templates"` |
| U02 | `DashboardWidgetDescriptor.TableName()` | — | `"dashboard_widgets"` |
| U03 | `ReportExportDescriptor.TableName()` | — | `"report_exports"` |
| U04 | Validate rejects invalid `category` | `"marketing"` | Error: invalid category |
| U05 | Validate rejects invalid `chart_type` | `"radar"` | Error: invalid chart_type |
| U06 | Validate rejects invalid `format` | `"doc"` | Error: invalid format |
| U07 | Validate rejects invalid `widget_type` | `"gauge"` | Error: invalid widget_type |
| U08 | Validate rejects invalid `target_role` | `"murid"` | Error: invalid target_role |
| U09 | Validate role access — kepsek sees akademik | role=kepsek, category=akademik | Allowed |
| U10 | Validate role access — orang_tua blocked from keuangan semua | role=orang_tua, category=keuangan, scope=all | Denied |

### Integration Tests

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Materialized view grade_summary populated | Insert grades, refresh view | Data aggregated correctly |
| I02 | Materialized view attendance_summary correct | Insert attendance, refresh | Counts match (hadir, sakit, izin, alpha) |
| I03 | Materialized view finance_summary correct | Insert invoices + payments, refresh | total_billed, total_paid, total_outstanding correct |
| I04 | Dashboard counters from mv_school_dashboard | Refresh view | active_students, active_teachers, total_tunggakan correct |
| I05 | Create custom report template | POST with query_config | 201, created |
| I06 | Preview custom report | POST /preview | 200, max 100 rows |
| I07 | System template cannot be deleted | DELETE system template | 403/422 |
| I08 | Export to PDF | POST export with format=pdf | 201, async job started |
| I09 | Export to XLSX | POST export with format=xlsx | 201, file generated |
| I10 | Export to CSV | POST export with format=csv | 201, file generated |
| I11 | Export status tracking | GET /report-exports/{id} | Status transitions: pending → generating → completed |
| I12 | Dashboard widget layout | POST widget with position | 201, grid position saved |
| I13 | Role-based filtering — guru sees own class | GET grade-summary as guru | Only own class data returned |
| I14 | Role-based filtering — orang_tua sees own child | GET academic report as parent | Only own child data returned |
| I15 | Category CHECK enforced | INSERT template with `category = 'marketing'` | DB error |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Materialized views tenant-scoped | Query mv_grade_summary with tenant filter | Only own tenant data |
| I17 | Cannot access other tenant's report templates | GET templates from other tenant | 404 |
| I18 | Export files tenant-isolated | Download other tenant's export | 404 |
