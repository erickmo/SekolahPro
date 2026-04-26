# 13 — Sistem Komunikasi

Modul Komunikasi menyediakan infrastruktur komunikasi digital sekolah yang menggantikan ketergantungan pada grup WhatsApp tidak terstruktur. Terdiri dari lima sub-domain: Portal Orang Tua (S042), Messaging (S043), Notification System (S044), Pengumuman & Berita (S045), dan Surat Menyurat TU (S046).

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S042 | Parent Portal & Dashboard | Vernon |
| ADR-S043 | Communication & Messaging | Vernon |
| ADR-S044 | Notification System | CQRS |
| ADR-S045 | Announcement & News | Vernon |
| ADR-S046 | Correspondence Management | Vernon |

---

## Domain Entities

### Parent Portal (S042)

**`parent_portal_settings`** — Preferensi portal per orang tua:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `user_id` | UUID | FK → users (unik per tenant+company) |
| `guardian_id` | UUID | FK → student_guardians (S003) |
| `language` | VARCHAR(5) | `id`, `en`, `ar` |
| `default_child_id` | UUID | Anak yang ditampilkan pertama (nullable) |
| `dashboard_layout` | JSONB | Konfigurasi widget dan urutan tampilan |
| `notification_prefs` | JSONB | Channel preference per event type |
| `theme` | VARCHAR(20) | `light`, `dark`, `auto` |

**View `parent_children_view`:** query convenience untuk daftar anak dari `student_guardians`.

Portal **tidak menyimpan data domain lain** — semua data dibaca langsung dari domain source (S001–S018) via API aggregation.

### Communication & Messaging (S043)

**`conversations`** — Thread percakapan:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `conversation_type` | VARCHAR(20) | `direct` (1-to-1), `class_group` (wali kelas → parents), `broadcast` (sekolah → semua), `custom_group` |
| `context_type` / `context_id` | VARCHAR + UUID | Scope polymorphic: `class_room`, `student`, `academic_year` |
| `last_message_preview` | TEXT | Denormalisasi preview pesan terakhir |
| `message_count` | INT | Counter total pesan (denormalisasi) |
| `is_archived` | BOOLEAN | Tersembunyi dari list default |
| `is_locked` | BOOLEAN | Hanya admin bisa kirim pesan |

**`conversation_participants`** — Peserta percakapan:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `user_id` | UUID | Peserta |
| `role` | VARCHAR(20) | `admin`, `member`, `readonly` |
| `last_read_at` | TIMESTAMPTZ | Waktu baca terakhir |
| `unread_count` | INT | Jumlah pesan belum dibaca (per peserta) |
| `is_muted` | BOOLEAN | Mute percakapan |

**`messages`** — Pesan individual:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `conversation_id` | UUID | FK |
| `sender_id` | UUID | Pengirim |
| `message_type` | VARCHAR(20) | `text`, `image`, `file`, `system` |
| `content` | TEXT | Isi pesan |
| `attachments` | JSONB | Array `{filename, url, mime_type, size_bytes}` (max 5 file, max 10MB/file) |
| `reply_to_id` | UUID | Threading — reply ke pesan tertentu |
| `is_edited` | BOOLEAN | Pesan pernah diedit (max 15 menit setelah kirim) |

### Notification System (S044)

**Menggunakan CQRS Pattern** (bukan Vernon) karena write-heavy, append-only, dan tidak memerlukan relationship autoloading.

**`notification_templates`** — Template per event type per channel:

| Field | Keterangan |
|-------|-----------|
| `event_type` | Format: `domain.event`, misal `attendance.absent`, `finance.payment_due` |
| `channel` | `push`, `sms`, `email`, `whatsapp`, `in_app` |
| `title_template` / `body_template` | Menggunakan placeholder `{{variable}}` |

**`notification_preferences`** — Preferensi per user:

| Field | Keterangan |
|-------|-----------|
| `preferences` | JSONB — channel per event type dengan flag enabled |
| `quiet_hours_start` / `quiet_hours_end` | Jam tenang (default 22:00–06:00) |
| `timezone` | Default `Asia/Jakarta` |

**`notification_logs`** — Log terkirim (append-only):

| Field | Keterangan |
|-------|-----------|
| `idempotency_key` | Unik: `{event_type}:{event_id}:{channel}:{user_id}` |
| `status` | `pending` → `sent` → `delivered` → `read` / `failed` |
| `retry_count` | Auto-retry max 3x |
| `source_type` + `source_id` | Polymorphic reference ke entity sumber |

**`user_device_tokens`** — Token FCM per device:

| Field | Keterangan |
|-------|-----------|
| `platform` | `android`, `ios`, `web` |
| `device_token` | Unik secara global |

### Announcements & News (S045)

**`announcements`** — Pengumuman resmi sekolah:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `target_audience` | VARCHAR(20) | `all`, `teachers`, `parents`, `students`, `specific_classes`, `specific_roles` |
| `target_roles` | JSONB | Array role jika `specific_roles` |
| `target_class_ids` | JSONB | Array UUID kelas jika `specific_classes` |
| `priority` | VARCHAR(10) | `low`, `normal`, `high`, `urgent` |
| `publish_at` | TIMESTAMPTZ | Scheduled publishing |
| `expire_at` | TIMESTAMPTZ | Auto-expire (nullable) |
| `is_pinned` | BOOLEAN | Selalu tampil di atas |
| `category` | VARCHAR(30) | `general`, `academic`, `finance`, `event`, `health`, `emergency`, `religious` |
| `total_target` | INT | Dihitung saat publish (denormalisasi) |
| `read_percentage` | NUMERIC(5,2) | `total_read / total_target × 100` (denormalisasi, async) |
| `status` | VARCHAR(20) | `draft` → `published` → `expired` / `archived` |

**`announcement_reads`** — Tracking pembaca:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `announcement_id` | UUID | FK |
| `user_id` | UUID | FK |
| `read_at` | TIMESTAMPTZ | Waktu baca |
| `acknowledged` | BOOLEAN | Konfirmasi baca (untuk `urgent`) |

### Correspondence Management (S046)

**`correspondences`** — Surat masuk & keluar:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `letter_number` | VARCHAR(100) | Nomor surat (nullable — diisi saat finalize untuk keluar) |
| `direction` | VARCHAR(10) | `incoming` atau `outgoing` |
| `classification` | VARCHAR(30) | `surat_dinas`, `surat_keterangan`, `surat_tugas`, `surat_undangan`, `surat_edaran`, `sk_kepala_sekolah` |
| `sequence_number` | INT | Auto-increment per fiscal_year per classification (untuk penomoran) |
| `sender_name` / `recipient_name` | VARCHAR(255) | Denormalisasi — surat eksternal tidak punya user_id |
| `body` | TEXT | Isi surat (nullable untuk scan surat masuk) |
| `template_id` | UUID | FK → correspondence_templates (nullable) |
| `signature_type` | VARCHAR(20) | `wet`, `digital`, `stamped` |
| `status` | VARCHAR(20) | `draft` → `registered` → `in_disposition` → `completed` → `archived` |

**`correspondence_templates`** — Template surat:

| Field | Keterangan |
|-------|-----------|
| `body_template` | Go template / Handlebars syntax |
| `required_variables` | JSONB array variabel yang harus diisi |
| `number_format` | Format penomoran, contoh: `{kode_sekolah}/{nomor_urut}/{bulan_romawi}/{tahun}` |
| `version` | Versi template (auto-increment saat update) |

**`correspondence_dispositions`** — Disposisi/routing surat masuk:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `correspondence_id` | UUID | FK |
| `from_user_id` | UUID | Pengirim disposisi |
| `to_user_id` | UUID | Penerima disposisi |
| `disposition_type` | VARCHAR(30) | `for_action`, `for_review`, `for_information`, `for_signature`, `for_filing` |
| `level` | INT | Level chain (kepsek=1, wakasek=2, staf=3) |
| `due_date` | DATE | Deadline penyelesaian |
| `status` | VARCHAR(20) | `pending` → `read` → `in_progress` → `completed` / `forwarded` |

---

## Business Rules

### Parent Portal

1. **Akses dibatasi ke anak yang ter-link di S003 (guardian).** Orang tua tidak bisa mengakses data anak orang lain — 403 Forbidden.
2. **Portal bersifat read-only** — orang tua tidak dapat mengubah data akademik, keuangan, atau kehadiran.
3. **Portal mendukung multi-anak secara native.** Semua anak yang ter-link via S003 otomatis muncul tanpa setup manual.
4. **Guardian harus ter-link ke user account** sebelum bisa mengakses portal. Proses onboarding diperlukan.
5. **Untuk pesantren (ADR-009 dual-mode),** endpoint tambahan tersedia: hafalan, nilai diniyah.

### Messaging

6. **Conversation class_group di-generate otomatis** dari data placement siswa (S014) — wali kelas pilih kelas, sistem auto-invite semua orang tua.
7. **Pesan hanya bisa diedit dalam 15 menit setelah dikirim.**
8. **Conversation `broadcast`** bisa di-lock (`is_locked = true`) agar hanya admin yang dapat mengirim.
9. **Unread count dijaga per peserta** — setiap pesan baru menambah `unread_count` semua peserta kecuali pengirim.
10. **Attachment max 5 file per pesan, max 10MB per file.**

### Notification System

11. **Idempotency key mencegah duplikat.** Format: `{event_type}:{event_id}:{channel}:{user_id}`. Notifikasi yang sama tidak dikirim dua kali.
12. **Quiet hours:** notifikasi non-urgent di-queue selama jam tenang dan dikirim setelah quiet hours berakhir.
13. **Auto-retry max 3x** dengan exponential backoff untuk channel eksternal yang gagal.
14. **MDR QRIS ditanggung merchant** per regulasi Bank Indonesia — `fee_bearer` untuk channel QRIS harus selalu `school`.
15. **WhatsApp Business API memerlukan template yang diapprove Meta** sebelum bisa digunakan — lead time perlu diperhitungkan.
16. **Event-driven:** domain lain mengirim event, notification service subscribe dan resolve recipients dari data guardian/relasi.

### Announcements

17. **Pengumuman tidak bisa diedit setelah dipublikasikan** — status berubah menjadi immutable.
18. **Pengumuman `urgent` memerlukan acknowledgement** dari target audience sebelum dianggap dibaca.
19. **Auto-expire** dijalankan via cron job — status berubah ke `expired` setelah `expire_at` terlewat.
20. **Kategori `religious`** tersedia untuk pengumuman kegiatan keagamaan pesantren.

### Correspondence

21. **Nomor surat keluar di-generate secara atomic** menggunakan `SELECT ... FOR UPDATE` untuk mencegah nomor ganda saat akses konkuren.
22. **Format nomor surat dapat dikustomisasi per sekolah** via `correspondence_templates.number_format`.
23. **Surat yang sudah ditandatangani tidak bisa diedit** — immutable setelah `signed_at` diisi.
24. **Disposisi adalah routing/assignment, bukan approval.** Workflow disposisi berbeda dari S047 approval workflow — disposisi tidak menggunakan approve/reject, melainkan `for_action`, `for_review`, dll.

---

## Channel Notifikasi & Biaya

| Channel | Provider | Biaya Estimasi |
|---------|----------|----------------|
| **Push (FCM)** | Firebase | Gratis |
| **In-App** | Internal | Gratis |
| **Email** | SMTP/SendGrid | Sangat murah (< Rp 1/email) |
| **SMS** | Zenziva/Twilio/dll | Rp 350–500/SMS |
| **WhatsApp** | Fonnte/Wablas/dll | Rp 100–500/pesan + template approval |

**Rekomendasi default:**
- Kehadiran absent, pelanggaran → Push + WhatsApp
- Tagihan jatuh tempo → Push + SMS
- Nilai dipublikasikan → Push saja
- Pengumuman → Push saja

---

## Key Decisions & Rationale

1. **Notification menggunakan CQRS** (bukan Vernon) karena write-heavy, append-only, dan tidak butuh relationship autoloading. Volume bisa ratusan ribu per hari.
2. **Portal tidak menyimpan data duplikat** — aggregation dari domain source saat request. Trade-off: N+1 query risk (mitigasi: parallel query + caching).
3. **Messaging dibatasi di dalam sistem** — WhatsApp hanya sebagai delivery channel untuk notifikasi, bukan platform komunikasi utama. Alasan: WhatsApp API mahal dan terlalu bergantung pada Meta.
4. **Correspondence terpisah dari S047** — disposisi TU adalah routing (kepada siapa diproses), bukan approval (setuju/tolak). S047 digunakan untuk persetujuan surat keluar penting, bukan disposisi rutin.
5. **Announcement dibedakan dari messaging** — pengumuman adalah komunikasi resmi satu arah dengan read tracking, berbeda dari percakapan dua arah.
6. **`notification_logs` adalah append-only, tidak punya `deleted_at`** — audit trail permanen.

---

## Integration Points

### Internal

| Domain | Arah | Deskripsi |
|--------|------|-----------|
| ADR-S001–S018 | → S042 | Portal membaca data dari semua domain akademik dan keuangan |
| ADR-S003 Guardians | → S042, S043 | Link orang tua–anak untuk akses portal dan auto-group messaging |
| S008 Attendance | → S044 | Event `absent` → notifikasi ke orang tua |
| S009 Finance | → S044 | Event tagihan jatuh tempo → notifikasi |
| S011 Grades | → S044 | Event nilai dipublikasikan → notifikasi |
| S012 Discipline | → S044 | Event pelanggaran → notifikasi ke orang tua |
| S043 Messaging | → S044 | Pesan baru → push notification ke penerima |
| S045 Announcements | → S044 | Pengumuman baru → notifikasi ke target audience |
| S046 Correspondence | → S044 | Disposisi baru → notifikasi ke penerima disposisi |
| S047 Approval | → S044 | Step approval aktif → notifikasi ke approver |
| ADR-011 ClassRoom | → S043 | Auto-group class_group menggunakan data kelas |
| S048 School Profile | → S046 | Header surat otomatis dari data profil sekolah |

### Eksternal

| Sistem | Keterangan |
|--------|-----------|
| **Firebase Cloud Messaging (FCM)** | Push notification mobile & web |
| **SMS Gateway** (Zenziva, Twilio, dll) | Fallback untuk orang tua tanpa smartphone |
| **WhatsApp Business API** (Fonnte, Wablas, dll) | Channel paling efektif di Indonesia |
| **SMTP / SendGrid** | Email untuk dokumen formal dan laporan |
| **BSrE / BSSN** | Tanda tangan digital untuk S046 (deferred — memerlukan infrastruktur tambahan) |
