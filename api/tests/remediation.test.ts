import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-rm-001';
let guruToken: string;
let userId: string;
let programId: string;
let sessionId: string;
let studentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'rm-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Remediation Test School',
      npsn: `RM${Date.now()}`,
      subdomain: 'rm-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const guruEmail = `guru_${Date.now()}@school.test`;
  const guruRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Guru Remedial', email: guruEmail, password: 'Password123!', role: 'GURU', schoolId: SCHOOL_ID });
  guruToken = guruRes.body.data.accessToken;
  userId = guruRes.body.data.user.id;

  const student = await prisma.student.create({
    data: {
      nisn: `RM${Date.now()}`.slice(-10),
      name: 'Remediation Student',
      birthDate: new Date('2008-11-22'),
      gender: 'FEMALE',
      birthPlace: 'Medan',
      religion: 'ISLAM',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.remediationSession.deleteMany({ where: { program: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.remediationProgram.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Remediation Programs (M50)', () => {
  it('POST /api/v1/remediation/programs - should create program', async () => {
    const res = await request(app)
      .post('/api/v1/remediation/programs')
      .set('Authorization', `Bearer ${guruToken}`)
      .send({
        teacherId: userId,
        subject: 'Matematika',
        gradeLevel: '9',
        type: 'REMEDIAL',
        title: 'Remedial Aljabar',
        targetKDs: ['3.1'],
      });
    expect(res.status).toBe(201);
    expect(res.body.data.title).toBe('Remedial Aljabar');
    programId = res.body.data.id;
  });

  it('GET /api/v1/remediation/programs - should list programs', async () => {
    const res = await request(app)
      .get('/api/v1/remediation/programs')
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});

describe('Remediation Sessions', () => {
  it('POST /api/v1/remediation/programs/:id/sessions - should add session', async () => {
    if (!programId) return;
    const res = await request(app)
      .post(`/api/v1/remediation/programs/${programId}/sessions`)
      .set('Authorization', `Bearer ${guruToken}`)
      .send({
        scheduledAt: new Date(Date.now() + 86400000).toISOString(),
        studentIds: [studentId],
      });
    expect(res.status).toBe(201);
    sessionId = res.body.data.id;
  });

  it('GET /api/v1/remediation/programs/:id/sessions - should list sessions', async () => {
    if (!programId) return;
    const res = await request(app)
      .get(`/api/v1/remediation/programs/${programId}/sessions`)
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('PATCH /api/v1/remediation/sessions/:id - should update session (mark attended)', async () => {
    if (!sessionId) return;
    const res = await request(app)
      .patch(`/api/v1/remediation/sessions/${sessionId}`)
      .set('Authorization', `Bearer ${guruToken}`)
      .send({ attendedIds: [studentId] });
    expect(res.status).toBe(200);
  });

  it('GET /api/v1/remediation/students/:studentId - should get student remediations', async () => {
    const res = await request(app)
      .get(`/api/v1/remediation/students/${studentId}`)
      .set('Authorization', `Bearer ${guruToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
