# 12 — API & Event Contract Spec

**Status**: Draft  
**Perspective**: System Analyst + System Engineer  
**Primary Inputs**: `ADR-004`, `ADR-005`, `ADR-014`, `ADR-015`, `03-cross-cutting-foundation.md`

## Purpose

Mendefinisikan aturan kontrak lintas HTTP API, internal events, sync events, idempotency, versioning, dan compatibility agar semua app dan service berbicara dengan pola yang konsisten.

## HTTP API Contract Rules

### URI & Versioning
- prefix standar: `/api/v1/...`
- version bump hanya untuk breaking contract
- resource naming konsisten, berbasis capability/domain

### Request / Response Shape
- pagination, filter, sort, dan lookup harus mengikuti pola seragam
- error response harus memiliki code, message, dan context minimum
- read model eventual-consistency harus mengungkapkan status/metadata bila relevan

### Auth & Scope Propagation
- semua request terproteksi membawa JWT sesuai mode tenant
- phase-2 scoped token wajib untuk endpoint operasional multi-tenant
- tenant/company/branch/warehouse scope tidak boleh diinfer diam-diam dari body jika sudah ada di token

### Idempotency
- create/import/payment-like endpoints harus mendukung idempotency key bila risiko duplikasi tinggi
- callback/integration endpoint wajib aman terhadap retry provider

## Event Contract Rules

### Event Envelope
Setiap event minimal memuat:
- `event_id`
- `event_type`
- `entity_type`
- `entity_id`
- `tenant_id`
- `company_id` bila applicable
- `occurred_at`
- `version`
- `correlation_id`
- `causation_id`
- `payload`

### Subject Naming
- `events.{domain}.{event_type}` untuk domain event
- `sync.{domain}.{event_type}` untuk sync/internal propagation
- penggunaan string literal bebas harus dihindari; pakai constants registry

### Delivery Semantics
- at-least-once delivery sebagai baseline
- consumer harus idempotent
- ordering hanya diasumsikan per subject yang terdokumentasi, bukan global
- retry, DLQ, dan replay harus memiliki prosedur operasional

## Contract Ownership

- producer owner bertanggung jawab atas backward compatibility event/API
- consumer tidak boleh bergantung pada field tidak terdokumentasi
- perubahan contract harus melewati review lintas producer-consumer

## Frontend Consumption Rules

- multi-app frontend (`admin`, `portal`, `pos`) mengonsumsi kontrak yang sama sesuai scope masing-masing
- generated client / shared contract package menjadi sumber kebenaran konsumsi frontend
- tidak boleh ada contract drift antar app

## Acceptance Criteria

- `AC-FUNC`: producer dan consumer lintas domain dapat berinteraksi tanpa coupling implisit
- `AC-AUTH`: seluruh contract menghormati auth dan scope model
- `AC-DATA`: event/API payload cukup untuk audit, replay, dan troubleshooting
- `AC-INT`: callback/retry/replay aman terhadap duplikasi dan partial failure
- `AC-NFR`: contract evolution tidak memaksa coordinated deploy yang tidak perlu

## Open Questions

- standar field error final
- pola idempotency header final untuk API publik
- daftar contract yang perlu schema registry formal
