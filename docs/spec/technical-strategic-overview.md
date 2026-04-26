# Tinjauan Strategis Teknis SekolahPro — Perspektif CTO

> Dokumen ini adalah asesmen teknis mendalam untuk CTO dan Tech Lead.
> Dibuat berdasarkan Wave 1 documentation: Core Architecture, Data Layer Strategy, Multi-Tenant Design, Compliance, dan domain-domain utama.

---

## 1. Asesmen Technology Bet: Mengapa Go + CQRS + Vernon

### Pilihan Stack dan Rasionalisasi

| Keputusan Teknis | Pilihan | Penilaian |
|-----------------|---------|----------|
| Backend language | Go 1.22+ | **TEPAT** — Performa tinggi, concurrency native (goroutine), ekosistem matang untuk backend layanan |
| Architecture pattern | Clean Architecture + CQRS | **TEPAT** — Separasi concern yang kuat; command/query path dioptimasi independen |
| Read optimization | Vernon Denormalized Read-Cache | **TEPAT tapi BERISIKO** — Performa luar biasa (450ms → 12ms pada 500K rows), namun kompleksitas operasional tinggi |
| Primary key | UUID v7 | **TEPAT** — Time-sortable, globally unique, B-tree friendly; throughput INSERT 73% lebih tinggi vs UUID v4 |
| Event bus | NATS JetStream (prod) / InMemory (dev) | **TEPAT** — Lightweight, low-latency, sesuai untuk event volume koperasi/sekolah |
| Dependency injection | Uber FX | **TEPAT** — Lifecycle management otomatis, fail-fast pada invalid dependency graph |
| ORM | sqlc (code generation) | **TEPAT** — Type-safe SQL, tidak ada magic, query mudah di-audit |

### Risiko Utama Stack Vernon

Vernon Pattern adalah bet terbesar secara teknis. Benefit sangat nyata: query zero-JOIN pada dataset besar. Namun ada biaya tersembunyi:

**Biaya Operasional Vernon:**
- Setiap relasi baru **wajib** didaftarkan di `SyncRegistry` secara manual
- ALTER TABLE pada parent entity bisa memicu **cascade sync ke ratusan ribu row** secara bersamaan
- Debug eventual consistency di produksi jauh lebih sulit dari query JOIN biasa
- Onboarding engineer baru memerlukan waktu lebih lama untuk memahami `_data` vs tabel relasi

**Mitigasi yang Sudah Dirancang:**
- `_sync_status` flag untuk mendeteksi row yang belum ter-sync
- Dead Letter Queue (DLQ) untuk sync yang gagal
- Manual resync endpoint untuk admin
- Selective field sync (hanya field yang berubah)
- Batch coalescing 100ms untuk mencegah burst event

**Penilaian Akhir:** Vernon adalah pilihan yang tepat untuk SekolahPro, tetapi harus dimonitor ketat. SyncEngine health harus masuk sebagai KPI operasional Platform Reliability, bukan sekadar log background.

---

## 2. Asesmen Skalabilitas: Bisakah Ini Menangani 10.000 Sekolah?

### Arsitektur Saat Ini (Modular Monolith)

```
Monolith Go Backend
    ↓
PostgreSQL (write + Vernon read-cache)
    ↓
NATS JetStream (async events)
    ↓
Redis (optional: hot cache layer)
```

### Analisis Kapasitas per Komponen

| Komponen | Kapasitas Estimasi | Bottleneck? |
|---------|-------------------|------------|
| Go backend (single instance) | ~5.000 req/detik | Tidak dengan horizontal scaling |
| PostgreSQL (single primary) | ~50.000 TPS | Ya — pada > 10.000 sekolah aktif |
| NATS JetStream | Juta msg/detik | Tidak untuk skala ini |
| Vernon sync workers | Bergantung goroutine pool | Perlu tuning pada > 1.000 concurrent syncs |

### Skala Realistis dengan Arsitektur Saat Ini (2026)

- **0–500 sekolah:** Arsitektur saat ini cukup tanpa modifikasi signifikan
- **500–2.000 sekolah:** Perlu PgBouncer (connection pooling), Redis read cache layer, horizontal scaling app nodes
- **2.000–10.000 sekolah:** Perlu read replica PostgreSQL, partitioning tabel per tenant_id, kemungkinan pemisahan database per modul (Sekolah vs Koperasi)
- **> 10.000 sekolah:** Kemungkinan butuh sharding atau pemisahan microservices pada modul dengan beban tertinggi (Transaksi Koperasi, Laporan Analytics)

### Rekomendasi Roadmap Skalabilitas

```
2026 (Fase 1 — 0-500 sekolah):
  Prioritas: Fondasi stabil, bukan skalabilitas prematur
  Action: PgBouncer untuk connection pooling

2027 (Fase 2 — 500-2.000 sekolah):
  Action: PostgreSQL read replica
  Action: Redis cluster untuk Vernon hot cache
  Action: Horizontal app scaling (load balancer + stateless nodes)
  Action: Monitoring terpusat (Prometheus + Grafana)

2028+ (Fase 3 — > 2.000 sekolah):
  Evaluasi: Apakah perlu extract microservice untuk Transaksi Koperasi?
  Action: Database partitioning atau sharding jika perlu
```

**Panduan:** Jangan ekstrak microservice sebelum ada bottleneck yang teridentifikasi dengan data metrik. Monolith modular yang sehat lebih mudah di-debug dan di-deploy daripada microservices prematur.

---

## 3. Inventaris Technical Debt

### Debt yang Sudah Diidentifikasi (dari Wave 1 Docs)

| Item | Kategori | Biaya Jangka Panjang | Prioritas Resolusi |
|------|---------|---------------------|-------------------|
| Feature gating system belum terimplementasi | Arsitektur | **KRITIS** — Pro tier tidak bisa diluncurkan tanpa ini | P0 |
| Usage metering (hitung siswa aktif) belum ada | Arsitektur | **KRITIS** — Billing tidak bisa berjalan tanpa ini | P0 |
| Billing & invoice system belum ada ADR | Arsitektur | **KRITIS** — Revenue tidak bisa dimonetisasi | P0 |
| Consent flow UU PDP belum terimplementasi | Compliance | **KRITIS** — Data siswa tidak bisa diproses secara legal | P0 |
| Data retention scheduler belum ada | Compliance | **TINGGI** — Deadline Q4 2026 | P1 |
| Hard delete cascade (right to erasure) belum ada | Compliance | **TINGGI** — Kewajiban hukum | P1 |
| Breach notification pipeline belum ada | Compliance | **TINGGI** — Kewajiban 72 jam per UU PDP | P1 |
| SyncRegistry harus diupdate manual setiap relasi baru | Operasional | **MEDIUM** — Risiko lupa yang bisa menyebabkan stale data silent | P2 |
| Dapodik API sync deferred (masih manual export) | Integrasi | **MEDIUM** — Bottleneck onboarding sekolah | P2 |
| SMS/WhatsApp usage tracking belum ada | Bisnis | **MEDIUM** — Revenue stream overage tidak bisa dimonetisasi | P2 |

### Catatan Khusus tentang `school_type` dan `coop_type` Immutability

Keputusan bahwa institution type tidak bisa diubah setelah tenant aktif adalah **keputusan yang tepat** dari perspektif data integrity, tetapi menciptakan biaya operasional:
- Tim support harus bisa membuat tenant baru dan migrasi data jika ada perubahan kebutuhan
- Prosedur "recreate tenant" harus terdokumentasi sebelum pelanggan Enterprise pertama onboard

---

## 4. Audit Build vs Buy

| Keputusan | Build (Pilihan SekolahPro) | Buy (Alternatif) | Penilaian |
|-----------|--------------------------|-----------------|----------|
| Backend framework | Go + Chi (custom) | Rails, Laravel, Django | **TEPAT** — Go performa sangat unggul untuk multi-tenant SaaS |
| Payment gateway | Integrasi Midtrans/Xendit/Duitku | Membangun sendiri | **TEPAT** — PCI-DSS compliance sangat mahal dibangun sendiri |
| E-learning | Integrasi Google Classroom/Moodle | Membangun LMS | **TEPAT** — SekolahPro bukan LMS; integrasi jauh lebih realistis |
| Notifikasi WhatsApp | WhatsApp Business API (Meta) | Membangun sendiri | **TEPAT** — Infrastruktur notifikasi bukan differensiator |
| Event bus | NATS JetStream | Kafka, RabbitMQ | **TEPAT** — NATS lebih ringan dan cukup untuk skala ini |
| Monitoring | OpenTelemetry + Prometheus | SaaS monitoring (Datadog) | **DAPAT DIEVALUASI** — Di awal OK, tapi Datadog/New Relic bisa lebih efisien jika tim kecil |
| Auth/JWT | Custom | Keycloak, Auth0 | **DAPAT DIEVALUASI** — Custom JWT lebih ringan, tapi Auth0 mengurangi beban compliance IAM |
| ML credit scoring (Phase 3) | Build (Python) | Buy dari credit bureau | **BELUM DIPUTUSKAN** — Evaluasi saat Phase 2 selesai dan ada data historis |

---

## 5. Strategi Data: Implikasi Vernon di Skala

### Model Konsistensi

SekolahPro menggunakan **model konsistensi hibrida**:

| Domain | Model | Latency Target |
|--------|-------|---------------|
| Auth, billing, transaksi keuangan | Strong (immediate) | < 500ms |
| Profil siswa, kelas, guru | Eventual (Vernon, < 30 detik) | < 100ms |
| Dashboard analytics | Eventual (materialized view, 5–30 menit) | < 1 detik |
| AML monitoring alerts | Near-real-time (rule-based, < 3 hari kerja) | Event-driven |

### Risiko Eventual Consistency yang Perlu Diantisipasi

1. **Rapor yang dicetak dengan data stale** — Jika guru mengubah nama siswa dan langsung cetak rapor sebelum SyncEngine selesai, rapor bisa menampilkan nama lama. **Solusi:** Sebelum generate rapor, force-refresh data dari tabel asli (bypass Vernon) untuk field kritikal.

2. **Saldo koperasi tampil stale di dashboard nasabah** — Data finansial dikategorikan sebagai `critical` (SLA < 1 detik), tetapi Vernon _data tidak pernah menyimpan saldo (by design). **Status:** Aman — saldo selalu dibaca dari tabel asli.

3. **Burst sync saat bulk import** — Import 10.000 siswa sekaligus memicu burst event ke SyncEngine. Coalescing 100ms sudah dirancang untuk ini, tapi perlu load testing sebelum produksi.

### Rekomendasi Monitoring Data Integrity

Tambahkan dashboard khusus untuk:
- `sync_pending_rows` per entity_type (harus mendekati 0 dalam kondisi normal)
- `sync_error_total` per entity_type (alert jika > 0 dalam 5 menit)
- `sync_lag_seconds` P95 (alert jika > 30 detik untuk normal, > 1 detik untuk critical)

---

## 6. Asesmen Arsitektur Keamanan

### Strengths

| Mekanisme | Implementasi | Status |
|-----------|-------------|--------|
| Multi-tenant isolation | `tenant_id` wajib di semua query | Didesain dengan baik |
| Two-phase JWT | Phase 1 (5 menit TTL) + Phase 2 (1 jam TTL) | Kuat — Phase 1 sangat pendek |
| RBAC + scope-based access | Permission per role + class_room_ids di token | Granular dan tepat |
| Immutable transactions | Koreksi via reversal, bukan edit | Sesuai standar audit keuangan |
| Encryption at-rest | AES-256 untuk data sensitif; display masking untuk KTP/NPWP | Sesuai UU PDP |
| Data sovereignty | Data wajib di Indonesia (region Jakarta) | Compliant |

### Gaps yang Perlu Diaddress

| Gap | Rating | Rekomendasi |
|----|--------|------------|
| mTLS antar service (jika microservices keluar) | **MEDIUM** | Diperlukan di Phase 2 jika ada service extraction |
| SIEM (Security Information & Event Management) | **MEDIUM** | Tambahkan di Phase 2; diperlukan untuk audit ISO 27001 |
| Penetration test belum terjadwal | **MEDIUM** | Wajib sebelum Enterprise tier dengan data finansial live |
| Session management (refresh token rotation) | **LOW** | Review protokol refresh token |
| API rate limiting per tenant | **LOW** | Penting untuk mencegah tenant "nakal" membebani sistem |

### Catatan Khusus: Data Anak < 18 Tahun

Hampir **100% pengguna modul sekolah adalah anak di bawah umur**. Ini bukan edge case — ini adalah kasus utama. UU PDP memberikan perlindungan ekstra untuk data anak. Implikasi teknis:
- Setiap field yang menyimpan data anak harus terdokumentasi dalam data catalog
- Hard delete cascade harus diuji secara komprehensif untuk data anak
- Audit log WAJIB memisahkan PII dari metadata (nama anak tidak boleh masuk audit log)

---

## 7. Strategi Ekosistem Integrasi

### Integrasi Kritis (P0 — diperlukan sebelum launch)

| Integrasi | Kesiapan | Risiko |
|-----------|---------|--------|
| Midtrans/Xendit payment gateway | Desain selesai | LOW — API provider sudah stabil |
| WhatsApp Business API (notifikasi) | Desain selesai | LOW |
| Dapodik manual export | Desain selesai | MEDIUM — Format Dapodik sering berubah |

### Integrasi Penting (P1 — diperlukan dalam 6 bulan pertama)

| Integrasi | Kesiapan | Risiko |
|-----------|---------|--------|
| PPATK (laporan AML/CFT) | Desain selesai | MEDIUM — Format laporan PPATK berubah regulasi |
| OJK reporting | Desain selesai | HIGH — Bergantung keputusan OJK tentang threshold |

### Integrasi Deferred (P2 — 2027+)

| Integrasi | Kesiapan | Risiko |
|-----------|---------|--------|
| Dapodik API sync | Deferred — API tidak stabil | HIGH — Kemendikbud belum punya public API yang reliable |
| Dukcapil KYC real-time | Planned 2027 | HIGH — Akses terbatas, biaya per-query |
| BI-FAST | Planned 2027 | MEDIUM — Regulasi koperasi belum jelas |
| Google Classroom / Moodle | Desain selesai | LOW |

### Keputusan Strategis: SekolahPro Bukan LMS

Ini adalah keputusan **tepat** — membangun LMS bersaing langsung dengan Google Classroom, Moodle, dan Canvas yang sudah gratis dan kuat. Posisi SekolahPro sebagai "integration hub" yang menarik data dari LMS eksternal adalah differensiator, bukan kelemahan.

---

## 8. Rekomendasi Roadmap Teknologi (12–24 Bulan)

### Segera (Q1–Q2 2026) — P0 Items

```
1. Implementasi feature gating system (ADR baru)
2. Implementasi usage metering (hitung siswa aktif per tenant)
3. Implementasi billing & invoice system (ADR baru)
4. Implementasi consent flow UU PDP (onboarding siswa)
5. Load testing SyncEngine dengan dataset 100.000+ siswa
6. Penetration test sebelum Enterprise tier live
```

### Jangka Menengah (Q3–Q4 2026) — P1 Items

```
7. Implementasi data retention scheduler (auto-anonymize setelah 5 tahun)
8. Implementasi hard delete cascade + data portability endpoint
9. Implementasi breach notification pipeline (72 jam)
10. Laporan regulasi Dapodik, OJK, PPATK format validation
11. PgBouncer untuk connection pooling (jika > 100 sekolah aktif)
```

### Jangka Menengah-Panjang (2027) — Automation & Scale

```
12. PostgreSQL read replica + Redis cluster
13. CI/CD pipeline otomatis (GitHub Actions)
14. OpenTelemetry + Grafana monitoring production
15. Dukcapil KYC integration (jika akses tersedia)
16. BI-FAST payment (jika regulasi koperasi jelas)
17. Unit test coverage ≥ 80% untuk core business logic
```

### Jangka Panjang (2028+) — Intelligence

```
18. ML credit scoring (butuh 2 tahun data historis)
19. Anomaly detection AML (ML-based, Phase 3)
20. Open banking integration (jika regulatory framework tersedia)
21. Data warehouse untuk analytics lintas tenant (BigQuery/Redshift)
```

---

## 9. Kesimpulan: Kekuatan dan Risiko Teknis

### Kekuatan Arsitektur

1. **Go + Clean Architecture** memberikan foundation yang kuat untuk pertumbuhan tim — engineer baru bisa onboard dengan cepat karena boundary layer yang jelas
2. **CQRS hybrid** memungkinkan optimasi independen antara write dan read path
3. **Vernon Pattern** memberikan performa exceptional untuk domain read-heavy (siswa, transaksi) tanpa kompromis
4. **Dual-mode architecture** menggunakan strategy pattern yang bersih — tidak ada `if islamic mode` yang tersebar di seluruh codebase
5. **Event-driven foundation** mempersiapkan platform untuk ML pipeline di Phase 3 tanpa refactoring besar

### Risiko Teknis yang Harus Dimonitor

1. **[HIGH] SyncRegistry manual** — Setiap relasi baru memerlukan developer ingat mendaftarkan ke registry. Satu lupa bisa menyebabkan stale data yang tidak terdeteksi.
2. **[HIGH] Feature gating belum ada** — Platform tidak bisa dimonetisasi tanpa ini.
3. **[MEDIUM] Vernon eventual consistency** — Perlu UI/UX yang jelas untuk user saat data sedang disinkronisasi.
4. **[MEDIUM] Skala PostgreSQL single primary** — Sudah cukup hingga ~1.000 sekolah; perlu read replica setelah itu.
5. **[MEDIUM] Test coverage** — Kompleksitas dual-mode memerlukan test matrix 4 kombinasi untuk setiap fitur yang berbeda antar mode.
