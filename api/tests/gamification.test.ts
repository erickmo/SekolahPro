import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-gam-001';
let token: string;
let ruleId: string;
let studentId: string;
let badgeId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'gam-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Gamification Test School',
      npsn: `GAM${Date.now()}`,
      subdomain: 'gam-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `gam_admin_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Gamifikasi', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  token = regRes.body.data.accessToken;

  const student = await prisma.student.create({
    data: {
      nisn: `GAM${Date.now()}`.slice(-10),
      name: 'Siswa Gamifikasi',
      birthDate: new Date('2008-11-05'),
      gender: 'FEMALE',
      birthPlace: 'Medan',
      religion: 'KRISTEN',
      address: 'Test Address',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.studentBadge.deleteMany({ where: { studentId } }).catch(() => null);
  await prisma.badge.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.pointTransaction.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.studentPoints.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.pointRule.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Gamification Module', () => {
  it('POST /api/v1/gamification/rules - should create point rule', async () => {
    const res = await request(app)
      .post('/api/v1/gamification/rules')
      .set('Authorization', `Bearer ${token}`)
      .send({
        type: 'ATTENDANCE',
        points: 10,
        description: 'Hadir tepat waktu',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.type).toBe('ATTENDANCE');
    expect(Number(res.body.data.points)).toBe(10);
    ruleId = res.body.data.id;
  });

  it('GET /api/v1/gamification/rules - should list rules', async () => {
    const res = await request(app)
      .get('/api/v1/gamification/rules')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/gamification/points - should award points to student', async () => {
    const res = await request(app)
      .post('/api/v1/gamification/points')
      .set('Authorization', `Bearer ${token}`)
      .send({
        studentId,
        action: 'ATTENDANCE',
        points: 10,
        reason: 'Hadir tepat waktu',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
    expect(res.body.data.totalPoints).toBeGreaterThanOrEqual(10);
  });

  it('GET /api/v1/gamification/points/:studentId - should get student points', async () => {
    const res = await request(app)
      .get(`/api/v1/gamification/points/${studentId}`)
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.data.totalPoints).toBeGreaterThanOrEqual(10);
  });

  it('GET /api/v1/gamification/leaderboard - should get leaderboard', async () => {
    const res = await request(app)
      .get('/api/v1/gamification/leaderboard')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/gamification/badges - should create badge', async () => {
    const res = await request(app)
      .post('/api/v1/gamification/badges')
      .set('Authorization', `Bearer ${token}`)
      .send({
        name: 'Rajin Hadir',
        description: 'Hadir 30 hari berturut-turut',
        criteria: { type: 'ATTENDANCE', count: 30 },
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Rajin Hadir');
    badgeId = res.body.data.id;
  });

  it('GET /api/v1/gamification/badges - should list badges', async () => {
    const res = await request(app)
      .get('/api/v1/gamification/badges')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/gamification/badges/award - should award badge to student', async () => {
    if (!badgeId) return;
    const res = await request(app)
      .post('/api/v1/gamification/badges/award')
      .set('Authorization', `Bearer ${token}`)
      .send({ studentId, badgeId });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
    expect(res.body.data.badgeId).toBe(badgeId);
  });

  it('GET /api/v1/gamification/badges/student/:studentId - should get student badges', async () => {
    const res = await request(app)
      .get(`/api/v1/gamification/badges/student/${studentId}`)
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
