# 13 — Security, Privacy & Data Lifecycle Spec

**Status**: Draft  
**Perspective**: Security + Risk + Engineering  
**Primary Inputs**: `ADR-018`, `ADR-004`, `ADR-013`, `08-compliance-risk-controls.md`

## Purpose

Merinci kontrol keamanan, perlindungan privasi, dan lifecycle data dari pembuatan sampai penghapusan agar sistem aman, patuh, dan operasional.

## Security Domains

### Identity & Access
- authentication lifecycle
- JWT issuance, refresh, revocation posture
- scoped access enforcement
- privileged access dan support impersonation guardrail

### Secrets & Key Management
- secret tidak boleh berada di repo/image/config non-secret
- rotation policy untuk secret kritikal
- akses secret dibatasi per environment dan role

### Application Security
- input validation
- upload safety
- audit untuk sensitive action
- protection terhadap misuse pada import, messaging, payment, dan admin endpoints

## Privacy Domains

### Child Data Protection
- data siswa diklasifikasikan high sensitivity
- consent wali sebagai gate pemrosesan
- visibility minimum untuk role non-esensial

### Data Minimization
- hanya data yang diperlukan yang dikumpulkan dan ditampilkan
- observability dan support workflow harus meminimalkan PII exposure

### Portability & Erasure
- request portability ter-track sampai selesai
- erasure menghapus PII namun menjaga audit non-PII yang diperlukan

## Data Lifecycle

### States
- created
- active
- archived
- anonymized
- deleted / erased

### Lifecycle Rules
- retention policy harus configurable per tenant dalam batas legal
- archive/anonymize/delete wajib terdokumentasi dan auditable
- backup/restore tidak boleh melanggar janji erasure tanpa prosedur kompensasi yang jelas

### Data Classes
- public
- internal
- confidential
- regulated/high-sensitivity

## Acceptance Criteria

- `AC-AUTH`: akses mengikuti least privilege dan scope yang eksplisit
- `AC-DATA`: lifecycle data berjalan konsisten termasuk archive, anonymize, erase
- `AC-AUDIT`: semua sensitive access dan sensitive mutation dapat diaudit
- `AC-INT`: export, delete, dan portability flow aman terhadap retry dan partial completion
- `AC-NFR`: security/privacy controls tidak membuat sistem inti tidak usable

## Open Questions

- model support impersonation final
- strategi revocation token untuk incident response
- pemetaan final data class per domain besar
