# ADR-K032: Biometric Authentication

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Untuk transaksi bernilai tinggi dan operasi sensitif, PIN/password saja tidak cukup aman. Koperasi yang mengelola uang nasabah perlu mekanisme autentikasi yang lebih kuat:

- Penarikan besar (> 25 juta) perlu identifikasi fisik yang kuat
- Approval pinjaman membutuhkan non-repudiation
- Teller perlu verifikasi identitas nasabah yang lebih reliable dari KTP visual
- Mengurangi fraud dari PIN yang bocor atau di-share

**Pertimbangan khusus:**
- Biometric data termasuk **data pribadi sensitif** (ref K028 UU PDP)
- Koperasi sekolah beroperasi di lingkungan yang mungkin infrastruktur terbatas
- Biometric harus accessible — tidak boleh terlalu mahal atau memerlukan hardware khusus
- Fallback mechanism wajib jika biometric gagal

## Decision

### 1. Biometric Types Supported

```
biometric_type:
├── FINGERPRINT         ← Primary — paling umum, murah, reliable
│   ├── Hardware: USB fingerprint scanner (teller), smartphone (mobile)
│   ├── Standard: ISO 19794-2
│   └── FAR/FRR: 0.001% / 1% (acceptable untuk financial)
│
├── FACE_RECOGNITION    ← Secondary — untuk mobile/self-service
│   ├── Hardware: Camera (standard), smartphone
│   ├── Liveness detection WAJIB (prevent spoofing)
│   └── FAR/FRR: 0.002% / 2%
│
└── VOICE_RECOGNITION   ← Tertiary (opsional) — untuk telepon/IVR
    ├── Hardware: Microphone / phone
    └── FAR/FRR: 0.5% / 5% (less reliable, not primary)
```

### 2. Architecture — On-Device Biometric Preference

```
Biometric Storage Strategy:
┌─────────────────────────────────────────────────────────┐
│ PREFERRED: On-Device (Secure Enclave / Keystore)        │
│                                                         │
│ ┌─────────────┐     ┌──────────────────┐               │
│ │ Biometric    │     │ Server stores    │               │
│ │ Template     │     │ only:            │               │
│ │ stored       │     │ - Hashed token   │               │
│ │ ON DEVICE    │     │ - Public key     │               │
│ │ (Secure      │     │ - NOT the actual │               │
│ │  Enclave/    │     │   biometric data │               │
│ │  Keystore)   │     │                  │               │
│ └─────────────┘     └──────────────────┘               │
│                                                         │
│ Advantage: Server breach → no biometric data leaked     │
│ Requirement: UU PDP compliance (data tidak keluar device)│
└─────────────────────────────────────────────────────────┘

FALLBACK: Server-Side Template (only when on-device not possible)
┌─────────────────────────────────────────────────────────┐
│ For: USB fingerprint scanner at teller station          │
│ Storage: Encrypted template in database                 │
│ Protection: AES-256 encryption, separate key management │
│ Access: Only biometric verification service can access   │
└─────────────────────────────────────────────────────────┘
```

### 3. Data Model

```
biometric_enrollment:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Enrollment Info ──
├── biometric_type        ENUM (fingerprint, face, voice)
├── device_type           ENUM (teller_scanner, smartphone, web_camera, ivr)
├── device_id             VARCHAR (nullable, identifier perangkat)
│
├── ── Template ──
├── storage_type          ENUM (on_device, server_encrypted)
├── template_hash         VARCHAR         ← Hash dari template (bukan template itu sendiri)
├── template_encrypted    BYTEA (nullable, hanya jika server-side)
├── encryption_key_ref    VARCHAR (nullable, reference ke key management)
│
├── ── Quality ──
├── quality_score         INT             ← 0-100, minimum 70 untuk accepted
├── enrollment_attempts   INT DEFAULT 0
├── liveness_verified     BOOLEAN DEFAULT false   ← WAJIB untuk face recognition
│
├── ── Status ──
├── status                ENUM (active, disabled, expired)
├── disabled_reason       TEXT (nullable)
├── expires_at            TIMESTAMPTZ (nullable)   ← Re-enrollment period
│
├── ── Consent ──
├── consent_id            UUID (FK → consent_record, ref K028)
├── consent_given_at      TIMESTAMPTZ
│
├── ── Audit ──
├── enrolled_at           TIMESTAMPTZ
├── enrolled_by           UUID (FK → user, staff yang melakukan enrollment)
└── created_at            TIMESTAMPTZ
```

```
biometric_verification_log:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
├── enrollment_id         UUID (FK → biometric_enrollment)
│
├── ── Verification ──
├── verification_type     ENUM (fingerprint, face, voice)
├── purpose               ENUM (transaction_approval, login, high_value_withdrawal,
│                                loan_approval, identity_verification, admin_action)
├── related_entity_type   VARCHAR (nullable, "transaksi", "pinjaman", dll)
├── related_entity_id     UUID (nullable)
│
├── ── Result ──
├── result                ENUM (success, failed, liveness_failed, no_match, error)
├── confidence_score      DECIMAL(5,2)    ← 0-100%
├── match_threshold       DECIMAL(5,2)    ← Threshold yang digunakan
├── failure_reason        TEXT (nullable)
│
├── ── Audit ──
├── verified_at           TIMESTAMPTZ
├── ip_address            VARCHAR (nullable)
├── device_info           VARCHAR (nullable)
└── created_at            TIMESTAMPTZ
```

### 4. When Biometric is Required

| Operation | Amount/Condition | Biometric Required | Type |
|---|---|---|---|
| Penarikan tunai | > Rp 25.000.000 | Ya | Fingerprint |
| Penarikan tunai | > Rp 50.000.000 | Ya | Fingerprint + Face |
| Transfer antar rekening | > Rp 25.000.000 | Ya | Fingerprint |
| Disbursement pinjaman | Semua | Ya (approver) | Fingerprint |
| Loan approval | > Rp 50.000.000 | Ya (approver) | Fingerprint |
| Rekening closure | Semua | Ya | Fingerprint |
| Password/PIN reset | Semua | Ya | Fingerprint or Face |
| Teller login | Semua | Ya | Fingerprint |
| Manager authorization | Semua | Ya | Fingerprint |
| Admin action (sensitive) | Semua | Ya | Fingerprint + Face |
| E-wallet transaction | > Rp 1.000.000 (student) | Ya (parent) | Face (mobile) |
| Balance inquiry | Semua | Tidak | - |
| Mini statement | Semua | Tidak | - |
| Setoran | Semua | Tidak | PIN saja cukup |

### 5. Liveness Detection (Anti-Spoofing)

```
Liveness Detection Requirements:
┌──────────────────────────────────────────────────────────┐
│ FINGERPRINT:                                             │
│ ├── Capacitive sensor (detects live tissue)             │
│ ├── Temperature check (if sensor supports)              │
│ └── Multi-finger enrollment (min 2 fingers)              │
│                                                          │
│ FACE RECOGNITION:                                        │
│ ├── Active liveness (blink, turn head, smile)            │
│ ├── 3D depth detection (if hardware supports)            │
│ ├── IR sensor for live detection (if hardware supports)  │
│ ├── Reject: photo, video, mask, deepfake                 │
│ └── Min 3 liveness challenges per verification           │
│                                                          │
│ VOICE (if used):                                         │
│ ├── Random phrase prompt (not pre-recorded)              │
│ ├── Background noise check                               │
│ └── Not primary — only supplementary                     │
└──────────────────────────────────────────────────────────┘
```

### 6. Fallback Mechanism

```
Fallback Chain:
┌──────────────────────────────────────────┐
│ Primary: Biometric (fingerprint/face)    │
└─────────────┬────────────────────────────┘
              │ Failed (3 attempts)
              v
┌──────────────────────────────────────────┐
│ Fallback 1: PIN + OTP                    │
│ ├── OTP via WhatsApp/SMS (valid 5 menit) │
│ └── PIN + OTP = two-factor               │
└─────────────┬────────────────────────────┘
              │ Failed
              v
┌──────────────────────────────────────────┐
│ Fallback 2: Manual Verification          │
│ ├── Staff verifies KTP + selfie photo    │
│ ├── Supervisor approval required          │
│ └── Logged with reason                   │
└─────────────┬────────────────────────────┘
              │ Failed
              v
┌──────────────────────────────────────────┐
│ Block: Transaction rejected              │
│ ├── Nasabah directed to branch office    │
│ ├── Investigation initiated              │
│ └── Report to compliance (if suspicious) │
└──────────────────────────────────────────┘
```

### 7. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Biometric requirement | Sama | Sama |
| DPS consent | Tidak perlu | DPS harus approve biometric policy |
| Liveness standard | Standard | Standard |
| Data retention | 5 tahun pasca closure | Sama + DPS audit trail |

### 8. Vernon _rels dan _data Structure

**Enrollment _rels:**
```json
{
  "tenant_id":   "018f...",
  "nasabah_id":  "018f..."
}
```

**Enrollment _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "biometric": {
    "type":         "fingerprint",
    "storage_type": "on_device",
    "status":       "active"
  }
}
```

### 9. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|---|---|---|---|---|
| Enroll own biometric | v | v | v | v |
| Enroll nasabah biometric | v | v | v | v |
| Verify biometric (transaction) | v | v | v | v |
| Disable biometric (own) | v | v | v | v |
| Disable nasabah biometric | - | v | v | v |
| View verification logs | - | v | v | v |
| Configure biometric thresholds | - | - | - | v |
| Force re-enrollment | - | - | v | v |
| Purge biometric data | - | - | - | v |
