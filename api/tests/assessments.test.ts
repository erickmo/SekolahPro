import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-as-001';
let guruToken: string;
let userId: string;
let rubricId: string;
let studentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'as-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Assessment Test School',
      npsn: `AS${Date.now()}`,
      subdomain: 'as-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const guruEmail = `guru_${Date.now()}@school.test`;
  const guruRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Guru Penilai', email: guruEmail, password: 'Password123!', role: 'GURU', schoolId: SCHOOL_ID });
  guruToken = guruRes.body.data.accessToken;
  userId = guruRes.body.data.user.id;

  const student = await prisma.student.create({
    data: {
      nisn: `AS${Date.now()}`.slice(-10),
      name: 'Assessment Student',
      birthDate: new Date('2009-04-10'),
      gender: 'MALE',
      birthPlace: 'Surabaya',
      religion: 'ISLAM',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.projectAssessment.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.assessmentRubric.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Assessment Rubrics (M49 Penilaian Proyek)', () => {
  it('POST /api/v1/assessments/rubrics - should create rubric', async () => {
    const res = await request(app)
      .post('/api/v1/assessments/rubrics')
      .set('Authorization', `Bearer ${guruToken}`)
      .send({
        teacherId: userId,
        subject: 'IPA',
        gradeLevel: '8',
        title: 'Rubrik Proyek Sains',
        criteria: { aspects: [] },
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Rubrik Proyek Sains');
    rubricId = res.body.data.id;
  });

  it('GET /api/v1/assessments/rubrics - should list rubrics', async () => {
    const res = await request(app)
      .get('/api/v1/assessments/rubrics')
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/assessments/rubrics/:id - should get rubric', async () => {
    if (!rubricId) return;
    const res = await request(app)
      .get(`/api/v1/assessments/rubrics/${rubricId}`)
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(rubricId);
  });
});

describe('Project Assessments', () => {
  it('POST /api/v1/assessments/projects - should create project assessment', async () => {
    if (!rubricId) return;
    const res = await request(app)
      .post('/api/v1/assessments/projects')
      .set('Authorization', `Bearer ${guruToken}`)
      .send({
        rubricId,
        studentId,
        assessorType: 'TEACHER',
        scores: { aspect1: 90 },
        totalScore: 90,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
  });

  it('GET /api/v1/assessments/projects - should list project assessments', async () => {
    const res = await request(app)
      .get('/api/v1/assessments/projects')
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/assessments/projects/student/:studentId - should get student assessments', async () => {
    const res = await request(app)
      .get(`/api/v1/assessments/projects/student/${studentId}`)
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
