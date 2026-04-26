# 14 — Koperasi Regulated Finance Guardrails

**Status**: Draft  
**Perspective**: C-Level + Risk + System Engineer  
**Primary Inputs**: `ADR-K027`, `ADR-K028`, `ADR-K031`, `ADR-K038`, `ADR-018`, `08-compliance-risk-controls.md`

## Purpose

Mendefinisikan guardrail operasional dan teknis untuk capability koperasi/BMT yang berada di area regulated finance agar pertumbuhan produk tidak melampaui kesiapan compliance, risk, dan operability.

## Guardrail Principles

- compliance lebih dulu daripada ekspansi fitur
- suspicious > investigate > decide, bukan auto-ignore
- stale reporting lebih baik daripada keputusan tanpa kontrol
- freeze capability tertentu lebih baik daripada membuka risiko sistemik

## Regulated Capability Zones

### Zone A — Customer Identity & Due Diligence
- KYC/CIP dasar
- CDD/EDD berdasarkan risk level
- PEP screening
- beneficial ownership
- review period berkala

### Zone B — Transaction Monitoring & Reporting
- threshold monitoring
- velocity/pattern/anomaly detection
- suspicious alert investigation
- LTKM/TKM workflow dan evidence
- retention 5 tahun pasca penutupan relasi sesuai kewajiban relevan

### Zone C — Privacy & Sensitive Financial Data
- financial PII diklasifikasikan highly confidential
- consent/legal basis tercatat
- access masking untuk KTP/NPWP/saldo/transaksi detail sesuai kebutuhan
- breach handling dan data subject request harus bisa dijalankan

### Zone D — Health & Sustainability Controls
- CAR, NPL, BOPO, liquidity, concentration, SHU, dan indikator kesehatan lain
- early warning dan escalation
- corrective action tracking
- stress test dan board/regulator reporting cadence

### Zone E — Exit / Dissolution Controls
- dissolution hanya boleh diinisiasi oleh actor yang tepat
- claim period, liquidation, distribution, dan closure harus mengikuti stage control
- archive dan decommission tidak boleh memutus audit trail yang diwajibkan

## Release Guardrails

- capability Zone B-E tidak boleh go-live tanpa owner risk/compliance yang jelas
- automated alert tanpa investigation workflow dianggap belum siap produksi
- auto-blocking transaksi hanya boleh aktif jika false-positive handling dan override governance siap
- fitur yang menyerupai e-money terbuka harus diblok sampai legal basis dan licensing position jelas

## Operational Guardrails

- high-risk alert wajib punya SLA investigasi
- evidence pack untuk PPATK/OJK/Dinas harus dapat dibentuk dari sistem
- sensitive override memerlukan dual control atau approval sesuai tingkat risiko
- dashboard kesehatan koperasi harus tersedia untuk manajemen sebelum scale transaksi tinggi

## Acceptance Criteria

- `AC-FUNC`: guarded workflow tersedia untuk onboarding, monitoring, reporting, dan dissolution-related flows
- `AC-AUTH`: akses finansial sensitif dibatasi oleh role, scope, dan approval tier
- `AC-DATA`: suspicious alert, consent, retention, dan health indicator dapat ditelusuri end-to-end
- `AC-AUDIT`: regulator-facing dan board-facing evidence dapat dihasilkan tanpa rekonstruksi manual besar
- `AC-INT`: threshold/report/export flows tahan terhadap retry, correction, dan submission ulang
- `AC-NFR`: control layer tidak merusak operasi teller/transaction engine secara tidak proporsional

## Open Questions

- ambang auto-block vs manual-review final
- owner persetujuan untuk EDD high-risk final
- kapan threshold tertentu memicu readiness review ulang untuk skala koperasi
