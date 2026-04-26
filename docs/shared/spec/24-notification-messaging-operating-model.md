# 24 — Notification & Messaging Operating Model

**Status**: Draft  
**Perspective**: Product Ops + Engineering + Support  
**Primary Inputs**: `ADR-S043`, `ADR-S044`, `10-parent-student-member-experience.md`

## Purpose

Mendefinisikan operating model untuk komunikasi dua arah dan notifikasi satu arah agar experience pengguna, kontrol channel, SLA, dan support handling berjalan konsisten.

## Operating Principles

- messaging dan notification dipisahkan secara konsep, data, dan operasional
- urgency menentukan channel, bukan kebiasaan manual semata
- preference pengguna dihormati selama tidak melanggar kewajiban operasional/regulasi
- auditability dan privacy lebih penting daripada convenience channel liar

## Messaging Model

### Use Cases
- direct teacher-parent communication
- class/group communication
- institutional broadcast dengan respons terbatas

### Operational Rules
- conversation harus scoped jelas
- moderation/locking tersedia untuk konteks tertentu
- attachment dan message retention mengikuti policy
- escalation ke support/admin harus punya jalur jelas bila terjadi misuse

## Notification Model

### Channel Strategy
- push untuk default fast path
- SMS/WhatsApp untuk high-importance atau reach fallback
- email untuk dokumen formal dan summary tertentu
- in-app timeline untuk audit/personal history

### Delivery Policy
- idempotent by event/user/channel
- retry dan fallback per channel
- quiet hours berlaku untuk non-urgent events
- regulated/high-priority events dapat override preference tertentu sesuai policy

## Operational Queues

- pending delivery queue
- failed delivery / retry queue
- template and preference management
- abuse / complaint / opt-out handling

## Metrics & Support Signals

- send rate, delivery rate, read rate
- failed channel rate
- muted/conversation abandonment signals
- duplicate send incidents
- user complaint categories

## Acceptance Criteria

- `AC-FUNC`: notifikasi dan messaging mendukung use case utama tanpa saling tumpang tindih
- `AC-AUTH`: user hanya bisa mengirim/terima sesuai scope conversation dan event eligibility
- `AC-DATA`: message thread, delivery log, dan preference state konsisten dan traceable
- `AC-AUDIT`: event komunikasi penting dan perubahan preference tercatat
- `AC-INT`: event source ke notification engine dan message-triggered notification berjalan konsisten
- `AC-NFR`: high-volume notification flow tetap operasional tanpa merusak UX pesan dua arah

## Open Questions

- kapan WhatsApp jadi primary vs fallback channel
- policy retention final untuk message attachments dan read receipts
