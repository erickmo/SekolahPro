import { prisma } from '../../lib/prisma';
import { callClaude } from '../../lib/claude';
import { NotFoundError, BadRequestError } from '../../shared/errors';
import { getPaginationParams, buildPaginationMeta } from '../../shared/types';
import { BloomLevel, Difficulty, QuestionType } from '@prisma/client';

interface ExamConfig {
  subject: string;
  topic: string;
  gradeLevel: string;
  curriculum: 'K13' | 'MERDEKA';
  difficulty: { easy: number; medium: number; hard: number };
  questionTypes: { multipleChoice: number; essay: number; trueOrFalse: number; matching: number };
  totalQuestions: number;
  timeLimit: number;
  bloomLevel?: BloomLevel;
}

interface GeneratedQuestion {
  question: string;
  type: QuestionType;
  options?: string[];
  answer: string;
  explanation?: string;
  bloomLevel: BloomLevel;
  difficulty: Difficulty;
}

export async function generateQuestions(schoolId: string, config: ExamConfig): Promise<GeneratedQuestion[]> {
  const systemPrompt = `Kamu adalah generator soal ujian untuk sekolah Indonesia.
Kurikulum: ${config.curriculum}. Tingkat kelas: ${config.gradeLevel}.
Buat soal berkualitas tinggi sesuai taksonomi Bloom yang diminta.
Selalu balas dalam format JSON yang valid.`;

  const userPrompt = `Buat ${config.totalQuestions} soal untuk:
Mata Pelajaran: ${config.subject}
Topik: ${config.topic}
Komposisi kesulitan: Mudah ${config.difficulty.easy}%, Sedang ${config.difficulty.medium}%, Sulit ${config.difficulty.hard}%
Jenis soal: PG ${config.questionTypes.multipleChoice}, Essay ${config.questionTypes.essay}, B/S ${config.questionTypes.trueOrFalse}, Menjodohkan ${config.questionTypes.matching}

Format JSON array:
[{
  "question": "teks soal",
  "type": "MULTIPLE_CHOICE|ESSAY|TRUE_FALSE|MATCHING",
  "options": ["A. ...", "B. ...", "C. ...", "D. ..."] (hanya untuk PG),
  "answer": "jawaban benar",
  "explanation": "penjelasan singkat",
  "bloomLevel": "C1|C2|C3|C4|C5|C6",
  "difficulty": "EASY|MEDIUM|HARD"
}]`;

  const raw = await callClaude(systemPrompt, userPrompt, 8000);
  const match = raw.match(/\[[\s\S]*\]/);
  if (!match) throw new BadRequestError('EXAM_001', 'Soal gagal di-generate AI');
  return JSON.parse(match[0]) as GeneratedQuestion[];
}

export async function saveGeneratedQuestions(schoolId: string, questions: GeneratedQuestion[], meta: { subject: string; topic: string; gradeLevel: string }) {
  const data = questions.map((q) => ({
    schoolId,
    subject: meta.subject,
    topic: meta.topic,
    gradeLevel: meta.gradeLevel,
    bloomLevel: q.bloomLevel,
    difficulty: q.difficulty,
    type: q.type,
    question: q.question,
    options: (q.options || null) as unknown as never,
    answer: q.answer,
    explanation: q.explanation,
    generatedBy: 'ai',
  }));
  await prisma.examQuestion.createMany({ data });
  return { count: data.length };
}

export async function getQuestions(schoolId: string, query: { subject?: string; topic?: string; gradeLevel?: string; difficulty?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.subject) where.subject = query.subject;
  if (query.topic) where.topic = { contains: query.topic, mode: 'insensitive' };
  if (query.gradeLevel) where.gradeLevel = query.gradeLevel;
  if (query.difficulty) where.difficulty = query.difficulty;
  const [questions, total] = await Promise.all([
    prisma.examQuestion.findMany({ where, skip, take: limit, orderBy: { createdAt: 'desc' } }),
    prisma.examQuestion.count({ where }),
  ]);
  return { questions, meta: buildPaginationMeta(total, page, limit) };
}

export async function createExam(schoolId: string, data: { title: string; subject: string; classId: string; semesterId: string; startTime: string; endTime: string; timeLimit: number; questionIds: string[]; createdBy: string }) {
  const exam = await prisma.exam.create({
    data: {
      title: data.title,
      subject: data.subject,
      schoolId,
      classId: data.classId,
      semesterId: data.semesterId,
      startTime: new Date(data.startTime),
      endTime: new Date(data.endTime),
      timeLimit: data.timeLimit,
      createdBy: data.createdBy,
      status: 'DRAFT',
      questions: {
        create: data.questionIds.map((qId, idx) => ({ questionId: qId, order: idx + 1 })),
      },
    },
    include: { questions: { include: { question: true } } },
  });
  return exam;
}

export async function getExams(schoolId: string, query: { classId?: string; semesterId?: string; status?: string; page?: number; limit?: number }) {
  const { page, limit, skip } = getPaginationParams(query);
  const where: Record<string, unknown> = { schoolId };
  if (query.classId) where.classId = query.classId;
  if (query.semesterId) where.semesterId = query.semesterId;
  if (query.status) where.status = query.status;
  const [exams, total] = await Promise.all([
    prisma.exam.findMany({ where, skip, take: limit, orderBy: { startTime: 'desc' }, include: { _count: { select: { questions: true, results: true } } } }),
    prisma.exam.count({ where }),
  ]);
  return { exams, meta: buildPaginationMeta(total, page, limit) };
}

export async function publishExam(schoolId: string, examId: string) {
  const exam = await prisma.exam.findFirst({ where: { id: examId, schoolId } });
  if (!exam) throw new NotFoundError('Ujian');
  return prisma.exam.update({ where: { id: examId }, data: { status: 'PUBLISHED' } });
}

export async function submitExamResult(schoolId: string, examId: string, studentId: string, answers: Record<string, string>) {
  const exam = await prisma.exam.findFirst({ where: { id: examId, schoolId }, include: { questions: { include: { question: true } } } });
  if (!exam) throw new NotFoundError('Ujian');

  let score = 0;
  let totalObjective = 0;
  for (const qmap of exam.questions) {
    const q = qmap.question;
    if (q.type !== 'ESSAY') {
      totalObjective++;
      const ans = answers[qmap.questionId];
      if (ans && ans.toLowerCase() === q.answer.toLowerCase()) score++;
    }
  }
  const objectiveScore = totalObjective > 0 ? (score / totalObjective) * 100 : 0;

  return prisma.examResult.upsert({
    where: { examId_studentId: { examId, studentId } },
    create: { examId, studentId, schoolId, answers: answers as unknown as never, objectiveScore, totalScore: objectiveScore, status: 'SUBMITTED' },
    update: { answers: answers as unknown as never, objectiveScore, status: 'SUBMITTED' },
  });
}

export async function getExamResults(schoolId: string, examId: string) {
  const exam = await prisma.exam.findFirst({ where: { id: examId, schoolId } });
  if (!exam) throw new NotFoundError('Ujian');
  return prisma.examResult.findMany({ where: { examId }, include: { student: { select: { name: true, nisn: true } } }, orderBy: { totalScore: 'desc' } });
}

export async function getExamAnalysis(schoolId: string, examId: string) {
  const results = await getExamResults(schoolId, examId);
  if (!results.length) return { message: 'Belum ada hasil ujian', stats: null };
  const scores = results.map((r) => r.totalScore);
  const avg = scores.reduce((s, v) => s + v, 0) / scores.length;
  const max = Math.max(...scores);
  const min = Math.min(...scores);
  const passing = results.filter((r) => r.totalScore >= 75).length;
  return { totalParticipants: results.length, average: Math.round(avg * 100) / 100, highest: max, lowest: min, passingRate: Math.round((passing / results.length) * 100), results };
}
