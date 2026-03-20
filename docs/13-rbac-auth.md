# 13 — Auth & RBAC

---

## Role Hierarchy

```
EDS_SUPERADMIN              ← Tim EDS (Developer Portal full)
EDS_SUPPORT                 ← Tim support (impersonation, read-only)
EDS_SALES                   ← Tim sales (onboarding sekolah)
LISTING_MANAGER             ← Kelola vendor listing

YAYASAN_ADMIN               ← Kelola semua tenant milik yayasan (M43)

ADMIN_SEKOLAH               ← Admin teknis per sekolah
  ├── KEPALA_SEKOLAH         ← Read-all + approve
  ├── KEPALA_KURIKULUM       ← M04, M05, M48, M49, M50 full
  ├── BENDAHARA              ← M03, M09, M18, M22, M51, M52
  ├── GURU                   ← M04, M05, M06, M26, M48, M49
  ├── WALI_KELAS             ← M01/M04 kelas sendiri, M13
  ├── GURU_BK                ← M23, M28, M29, M55
  ├── GPK                    ← M55 full (data ABK)
  ├── OPERATOR_SIMS          ← M01, M02, M08, M56
  ├── KASIR_KOPERASI         ← M03, M18, M19
  ├── PUSTAKAWAN             ← M21
  ├── PETUGAS_UKS            ← M27, M30
  ├── TATA_USAHA             ← M01, M07, M09, M20, M31, M57
  ├── SECURITY_OFFICER       ← M45, M47 (audit & security)
  └── KOMITE_SEKOLAH         ← M52, M53 (read keuangan, voting)

SISWA                        ← M02, M05, M06, M13, M15, M17, M19, M21, M24, M37, M60
ORANG_TUA                    ← M13, M18, M19, M22, M52, M53, M54
ALUMNI                       ← M36, M61 (upload konten)
GURU_LES                     ← M11 (profil & jadwal sendiri)
MITRA_DU_DI                  ← M38 (nilai & jurnal prakerin)
PUBLISHER                    ← M10 (katalog & pesanan)
DEVELOPER                    ← M44 (API + sandbox)
VOLUNTEER                    ← M54 (jadwal & rekam kontribusi)
LISTING_VENDOR               ← Developer Portal listing management
```

---

## JWT & Session

```typescript
interface JWTPayload {
  sub: string;          // userId
  schoolId: string;     // null untuk EDS_SUPERADMIN
  role: UserRole;
  tier: TenantTier;     // tier sekolah — untuk feature gate cepat
  iat: number;
  exp: number;          // 15 menit untuk access token
}

// Refresh token: 7 hari, disimpan di httpOnly cookie
// Access token: 15 menit, disimpan di memory (bukan localStorage)
```

---

## Permission Matrix (Ringkasan)

| Modul | Kepsek | Ka.Kurikulum | Guru | Wali Kelas | Siswa | Ortu |
|-------|--------|--------------|------|------------|-------|------|
| M01 SIMS | R | R | R (kelas) | RW (kelas) | R (self) | R (anak) |
| M04 Nilai | R | RW | RW (mapel) | R (kelas) | R (self) | R (anak) |
| M05 Ujian | R | RW | RW | - | R (ikut) | - |
| M03 Tabungan | R | - | - | - | R (self) | R (anak) |
| M18 SPP | R | - | - | - | - | RW |
| M28 BK | R | - | - | - | RW (self) | - |
| M45 Audit | R | - | - | - | - | - |
| M47 Security | R | - | - | - | - | - |
| M51 RKAS | R | R | - | - | - | - |
| M52 Dana Komite | R | - | - | - | - | R |
| M55 ABK | R | R | R (kelas) | - | - | R (anak) |

---

## SSO & OAuth

Untuk integrasi M41 (LMS), siswa dan guru bisa SSO ke platform eksternal:

```typescript
// OAuth2 PKCE flow
// EDS sebagai Identity Provider (IdP)
const oauthConfig = {
  issuer: 'https://eds.id',
  authorizationEndpoint: '/oauth/authorize',
  tokenEndpoint: '/oauth/token',
  scopes: ['openid', 'profile', 'email', 'student_id', 'school_id'],
  supportedClients: ['google-classroom', 'moodle', 'ruangguru'],
};
```
