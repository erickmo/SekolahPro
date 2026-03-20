# 15 — Developer Guide

---

## Quick Start

```bash
git clone https://github.com/eds-platform/eds.git && cd eds
pnpm install

# Setup env
cp .env.example .env
# Edit .env dengan credentials lokal

# Setup database
cd packages/database
pnpm prisma migrate dev
pnpm prisma db seed   # seed: 2 tenant demo, user admin, modul config

# Start semua services
docker-compose up -d  # postgres:17, redis:7, minio, clickhouse, jitsi, mosquitto
pnpm dev              # semua apps via turbo (web, developer-portal, api, mobile)
```

**Default dev accounts setelah seed**:
- EDS Superadmin: `admin@eds.id` / `admin123`
- Demo sekolah: `sman1demo.eds.localhost:3000` → `kepsek@sman1.sch.id` / `demo123`
- Demo vendor: `dev.eds.localhost:3001` (Developer Portal)

---

## Struktur Apps

```
apps/
├── web/                    # School portals (port 3000)
├── developer-portal/       # SaaS ops + listing (port 3001)
├── mobile/                 # React Native (Expo)
├── ppdb/                   # PPDB standalone (port 3002)
└── foundation/             # Dashboard yayasan (port 3003)

packages/
├── api/                    # Fastify API server (port 4000)
├── database/               # Prisma + migrations
├── warehouse/              # ETL Airflow + ClickHouse schema
├── ai/                     # Claude AI wrappers + prompt library
├── iot/                    # MQTT client + Node-RED flows
├── notifications/          # WA, push, email, SMS
├── billing/                # Stripe + Midtrans subscription
├── listing/                # Listing marketplace engine
├── ui/                     # shadcn/ui + custom components
└── config/                 # ESLint, TypeScript, Tailwind shared config
```

---

## Konvensi Kode

- **Files**: `kebab-case.ts`
- **Components**: `PascalCase`
- **Functions/Variables**: `camelCase`
- **Database fields**: `camelCase` (Prisma)
- **API endpoints**: `/api/v1/[domain]/[resource]`
- **Env vars**: `UPPER_SNAKE_CASE`

### API Response Format

```typescript
// Sukses
{ success: true, data: T, meta?: { page, limit, total } }

// Error
{ success: false, error: { code: string, message: string, details?: unknown } }
```

### Error Codes

```
AUTH_001  Token tidak valid           M05_001  Soal gagal digenerate AI
AUTH_002  Tidak punya akses           M05_002  Quota AI habis bulan ini
AUTH_003  Tier tidak mendukung fitur  M14_001  Gagal kirim WhatsApp
TENANT_001 Sekolah tidak ditemukan   M14_002  Quota WA habis
TENANT_002 Tenant suspended          M18_001  Transaksi SPP gagal
M01_001  NISN sudah terdaftar        M40_001  Sinkronisasi Dapodik gagal
M03_001  Saldo tidak mencukupi       LIST_001 Vendor belum diapprove
```

---

## Environment Variables

```env
# ─── Database ───────────────────────────────────
DATABASE_URL="postgresql://eds:eds@localhost:5432/eds_dev"
REDIS_URL="redis://localhost:6379"
CLICKHOUSE_URL="http://localhost:8123"

# ─── Auth ───────────────────────────────────────
NEXTAUTH_SECRET="dev-secret-change-in-prod"
NEXTAUTH_URL="http://localhost:3000"
JWT_SECRET="dev-jwt-secret"
JWT_EXPIRES_IN="15m"
REFRESH_TOKEN_EXPIRES_IN="7d"

# ─── AI ─────────────────────────────────────────
ANTHROPIC_API_KEY="sk-ant-..."

# ─── File Storage ───────────────────────────────
MINIO_ENDPOINT="localhost"
MINIO_PORT="9000"
MINIO_ACCESS_KEY="minioadmin"
MINIO_SECRET_KEY="minioadmin"
MINIO_BUCKET="eds-dev"

# ─── Payment ────────────────────────────────────
MIDTRANS_SERVER_KEY="SB-Mid-server-..."
MIDTRANS_CLIENT_KEY="SB-Mid-client-..."
XENDIT_SECRET_KEY="xnd_development_..."
STRIPE_SECRET_KEY="sk_test_..."         # SaaS subscription billing
STRIPE_WEBHOOK_SECRET="whsec_..."

# ─── Notifications ──────────────────────────────
WHATSAPP_API_URL="https://waba.360dialog.io/v1"
WHATSAPP_API_KEY="..."
WHATSAPP_PHONE_NUMBER_ID="..."
SMTP_HOST="localhost"
SMTP_PORT="1025"                        # Mailhog di dev
SMTP_USER=""
SMTP_PASS=""
FCM_SERVER_KEY="..."

# ─── Video Conference ────────────────────────────
JITSI_APP_ID="eds-jitsi"
JITSI_APP_SECRET="..."
JITSI_URL="https://meet.eds.id"

# ─── IoT ────────────────────────────────────────
MQTT_BROKER_URL="mqtt://localhost:1883"
MQTT_USERNAME="eds"
MQTT_PASSWORD="..."

# ─── Government APIs ────────────────────────────
DAPODIK_API_URL="https://api.dapodik.kemdikbud.go.id"
DAPODIK_API_KEY="..."
ARKAS_API_URL="..."

# ─── SaaS ───────────────────────────────────────
TENANT_BASE_DOMAIN="eds.localhost"      # eds.id di production
DEVELOPER_PORTAL_SECRET="..."           # JWT secret khusus Developer Portal

# ─── Vernon HRM ─────────────────────────────────
VERNON_HRM_API_URL="http://localhost:5100"
VERNON_HRM_API_KEY="..."
VERNON_HRM_TENANT_PREFIX="eds_"      # prefix schoolId sebagai tenant HRM

# ─── Vernon Accounting ───────────────────────────
VERNON_ACC_API_URL="http://localhost:5200"
VERNON_ACC_API_KEY="..."
VERNON_ACC_TENANT_PREFIX="eds_"
VERNON_ACC_DEFAULT_COA="K13_DEFAULT"  # chart of accounts default sekolah

# ─── Monitoring ─────────────────────────────────
SENTRY_DSN="..."
POSTHOG_API_KEY="..."                   # analytics (opsional)
```

---

## Testing Strategy

```
Unit Tests      → Jest — business logic, AI prompts, utils
Integration     → Jest + Supertest — API endpoints per modul
E2E             → Playwright — critical flows (login, PPDB, ujian, SPP)
AI Tests        → Mock Anthropic SDK — CI/CD tanpa cost AI
Load Tests      → k6 — multi-tenant performance (target: 1000 req/s)
Security Tests  → OWASP ZAP — otomatis di staging sebelum deploy prod
```

### Contoh Test Feature Gate

```typescript
// packages/api/modules/exam/__tests__/quota.test.ts
describe('AI Quota Gate', () => {
  it('blocks request when quota exceeded', async () => {
    await setAIUsage(schoolId, 'M05', 500); // habiskan quota
    const res = await request(app)
      .post('/api/v1/exam/generate')
      .set('Authorization', `Bearer ${token}`)
      .send(examConfig);
    expect(res.status).toBe(429);
    expect(res.body.error.code).toBe('M05_002');
  });
});
```

---

## Deployment

```bash
# Staging (otomatis via GitHub Actions saat push ke main)
git push origin main

# Production (manual approval di GitHub Actions)
# 1. Buat release tag
git tag v1.2.0 && git push origin v1.2.0
# 2. Approve deployment di GitHub Actions
# 3. Blue-green swap otomatis (zero downtime)

# Rollback darurat
kubectl rollout undo deployment/eds-api
kubectl rollout undo deployment/eds-web
```

### Deployment Rules

1. **Jangan deploy saat jam 07.00–16.00** (jam sekolah aktif)
2. **Wajib migration review** sebelum deploy yang mengubah schema
3. **Smoke test otomatis** setelah deploy (cek 10 endpoint kritis)
4. **Rollback otomatis** jika error rate > 5% dalam 5 menit pertama

---

## Skills yang Tersedia

Sebelum membangun fitur yang menghasilkan file output, baca SKILL.md yang relevan:

```
PDF (rapor, kartu siswa, soal) → /mnt/skills/public/pdf/SKILL.md
Word (RPP, surat, laporan)     → /mnt/skills/public/docx/SKILL.md
Excel (import/export data)     → /mnt/skills/public/xlsx/SKILL.md
UI/Dashboard                   → /mnt/skills/public/frontend-design/SKILL.md
MCP/API (integrasi Dapodik)    → /mnt/skills/examples/mcp-builder/SKILL.md
```
