# 05 — AI & Analitik (M23–M26b, M58)

Semua fitur AI menggunakan `claude-sonnet-4-20250514` via Anthropic API.
Semua modul di kelompok ini tier PRO ke atas.

```typescript
// Pola standar pemanggilan Claude AI di EDS
// packages/ai/client.ts
import Anthropic from '@anthropic-ai/sdk';
const client = new Anthropic(); // key dari env ANTHROPIC_API_KEY

export async function callClaude(
  systemPrompt: string,
  userContent: string,
  options?: { maxTokens?: number; schoolId?: string; moduleId?: string }
) {
  const response = await client.messages.create({
    model: 'claude-sonnet-4-20250514',
    max_tokens: options?.maxTokens ?? 1024,
    system: systemPrompt,
    messages: [{ role: 'user', content: userContent }],
  });
  // Catat usage untuk quota tracking
  if (options?.schoolId && options?.moduleId) {
    await recordAIUsage(options.schoolId, options.moduleId, response.usage);
  }
  return response.content[0].text;
}
```

---

## M23 — Early Warning System (EWS)
**Tier**: PRO | **Integrasi**: M01, M04, M05, M12, M28

Deteksi dini siswa berisiko dropout dari analisis multi-indikator.

```typescript
interface RiskIndicators {
  academic: {
    gradeDropPercent: number;       // % penurunan dari baseline
    failedSubjectsCount: number;    // mapel di bawah KKM
    homeworkCompletionRate: number; // % tugas dikumpulkan
  };
  attendance: {
    absenceRate: number;            // % ketidakhadiran bulan ini
    consecutiveAbsences: number;    // berapa hari berturut-turut
    lateArrivalFrequency: number;
  };
  behavioral: {
    disciplinaryRecords: number;
    extracurricularDropout: boolean;
    libraryVisitsDecline: number;
  };
  social: {
    counselingVisits: number;
    forumActivityDecline: number;
  };
}
```

**Output & Aksi**:
- [ ] Risk score 0–100 per siswa, diperbarui harian
- [ ] Alert otomatis ke guru BK & wali kelas saat skor > threshold (configurable, default 70)
- [ ] Widget EWS di dashboard M12
- [ ] Rekomendasi intervensi personal via Claude AI
- [ ] Log tindakan intervensi & hasilnya (untuk evaluasi efektivitas)

---

## M24 — Personalized Learning Path
**Tier**: PRO | **Integrasi**: M04, M05, M11, M21

- [ ] Analisis pola belajar dari nilai, waktu ujian, hasil per KD
- [ ] Generate rekomendasi topik prioritas per siswa (Claude AI)
- [ ] Saran materi: video YouTube, artikel, latihan soal dari bank soal (M05)
- [ ] Integrasi ke M11 (rekomendasi guru les yang cocok)
- [ ] Integrasi ke M21 (rekomendasi buku perpustakaan)
- [ ] Progress tracker visual di portal siswa
- [ ] Reminder belajar mingguan via M14

---

## M25 — Analitik Prediktif
**Tier**: ENTERPRISE | **Integrasi**: M04, M05, M18, M31

- [ ] Prediksi nilai akhir semester per siswa per mapel
- [ ] Proyeksi tingkat kelulusan per kelas (2 bulan sebelum)
- [ ] Prediksi siswa tidak naik kelas (trigger review dini)
- [ ] Forecasting pendapatan SPP (bantu perencanaan anggaran M51)
- [ ] Proyeksi kebutuhan guru & ruang kelas 3 tahun ke depan
- [ ] Benchmark performa vs. sekolah sejenis (data anonim, opt-in)

---

## M26 — AI Tutor & Anti-Plagiarisme
**Tier**: PRO (AI quota gabungan dengan M05) | **Integrasi**: M04, M15

### Sub-modul: AI Tutor

```typescript
// Chatbot tutor per mata pelajaran
// System prompt di-inject dengan konteks kurikulum siswa
const TUTOR_SYSTEM = `
Kamu adalah tutor ${subject} untuk siswa kelas ${grade} 
kurikulum ${curriculum}. Topik yang sedang dipelajari: ${currentTopics}.
Jawab pertanyaan dengan sabar dan gunakan contoh yang relevan 
dengan kehidupan siswa Indonesia.
Jangan langsung beri jawaban PR — bantu siswa berpikir sendiri.`;
```

- [ ] Chatbot tutor per mata pelajaran (konteks kurikulum real-time)
- [ ] Mode: tanya-jawab, penjelasan konsep, bantuan latihan soal
- [ ] Batasan: tidak menjawab soal ujian aktif (cek status ujian M05)
- [ ] Log percakapan untuk review guru (privasi: siswa tahu log tersimpan)

### Sub-modul: Anti-Plagiarisme

- [ ] Deteksi plagiarisme pada tugas yang dikumpulkan siswa
- [ ] Similarity check antar tugas dalam satu angkatan
- [ ] Deteksi parafrase (bukan hanya copy-paste) via AI
- [ ] Laporan detail: bagian yang terindikasi + sumber yang mirip
- [ ] Threshold configurable oleh guru (default: flagged jika > 30% mirip)

---

## M26b — Sentiment Analysis & Voice of School
**Tier**: PRO | **Integrasi**: M12, M13, M15

- [ ] Survei kepuasan otomatis: siswa, ortu, guru (tiap semester)
- [ ] Analisis sentimen dari pesan forum (M15) & chat ortu (M13)
- [ ] Dashboard indeks kepuasan sekolah (di M12)
- [ ] Identifikasi topik keluhan yang paling sering muncul
- [ ] Laporan bulanan untuk kepala sekolah
- [ ] Perbandingan tren sentimen antar semester

---

## M58 — Data Warehouse & Business Intelligence
**Tier**: ENTERPRISE | **Integrasi**: semua modul

### Arsitektur

```
Semua Modul (PostgreSQL per tenant)
    ↓  ETL Pipeline (Apache Airflow, jadwal: malam hari)
Data Warehouse (ClickHouse — kolumnar, cepat untuk agregasi)
    ↓
BI Layer (Apache Superset)
    ↓
Export / Scheduled Report
```

### Fitur

- [ ] ETL pipeline: transformasi data dari 62 modul ke warehouse
- [ ] Pre-built dashboard BI lintas modul
- [ ] Analisis korelasi lintas domain: misal korelasi gizi (M30) vs nilai (M04)
- [ ] Cohort analysis: perbandingan angkatan per angkatan
- [ ] Custom report builder — kepala sekolah drag-drop, tanpa SQL
- [ ] Natural language query: "tampilkan tren nilai kelas X semester 2" (Claude AI)
- [ ] Historical data multi-tahun untuk benchmarking
- [ ] Scheduled report: laporan terkirim otomatis tiap Senin pagi via email
- [ ] Export: CSV, Excel (skill `xlsx`), PDF (skill `pdf`)

**Catatan Multi-tenant**: Data warehouse di-isolasi per tenant. Query lintas tenant hanya bisa dilakukan oleh EDS superadmin untuk keperluan agregat anonim (benchmark industri).
