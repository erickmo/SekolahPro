# 35 — Broadcast Communication Governance

**Status**: Draft  
**Perspective**: School Ops + Product Ops + Compliance  
**Primary Inputs**: `24-notification-messaging-operating-model.md`, `10-parent-student-member-experience.md`, `19-reporting-export-compliance-spec.md`

## Purpose

Mendefinisikan governance untuk komunikasi broadcast agar pengumuman massal, reminder, dan pesan penting terkirim efektif tanpa spam, abuse, atau pelanggaran privasi.

## Governance Principles

- broadcast harus punya owner, audience, dan tujuan jelas
- high-reach communication perlu kontrol lebih ketat daripada direct messaging
- urgent broadcast tidak boleh jadi alasan chaos channel
- preference dan privacy tetap dihormati dalam batas policy

## Control Areas

- audience targeting rules
- template approval untuk pesan sensitif/berdampak luas
- send window / quiet hour policy
- escalation untuk mistaken broadcast
- audit dan complaint review

## Acceptance Criteria

- `AC-FUNC`: institusi dapat mengirim broadcast penting dengan proses yang terkendali
- `AC-AUTH`: hanya role yang berwenang dapat mengirim ke audience besar
- `AC-DATA`: audience, template, delivery outcome, dan complaint signal tercatat
- `AC-AUDIT`: broadcast massal dapat direview ulang secara forensik
- `AC-INT`: governance selaras dengan notification preferences dan channel policy
- `AC-NFR`: proses kontrol cukup ketat tanpa menghambat komunikasi penting yang urgent

## Open Questions

- threshold audience yang memicu approval tambahan
- kapan mistaken broadcast perlu mandatory follow-up notice
