# 29 — Cross-Domain Event Catalog

**Status**: Draft  
**Perspective**: System Analyst + Integration Architect + Engineering  
**Primary Inputs**: `ADR-005`, `12-api-event-contract-spec.md`, `ADR-S044`, `ADR-014`

## Purpose

Mendefinisikan katalog event lintas domain sebagai referensi operasional dan desain untuk producer, consumer, observability, dan replay/recovery.

## Catalog Principles

- subject naming konsisten dan terdokumentasi
- event owner, primary consumers, dan business meaning harus jelas
- event catalog tidak menggantikan contract detail, tetapi menjadi peta hubungan lintas domain
- replay sensitivity dan idempotency expectation harus ditandai

## Event Families

### Domain Events
- `events.students.*`
- `events.finance.*`
- `events.koperasi.*`
- `events.approval.*`
- `events.messaging.*`
- `events.reporting.*`

### Sync / Internal Events
- `sync.*` untuk Vernon propagation dan internal cache maintenance

### Integration / Callback-Derived Events
- payment callback processed
- settlement reconciled
- notification delivery result
- import/export lifecycle events

## Catalog Entry Minimum

Setiap entry wajib punya:
- subject
- event_type
- producer
- trigger condition
- payload summary
- idempotency expectation
- primary consumers
- replay sensitivity
- business/operational consequence if missed

## Operational Uses

- onboarding engineer ke event ecosystem
- impact analysis saat change contract
- incident debugging dan replay scope selection
- governance review untuk event sprawl

## Acceptance Criteria

- `AC-FUNC`: tim dapat memahami siapa menerbitkan event apa dan siapa yang mengonsumsinya
- `AC-DATA`: event catalog cukup untuk menuntun audit, observability, dan troubleshooting awal
- `AC-AUDIT`: perubahan event penting dapat dibandingkan terhadap catalog baseline
- `AC-INT`: catalog selaras dengan contract spec dan event bus subject convention
- `AC-NFR`: catalog dapat dipelihara tanpa overhead dokumentasi yang berlebihan

## Open Questions

- apakah catalog dijaga manual atau digenerate sebagian dari code/contracts
- event mana yang wajib punya replay runbook terpisah
