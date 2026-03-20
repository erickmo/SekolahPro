import { prisma } from '../../lib/prisma';
import { callClaude } from '../../lib/claude';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';

export async function chat(schoolId: string, userId: string, data: { subject: string; gradeLevel: string; message: string; sessionId?: string }) {
  const systemPrompt = `Kamu adalah tutor AI untuk siswa sekolah Indonesia.
Mata pelajaran: ${data.subject}. Tingkat kelas: ${data.gradeLevel}.
Berikan penjelasan yang mudah dipahami, gunakan bahasa Indonesia yang baik.
Jangan langsung memberikan jawaban PR — pandu siswa untuk menemukan jawabannya sendiri.`;

  const reply = await callClaude(systemPrompt, data.message);

  const interaction = await prisma.aIInteraction.create({
    data: { schoolId, userId, type: 'TUTOR', subject: data.subject, userMessage: data.message, aiResponse: reply, sessionId: data.sessionId },
  });

  return { reply, interactionId: interaction.id };
}

export async function checkPlagiarism(schoolId: string, data: { studentId: string; subjectId: string; content: string; compareWith?: string[] }) {
  const systemPrompt = 'Kamu adalah sistem deteksi plagiarisme akademik. Analisis teks dan berikan skor kemiripan (0-100) dan bagian yang terindikasi plagiat.';
  const userPrompt = `Analisis teks berikut untuk plagiarisme:\n\n${data.content}\n\nBerikan analisis dalam JSON: {"similarityScore": number, "suspiciousParts": string[], "verdict": "ORIGINAL|SUSPICIOUS|PLAGIAT", "notes": string}`;

  const result = await callClaude(systemPrompt, userPrompt);
  const match = result.match(/\{[\s\S]*\}/);
  const analysis = match ? JSON.parse(match[0]) : { similarityScore: 0, verdict: 'ORIGINAL', suspiciousParts: [], notes: 'Analisis gagal' };

  const record = await prisma.plagiarismCheck.create({
    data: { schoolId, studentId: data.studentId, subjectId: data.subjectId, content: data.content, similarityScore: analysis.similarityScore, verdict: analysis.verdict, details: analysis },
  });
  return { ...analysis, checkId: record.id };
}

export async function generateLearningPath(schoolId: string, studentId: string) {
  const grades = await prisma.grade.findMany({ where: { studentId, schoolId }, include: { subject: true }, orderBy: { score: 'asc' } });
  const weakSubjects = grades.filter((g) => g.score < 75).map((g) => `${g.subject.name}: ${g.score}`);

  if (!weakSubjects.length) return { message: 'Nilai siswa sudah baik semua', recommendations: [] };

  const result = await callClaude(
    'Kamu adalah konsultan pendidikan. Buat jalur belajar personal berdasarkan kelemahan siswa.',
    `Siswa memiliki nilai rendah pada: ${weakSubjects.join(', ')}. Buat learning path dalam JSON: {"recommendations": [{"subject": string, "priority": "HIGH|MEDIUM|LOW", "topics": string[], "resources": string[], "estimatedWeeks": number}]}`,
  );
  const match = result.match(/\{[\s\S]*\}/);
  return match ? JSON.parse(match[0]) : { recommendations: [] };
}

export async function getSentimentAnalysis(schoolId: string, query: { startDate?: string; endDate?: string }) {
  const where: Record<string, unknown> = { schoolId, type: 'SENTIMENT' };
  const interactions = await prisma.aIInteraction.findMany({ where, orderBy: { createdAt: 'desc' }, take: 100 });
  return { total: interactions.length, interactions };
}

export async function getInteractionHistory(schoolId: string, userId: string, query: { page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const [interactions, total] = await Promise.all([
    prisma.aIInteraction.findMany({ where: { userId, schoolId }, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.aIInteraction.count({ where: { userId, schoolId } }),
  ]);
  return { interactions, meta: buildPaginationMeta(total, page, limit) };
}
