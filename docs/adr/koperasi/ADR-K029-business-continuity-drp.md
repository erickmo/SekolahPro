# ADR-K029: Business Continuity & Disaster Recovery (BCP/DRP)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Koperasi yang mengelola uang nasabah wajib memiliki rencana Business Continuity dan Disaster Recovery. Regulasi OJK (POJK 13/POJK.03/2016 tentang Pengelolaan Risiko) mewajibkan LKM memiliki BCP/DRP yang terdokumentasi dan diuji secara berkala.

Risiko yang harus dimitigasi:
- **Hardware failure**: server mati, disk corrupt, network down
- **Software failure**: bug kritis, data corruption, deployment gagal
- **Human error**: accidental deletion, wrong configuration
- **Security breach**: ransomware, hack, insider threat
- **Force majeure**: bencana alam, pandemi, kebakaran, listrik padam
- **Third-party failure**: cloud provider down, payment gateway outage

Tanpa BCP/DRP:
- Koperasi tidak bisa beroperasi → nasabah tidak bisa akses dana
- Data loss → saldo nasabah hilang, tidak bisa rekonsiliasi
- Regulatory violation → sanksi OJK
- Reputasi hancur → nasabah kabur

## Decision

### 1. Recovery Objectives

```
recovery_objectives:
├── RPO (Recovery Point Objective)
│   ├── Critical data (saldo, transaksi): 0 (zero data loss)
│   ├── Operational data (session, log): ≤ 1 jam
│   └── Reporting data (laporan): ≤ 24 jam
│
├── RTO (Recovery Time Objective)
│   ├── Transaction processing: ≤ 15 menit (failover)
│   ├── Online banking/portal: ≤ 30 menit
│   ├── Full system recovery: ≤ 4 jam
│   └── Reporting/analytics: ≤ 24 jam
│
└── Availability SLA
    ├── Business hours (08:00-17:00 WIB): 99.9%
    ├── After hours: 99.5%
    └── Planned maintenance: Sunday 02:00-06:00 WIB
```

### 2. Backup Strategy

```
backup_strategy:
├── DATABASE (PostgreSQL)
│   ├── Continuous WAL archiving → zero data loss untuk transaksi
│   ├── Full backup: daily at 02:00 WIB
│   ├── Incremental backup: setiap 4 jam
│   ├── Retensi:
│   │   ├── Daily backup: 30 hari
│   │   ├── Weekly backup: 12 minggu
│   │   ├── Monthly backup: 12 bulan
│   │   └── Yearly backup: 5 tahun
│   └── Storage: primary + off-site (different region/availability zone)
│
├── APPLICATION FILES (dokumen, upload)
│   ├── Real-time sync ke secondary storage
│   ├── Versioning enabled (keep last 10 versions)
│   └── Retensi: sama dengan database
│
├── CONFIGURATION (env, secrets, settings)
│   ├── Version controlled (git)
│   ├── Encrypted secrets vault (HashiCorp Vault / AWS Secrets Manager)
│   └── Infrastructure as Code (Terraform/Pulumi)
│
└── VERNON READ CACHE (Redis/Elasticsearch)
    ├── Tidak perlu backup (rebuildable dari PostgreSQL)
    ├── Rebuild time: ≤ 30 menit untuk full dataset
    └── Priority rebuild: tenant aktif dulu
```

### 3. Architecture for High Availability

```
┌─────────────────────────────────────────────────────────┐
│                    LOAD BALANCER                         │
│                  (Active-Active)                         │
└──────────┬──────────────────────┬───────────────────────┘
           │                      │
    ┌──────▼──────┐        ┌──────▼──────┐
    │  APP NODE 1 │        │  APP NODE 2 │    ← Min 2 nodes
    │  (Primary)  │        │ (Secondary) │       active-active
    └──────┬──────┘        └──────┬──────┘
           │                      │
    ┌──────▼──────────────────────▼──────┐
    │         CONNECTION POOLER          │
    │         (PgBouncer / Odyssey)      │
    └──────┬──────────────────────┬──────┘
           │                      │
    ┌──────▼──────┐        ┌──────▼──────┐
    │   DB PRIMARY│◄──────►│ DB REPLICA  │  ← Streaming replication
    │   (Read-Write)│ sync  │ (Read-Only)│     synchronous mode
    └─────────────┘        └─────────────┘

    ┌──────────────┐        ┌──────────────┐
    │ REDIS PRIMARY│◄──────►│REDIS REPLICA │  ← For Vernon cache
    └──────────────┘  sync  └──────────────┘
```

**Failover sequence:**
1. Health check mendeteksi primary node failure (dalam 5 detik)
2. Automatic DNS/load balancer switch ke secondary (dalam 10 detik)
3. Database replica promoted ke primary (dalam 30 detik)
4. Vernon cache rebuild dimulai (background, target 30 menit)
5. Alert ke on-call team (immediate via WhatsApp/PagerDuty)

### 4. Disaster Recovery Tiers

```
dr_tier:
├── TIER 1 — ACTIVE-ACTIVE (Recommended for > 1000 nasabah)
│   ├── Two data centers / availability zones
│   ├── Zero-downtime failover
│   ├── RPO: 0, RTO: 15 menit
│   └── Biaya: tinggi
│
├── TIER 2 — WARM STANDBY (Recommended for 100-1000 nasabah)
│   ├── Primary + standby in different zone
│   ├── Standby database replicated, app scaled down
│   ├── RPO: < 1 menit, RTO: 30 menit
│   └── Biaya: menengah
│
└── TIER 3 — COLD STANDBY (Minimum untuk < 100 nasabah)
    ├── Daily backup ke off-site storage
    ├── Manual recovery procedure
    ├── RPO: 24 jam, RTO: 4 jam
    └── Biaya: rendah
```

**Rekomendasi per ukuran koperasi:**
- Koperasi kecil (< 100 anggota, 1 sekolah): Tier 3 minimum
- Koperasi medium (100-1000 anggota): Tier 2
- Koperasi besar (> 1000 anggota, multi-cabang): Tier 1

### 5. Incident Response

```
incident_response:
├── SEVERITY LEVELS
│   ├── P1 — CRITICAL: System down, data loss, security breach
│   │   ├── Response: immediate (< 15 menit)
│   │   ├── Escalation: on-call → manager → ketua
│   │   └── Communication: WhatsApp group + phone call
│   │
│   ├── P2 — HIGH: Major feature down, degraded performance
│   │   ├── Response: < 30 menit
│   │   ├── Escalation: on-call → manager
│   │   └── Communication: WhatsApp group
│   │
│   ├── P3 — MEDIUM: Minor feature impaired, workaround available
│   │   ├── Response: < 2 jam
│   │   ├── Escalation: on-call
│   │   └── Communication: ticket system
│   │
│   └── P4 — LOW: Cosmetic, enhancement, non-urgent
│       ├── Response: next business day
│       ├── Escalation: none
│       └── Communication: ticket system
│
└── ESCALATION MATRIX
    ├── Level 1: On-call engineer (7x24 untuk P1)
    ├── Level 2: Engineering manager
    ├── Level 3: CTO / Ketua Koperasi
    └── External: Vendor support (cloud provider, database vendor)
```

```
incident_record:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Incident Info ──
├── incident_number       VARCHAR         ← "INC-2026-00001"
├── severity              ENUM (P1, P2, P3, P4)
├── incident_type         ENUM (system_down, data_corruption, security_breach,
│                                performance_degradation, data_loss,
│                                third_party_outage, human_error)
├── title                 VARCHAR
├── description           TEXT
│
├── ── Timeline ──
├── detected_at           TIMESTAMPTZ
├── acknowledged_at       TIMESTAMPTZ (nullable)
├── mitigated_at          TIMESTAMPTZ (nullable)
├── resolved_at           TIMESTAMPTZ (nullable)
├── post_mortem_at        TIMESTAMPTZ (nullable)
│
├── ── Impact ──
├── affected_services     VARCHAR[]
├── affected_tenants      UUID[]
├── affected_nasabah      INT
├── data_loss             BOOLEAN DEFAULT false
├── data_loss_description TEXT (nullable)
│
├── ── Response ──
├── responder_ids         UUID[] (FK → user)
├── actions_taken         JSONB           ← [{time, action, by}]
├── root_cause            TEXT (nullable)
├── remediation           TEXT (nullable)
├── post_mortem_doc_id    UUID (nullable, FK → document)
│
├── ── Status ──
├── status                ENUM (detected, acknowledged, investigating,
│                                mitigated, resolved, closed)
│
└── created_at            TIMESTAMPTZ
```

### 6. Offline Mode — Business Continuity Procedures

Ketika sistem benar-benar down, operasional koperasi harus tetap berjalan:

```
offline_procedures:
├── TRANSAKSI MANUAL
│   ├── Setoran: catat di form manual (triplicate)
│   ├── Penarikan: Hanya yang ≤ 2 juta, catat manual
│   ├── Angsuran: Terima pembayaran, berikan kwitansi manual
│   └── Semua transaksi manual → di-input ulang saat sistem recovery
│
├── RECONCILIATION POST-RECOVERY
│   ├── Verifikasi setiap transaksi manual terhadip input digital
│   ├── Matching: manual form number ↔ digital transaction ID
│   ├── Discrepancy resolution sebelum re-open ke nasabah
│   └── Sign-off oleh supervisor setelah reconciliation selesai
│
├── AUTHORIZATION DURING OUTAGE
│   ├── Pre-printed authorization forms untuk approval manual
│   ├── Dual-signature untuk transaksi > threshold
│   └── Semua manual approvals → di-enter ke sistem saat recovery
│
└── COMMUNICATION
    ├── Pengumuman ke nasabah via WhatsApp (template pre-defined)
    ├── Estimasi recovery time
    ├── Layanan alternatif yang tersedia
    └── Update setiap 30 menit sampai resolved
```

### 7. Testing & Maintenance Schedule

| Test Type | Frequency | Description |
|---|---|---|
| Backup restore test | Bulanan | Restore database ke test environment, verifikasi integritas |
| Failover test | Triwulanan | Simulasi primary failure, verifikasi auto-failover |
| Full DR test | Tahunan | Simulasi disaster total, recovery dari backup |
| Incident response drill | Semester | Simulasi P1 incident, test komunikasi & eskalasi |
| Penetration test | Tahunan | Security testing oleh pihak ketiga |
| BCP review | Tahunan | Update BCP berdasarkan perubahan infrastruktur |

### 8. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Priority data | Transaksi, saldo | + zakat/infaq records |
| DPS notification | Tidak perlu | DPS di-notify jika ada P1/P2 |
| Audit trail recovery | Standard | DPS harus verifikasi post-recovery |
| Social fund integrity | Tidak ada | Dana sosial harus verified post-recovery |

### 9. Vernon _rels dan _data Structure

Tidak memerlukan Vernon read cache — ini adalah meta-data operasional, bukan domain data yang di-cache.

Incident records disimpan langsung di PostgreSQL dengan standard CRUD.

### 10. Authorization — RBAC

| Permission | On-Call | Manager | Admin | Ketua |
|---|---|---|---|---|
| View incident dashboard | v | v | v | v |
| Create incident | v | v | v | v |
| Acknowledge incident (P3-P4) | v | v | v | v |
| Acknowledge incident (P1-P2) | - | v | v | v |
| Update incident status | v | v | v | v |
| Escalate incident | v | v | v | v |
| Trigger failover | - | - | v | v |
| Approve manual transactions | - | v | v | - |
| Post-recovery sign-off | - | v | v | - |
| Configure backup schedule | - | - | v | - |
| Run DR test | - | - | v | - |
| Approve BCP changes | - | - | v | v |

### 11. Consequences

**Keuntungan:**
- Kepatuhan terhadap POJK risk management requirements
- Minimal downtime → nasabah tetap bisa bertransaksi
- Data loss prevention → saldo nasabah aman
- Clear incident response → tidak panik saat terjadi masalah

**Risiko:**
- Cost infrastruktur HA (minimal 2 nodes)
- Complexity operasional untuk failover
- DR testing membutuhkan waktu & resource

**Mitigasi:**
- Gunakan managed cloud services (RDS, etc.) untuk reduce ops burden
- Automated failover mengurangi human error
- DR test bisa dijadwalkan di weekend/maintenance window
- Tier selection berdasarkan ukuran koperasi (cost proportionate)
