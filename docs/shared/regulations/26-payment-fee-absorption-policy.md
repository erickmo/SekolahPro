# 26 — Payment Fee Absorption Policy

**Status**: Draft  
**Perspective**: CFO + Product + Finance Ops  
**Primary Inputs**: `ADR-S051`, `07-billing-metering-spec.md`, `23-financial-reconciliation-ops-spec.md`

## Purpose

Mendefinisikan kebijakan siapa yang menanggung biaya pembayaran per channel, bagaimana transparansinya ditampilkan, dan batas kebijakan komersial/regulasi agar fee model konsisten dan dapat dipertanggungjawabkan.

## Policy Principles

- fee policy harus selaras dengan regulasi channel
- transparansi ke parent/payer wajib saat fee dibebankan
- biaya tidak boleh menggerus margin tanpa disadari pada segmen tertentu
- exception komersial harus time-bound dan auditable

## Fee Bearer Modes

### School
- sekolah menyerap biaya sebagai operating/service cost
- cocok untuk premium service posture atau constraint regulasi tertentu

### Parent
- orang tua/payer membayar fee secara transparan
- cocok untuk VA/retail channels pada segmen tertentu bila diperbolehkan

### Split
- biaya dibagi sesuai kebijakan komersial
- perlu formula dan disclosure yang jelas

## Channel Guardrails

- **QRIS**: merchant/school bears fee sesuai catatan regulasi pada ADR
- **Virtual Account**: school atau parent sesuai policy tenant/segment
- **Bank Transfer Manual**: labor cost dan reconciliation cost harus dipertimbangkan walau fee provider nol
- **E-wallet/Retail**: hanya aktif bila economics dan disclosure cukup jelas

## Policy Inputs

- segment tenant
- package/tier
- channel type
- invoice/payment type
- promotional exception / launch subsidy

## Acceptance Criteria

- `AC-FUNC`: sistem dapat menentukan fee bearer dan total payable secara konsisten per transaksi
- `AC-DATA`: policy, exception, surcharge visibility, dan effective period tercatat
- `AC-AUDIT`: perubahan kebijakan dan alasan komersial/regulasi dapat ditelusuri
- `AC-INT`: payment creation, receipt, settlement, dan reconciliation memakai policy yang sama
- `AC-NFR`: policy lookup cukup cepat dan tidak membuat payment flow rumit bagi end user

## Open Questions

- batas maksimal split/surcharge final per segmen
- governance untuk promo/subsidi fee sementara
