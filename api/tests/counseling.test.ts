import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-bk-001';
let bkToken: string;
let studentId: string;
let sessionId: string;
let reportId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'bk-test' },
    create: {
      id: SCHOOL_ID,
      name: 'BK Test School',
      npsn: `BKS${Date.now()}`,
      subdomain: 'bk-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `bk_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Guru BK', email, password: 'Password123!', role: 'GURU_BK', schoolId: SCHOOL_ID });
  bkToken = regRes.body.data.accessToken;
  const counselorId = regRes.body.data.user.id;

  const student = await prisma.student.create({
    data: {
      nisn: `BK${Date.now()}`.slice(-10),
      name: 'BK Student',
      birthDate: new Date('2007-03-20'),
      gender: 'MALE',
      birthPlace: 'Yogyakarta',
      religion: 'ISLAM',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.counselingSession.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.bullyingReport.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.$disconnect();
});

describe('Counseling Sessions (BK)', () => {
  it('POST /api/v1/counseling/sessions - should book session', async () => {
    const res = await request(app)
      .post('/api/v1/counseling/sessions')
      .set('Authorization', `Bearer ${bkToken}`)
      .send({
        studentId,
        counselorId: (await prisma.user.findFirst({ where: { schoolId: SCHOOL_ID } }))?.id || 'test-counselor',
        scheduledAt: new Date(Date.now() + 86400000).toISOString(),
        topic: 'Konsultasi akademik',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
    sessionId = res.body.data.id;
  });

  it('GET /api/v1/counseling/sessions - should list sessions', async () => {
    const res = await request(app)
      .get('/api/v1/counseling/sessions')
      .set('Authorization', `Bearer ${bkToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('PUT /api/v1/counseling/sessions/:id - should update session notes', async () => {
    if (!sessionId) return;
    const res = await request(app)
      .put(`/api/v1/counseling/sessions/${sessionId}`)
      .set('Authorization', `Bearer ${bkToken}`)
      .send({
        notes: 'Siswa menunjukkan perkembangan positif',
        status: 'COMPLETED',
        outcome: 'Siswa berkomitmen untuk meningkatkan prestasi',
      });
    expect(res.status).toBe(200);
    expect(res.body.data.status).toBe('COMPLETED');
  });
});

describe('Anti-Bullying Reports', () => {
  it('POST /api/v1/counseling/bullying-reports - should report bullying', async () => {
    const res = await request(app)
      .post('/api/v1/counseling/bullying-reports')
      .set('Authorization', `Bearer ${bkToken}`)
      .send({
        victimId: studentId,
        perpetratorDesc: 'Siswa kelas XI tidak dikenal',
        type: 'VERBAL',
        description: 'Diejek dan dihina di koridor sekolah',
        location: 'Koridor lantai 2',
        isAnonymous: false,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.type).toBe('VERBAL');
    reportId = res.body.data.id;
  });

  it('GET /api/v1/counseling/bullying-reports - should list reports', async () => {
    const res = await request(app)
      .get('/api/v1/counseling/bullying-reports')
      .set('Authorization', `Bearer ${bkToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('PUT /api/v1/counseling/bullying-reports/:id - should update report status', async () => {
    if (!reportId) return;
    const res = await request(app)
      .put(`/api/v1/counseling/bullying-reports/${reportId}`)
      .set('Authorization', `Bearer ${bkToken}`)
      .send({
        status: 'INVESTIGATING',
        actionTaken: 'Memanggil pihak-pihak yang terlibat untuk mediasi',
      });
    expect(res.status).toBe(200);
  });
});
