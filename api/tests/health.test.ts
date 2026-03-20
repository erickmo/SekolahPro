import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-health-001';
let staffToken: string;
let studentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'health-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Health Test School',
      npsn: `HLT${Date.now()}`,
      subdomain: 'health-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `uks_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'UKS Staff', email, password: 'Password123!', role: 'PETUGAS_UKS', schoolId: SCHOOL_ID });
  staffToken = regRes.body.data.accessToken;

  const student = await prisma.student.create({
    data: {
      nisn: `HLT${Date.now()}`.slice(-10),
      name: 'Health Student',
      birthDate: new Date('2008-06-15'),
      gender: 'FEMALE',
      birthPlace: 'Bandung',
      religion: 'ISLAM',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.nutritionRecord.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.uKSVisit.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.healthRecord.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.$disconnect();
});

describe('Health Records (UKS)', () => {
  it('POST /api/v1/health/records - should create health record', async () => {
    const res = await request(app)
      .post('/api/v1/health/records')
      .set('Authorization', `Bearer ${staffToken}`)
      .send({
        studentId,
        bloodType: 'A',
        allergies: 'debu',
        medicalConditions: 'Tidak ada',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
  });

  it('GET /api/v1/health/records/:studentId - should get health record', async () => {
    const res = await request(app)
      .get(`/api/v1/health/records/${studentId}`)
      .set('Authorization', `Bearer ${staffToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.bloodType).toBe('A');
  });

  it('POST /api/v1/health/uks-visits - should record UKS visit', async () => {
    const res = await request(app)
      .post('/api/v1/health/uks-visits')
      .set('Authorization', `Bearer ${staffToken}`)
      .send({
        studentId,
        complaint: 'Sakit kepala',
        action: 'Diberi obat dan istirahat',
      });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
  });

  it('GET /api/v1/health/uks-visits - should list UKS visits', async () => {
    const res = await request(app)
      .get('/api/v1/health/uks-visits')
      .set('Authorization', `Bearer ${staffToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/health/nutrition - should record nutrition', async () => {
    const res = await request(app)
      .post('/api/v1/health/nutrition')
      .set('Authorization', `Bearer ${staffToken}`)
      .send({
        studentId,
        height: 155,
        weight: 48,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
  });
});
