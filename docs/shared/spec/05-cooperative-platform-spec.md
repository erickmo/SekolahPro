# 05 — Cooperative Platform Spec

**Status**: Draft  
**Perspective**: C-Level + System Analyst + System Engineer  
**Scope**: Management Koperasi Sekolah

## Purpose

Mendefinisikan capability, dependency, compliance, dan sequencing untuk domain koperasi/BMT berdasarkan ADR-K series.

## Primary Actors

- admin koperasi
- teller
- manager koperasi
- pengurus
- anggota/nasabah
- auditor/internal control
- bendahara yayasan/sekolah
- regulator / sistem eksternal

## Capability Domains

### 1. Cooperative Core
**ADRs**: K001, K002, K003  
**Goal**: fondasi data anggota, rekening, produk, dan akad.

### 2. Savings & Financing
**ADRs**: K004-K010  
**Goal**: mengelola simpanan, pembiayaan, angsuran, penalti, dan agunan.

### 3. Operations & Transactions
**ADRs**: K011-K014  
**Goal**: engine transaksi operasional koperasi yang bisa diaudit.

### 4. Accounting & Distribution
**ADRs**: K015-K018  
**Goal**: jurnal otomatis, SHU, regulasi, dan mode syariah/islamic.

### 5. Governance, Compliance & Sustainability
**ADRs**: K024, K027, K028, K031, K033, K034, K035, K038  
**Goal**: integrasi, AML/CFT, privacy, indikator kesehatan, reserve fund, takaful, dissolution.

## Rollout by Wave

### Wave 1 — Cooperative Core
- member/nasabah
- rekening
- produk/akad
- tabungan / simpanan dasar

### Wave 2 — Financing & Transaction Engine
- pembiayaan/pinjaman
- angsuran
- denda dan agunan
- transaksi rekening/non-rekening
- teller session dan kas

### Wave 3 — Accounting, Compliance, Enterprise Readiness
- jurnal & COA
- SHU
- laporan regulasi
- zakat/infaq
- AML/CFT
- UU PDP
- health indicators
- reserve fund
- takaful
- dissolution

## Strategic Constraints

- fitur koperasi/BMT adalah monetization expansion, bukan entry point utama
- launch capability finansial harus tunduk pada readiness compliance
- wallet / payment rail harus dibatasi oleh guardrail legal dan operasional
- accounting dan regulatory reporting tidak boleh dianggap add-on opsional untuk tenant enterprise aktif

## Dependency Rules

- rekening bergantung pada member dan produk/akad
- simpanan/pembiayaan bergantung pada rekening dan produk
- transaksi bergantung pada rekening aktif dan otorisasi operasional
- jurnal dan laporan regulasi bergantung pada engine transaksi yang stabil
- capability compliance mengunci go-live untuk module regulated

## Acceptance Baseline

- `AC-FUNC`: operasional koperasi dasar berjalan tanpa proses manual utama
- `AC-DATA`: saldo, transaksi, dan jurnal konsisten dan bisa direkonsiliasi
- `AC-AUTH`: aksi finansial dibatasi oleh role, scope, dan approval
- `AC-AUDIT`: seluruh transaksi kritikal memiliki jejak audit lengkap
- `AC-INT`: integrasi pembayaran dan pihak ketiga dapat diverifikasi dan direplay
- `AC-NFR`: sistem tahan terhadap retry, replay, partial failure, dan kebutuhan close-of-day
