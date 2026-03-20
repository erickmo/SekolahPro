# 12 — Platform & Infrastruktur (M45–M47)

> **Terakhir diperbarui**: 20 Maret 2026

Tiga modul ini aktif di **semua tier termasuk FREE** dan tidak bisa dimatikan. Mereka adalah fondasi keamanan dan kepatuhan seluruh platform EDS.

---

## M45 — Audit Log & Compliance Center
**Tier**: ALL (tidak bisa dimatikan)

### Prinsip
- **Immutable** — log tidak bisa diedit atau dihapus oleh siapapun, termasuk EDS superadmin
- **Complete** — semua perubahan data signifikan dicatat
- **Attributable** — setiap entri punya userId, role, IP, timestamp
- **Retained** — minimum 5 tahun (sesuai Peraturan Arsip Nasional)

### Schema

```prisma
model AuditLog {
  id            String      @id @default(cuid())
  timestamp     DateTime    @default(now())
  schoolId      String?     // null untuk aksi di Developer Portal
  userId        String
  userRole      String
  module        String      // 'M04', 'M18', 'DEV_PORTAL', dll
  action        AuditAction
  entityType    String      // 'Student', 'Grade', 'Payment', dll
  entityId      String
  oldValue      Json?       // snapshot sebelum perubahan
  newValue      Json?       // snapshot sesudah perubahan
  ipAddress     String
  userAgent     String
  sessionId     String?
  impersonatedBy String?    // userId EDS support jika via impersonation
  isSensitive   Boolean     @default(false) // rekam medis, BK, ABK
  requestId     String?     // correlation ID untuk tracing

  @@index([schoolId, timestamp])
  @@index([userId, timestamp])
  @@index([entityType, entityId])
}

enum AuditAction {
  CREATE UPDATE DELETE VIEW EXPORT
  LOGIN LOGOUT LOGIN_FAILED
  IMPERSONATE_START IMPERSONATE_END
  FEATURE_TOGGLE TENANT_SUSPEND
  BULK_IMPORT BULK_EXPORT
  PASSWORD_CHANGE TWO_FA_ENABLED TWO_FA_DISABLED
}
```

### Middleware — Auto-inject Audit

```typescript
// packages/api/middleware/audit.ts
export function auditMiddleware(module: string, entityType: string) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const original = res.json.bind(res);
    res.json = (body: any) => {
      if (res.statusCode < 400 && req.method !== 'GET') {
        // Async — tidak block response
        logAudit({
          schoolId: req.user.schoolId,
          userId: req.user.id,
          userRole: req.user.role,
          module,
          action: httpMethodToAction(req.method),
          entityType,
          entityId: body?.data?.id ?? req.params.id,
          newValue: sanitize(body?.data),
          ipAddress: req.ip,
          userAgent: req.headers['user-agent'] ?? '',
          impersonatedBy: req.headers['x-impersonated-by'] as string,
        }).catch(console.error);
      }
      return original(body);
    };
    next();
  };
}
```

### Anomaly Detection

```typescript
// Cron setiap 15 menit — cek pola akses anomali
const ANOMALY_RULES = [
  { name: 'ODD_HOURS_LOGIN',    check: (log) => isOddHour(log.timestamp) && log.action === 'LOGIN' },
  { name: 'BULK_EXPORT_LARGE',  check: (log) => log.action === 'BULK_EXPORT' },
  { name: 'RAPID_DELETE',       check: (log) => log.action === 'DELETE', rateLimit: { count: 10, windowMin: 5 } },
  { name: 'SENSITIVE_MASS_VIEW', check: (log) => log.isSensitive && log.action === 'VIEW',
    rateLimit: { count: 50, windowMin: 60 } },
];
```

### API
```
GET  /api/v1/audit?module=&userId=&from=&to=  — query audit log
GET  /api/v1/audit/export                      — export ke Excel (skill `xlsx`)
GET  /api/v1/audit/anomalies                   — daftar anomali terdeteksi
```

---

## M46 — Backup, Recovery & Business Continuity
**Tier**: ALL (tidak bisa dimatikan)

### Strategi Backup

```
Level 1: WAL Archiving (PostgreSQL 17)
  → Continuous, ke S3 offsite, RPO: ~1 menit

Level 2: Logical Backup (pg_dump per tenant)
  → Setiap hari jam 02.00 WIB, retained 30 hari

Level 3: Full Snapshot (disk-level)
  → Setiap minggu, retained 12 minggu

Level 4: Annual Archive
  → Setiap 1 Januari, retained 7 tahun (compliance)
```

### Konfigurasi PostgreSQL 17 untuk HA

```sql
-- postgresql.conf (PostgreSQL 17)
wal_level = replica
max_wal_senders = 5
wal_keep_size = 2GB   -- PostgreSQL 17: wal_keep_size (bukan wal_keep_segments)
archive_mode = on
archive_command = 'aws s3 cp %p s3://eds-wal-archive/%f'

-- Fitur baru PostgreSQL 17: logical replication slot failover
-- Memungkinkan replica langsung menjadi primary tanpa kehilangan slot
```

### Point-in-Time Recovery

```bash
# Restore ke waktu tertentu (contoh: sebelum insiden)
pg_restore_pitr \
  --target-time "2026-03-20 10:30:00+07" \
  --school-id "ckxyz123" \
  --output-db "eds_restore_temp"

# Verifikasi data, lalu promote jika OK
```

### SLA

| Metrik | Target |
|--------|--------|
| RTO (Recovery Time Objective) | < 4 jam |
| RPO (Recovery Point Objective) | < 1 jam |
| Uptime | 99.5% (BASIC) / 99.9% (ENTERPRISE) |
| Backup verification drill | Tiap semester |

### Status Page

```
status.eds.id — publik, realtime
├── API Gateway
├── Web Portal (per region)
├── Database Cluster
├── AI Engine (Anthropic API)
├── WhatsApp Gateway
├── Payment Gateway
└── MinIO Storage
```

### Docker Compose untuk PostgreSQL 17

```yaml
# infrastructure/docker-compose.yml
services:
  postgres-primary:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: eds_db
      POSTGRES_USER: eds
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    command: >
      postgres
        -c wal_level=replica
        -c max_wal_senders=5
        -c wal_keep_size=2GB
        -c archive_mode=on
        -c archive_command='cp %p /var/lib/postgresql/wal_archive/%f'
    volumes:
      - pgdata:/var/lib/postgresql/data
      - wal_archive:/var/lib/postgresql/wal_archive

  postgres-replica:
    image: postgres:17-alpine
    environment:
      PGUSER: eds
      PGPASSWORD: ${DB_PASSWORD}
    command: >
      bash -c "
        pg_basebackup -h postgres-primary -D /var/lib/postgresql/data -P -U eds -Xs -R
        postgres
      "
    depends_on:
      - postgres-primary
    volumes:
      - pgdata-replica:/var/lib/postgresql/data

volumes:
  pgdata:
  pgdata-replica:
  wal_archive:
```

---

## M47 — Security Center
**Tier**: ALL (tidak bisa dimatikan)

### Checklist Keamanan per Release

```
[ ] Dependency audit: pnpm audit --audit-level=high
[ ] SAST scan: ESLint security rules, Semgrep
[ ] Secret scan: truffleHog / gitleaks di CI
[ ] OWASP Top 10 manual check (tiap sprint)
[ ] Penetration test: tiap kuartal (eksternal tiap tahun)
```

### 2FA / MFA

```typescript
// Wajib untuk role: ADMIN_SEKOLAH, KEPALA_SEKOLAH, BENDAHARA, EDS_*
export const MFA_REQUIRED_ROLES = [
  'ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'BENDAHARA',
  'KASIR_KOPERASI', 'SECURITY_OFFICER',
  'EDS_SUPERADMIN', 'EDS_SUPPORT', 'EDS_SALES',
];

// Method yang didukung
type MFAMethod = 'TOTP' | 'SMS_OTP' | 'EMAIL_OTP';
// TOTP (Google Authenticator) adalah default; SMS/email sebagai fallback
```

### Rate Limiting per Endpoint

```typescript
export const RATE_LIMITS = {
  '/api/auth/login':           { points: 5,    duration: 60 },   // 5x/menit
  '/api/auth/forgot-password': { points: 3,    duration: 3600 }, // 3x/jam
  '/api/v1/exam/generate':     { points: 10,   duration: 60 },   // 10 req/menit
  '/api/v1/notifications/send':{ points: 100,  duration: 60 },
  '/dev/tenants/:id/impersonate': { points: 3, duration: 3600 }, // impersonation limit
};
```

### Session Management

```typescript
export const SESSION_CONFIG = {
  accessTokenTTL:   '15m',
  refreshTokenTTL:  '7d',
  idleTimeout:      '2h',      // auto-logout jika idle
  maxActiveSessions: 3,        // maks 3 device per user
  // Saat password diubah, semua session lain otomatis di-revoke
  revokeOnPasswordChange: true,
};
```

### Enkripsi Data Sensitif

```typescript
// Three-tier encryption untuk data paling sensitif
// Tier 1: Database encryption at-rest (PostgreSQL TDE atau disk encryption)
// Tier 2: Column-level AES-256-GCM (untuk kesehatan, BK, ABK)
// Tier 3: Envelope encryption dengan per-schoolId data key

export async function encryptSensitiveField(
  data: string,
  schoolId: string,
  module: 'HEALTH' | 'COUNSELING' | 'ABK'
): Promise<Buffer> {
  const dataKey = await getOrCreateDataKey(schoolId, module);
  const iv = crypto.randomBytes(12);
  const cipher = crypto.createCipheriv('aes-256-gcm', dataKey, iv);
  const encrypted = Buffer.concat([cipher.update(data, 'utf8'), cipher.final()]);
  const authTag = cipher.getAuthTag();
  // Format: iv (12B) + authTag (16B) + ciphertext
  return Buffer.concat([iv, authTag, encrypted]);
}
```

### Incident Response Playbook

```
Level 1 — Data breach suspected
  1. Alert ke tim security (PagerDuty)
  2. Isolasi tenant yang terdampak (suspend API access)
  3. Snapshot forensik database & logs
  4. Notifikasi ke sekolah terdampak dalam 72 jam (UU PDP pasal 46)
  5. Root cause analysis
  6. Patch & re-enable setelah clear

Level 2 — DDoS / availability
  1. Aktifkan Cloudflare Under Attack Mode
  2. Scale up Kubernetes replicas
  3. Block IP ranges via WAF
  4. Update status.eds.id

Level 3 — Credential compromise
  1. Force logout semua session
  2. Require password reset + 2FA re-enroll
  3. Audit log review untuk akses janggal
  4. Notify user via email
```
