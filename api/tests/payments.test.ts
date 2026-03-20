import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-pay-001';
let bendaharaToken: string;
let studentId: string;
let semesterId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'pay-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Payment Test School',
      npsn: `PAY${Date.now()}`,
      subdomain: 'pay-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `bendahara_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Bendahara', email, password: 'Password123!', role: 'BENDAHARA', schoolId: SCHOOL_ID });
  bendaharaToken = regRes.body.data.accessToken;

  // Create academic year and semester for testing
  const ay = await prisma.academicYear.create({
    data: { name: '2024/2025', startDate: new Date('2024-07-15'), endDate: new Date('2025-06-30'), isCurrent: true, schoolId: SCHOOL_ID },
  });
  const sem = await prisma.semester.create({
    data: { name: 'Semester 1', academicYearId: ay.id, startDate: new Date('2024-07-15'), endDate: new Date('2024-12-20') },
  });
  semesterId = sem.id;

  const student = await prisma.student.create({
    data: {
      nisn: `PAY${Date.now()}`.slice(-10),
      name: 'Pay Student',
      birthDate: new Date('2008-08-17'),
      gender: 'MALE',
      birthPlace: 'Malang',
      religion: 'ISLAM',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.invoice.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.semester.deleteMany({ where: { academicYear: { schoolId: SCHOOL_ID } } });
  await prisma.academicYear.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.$disconnect();
});

describe('Payments (SPP)', () => {
  it('POST /api/v1/payments/invoices/generate - should generate invoices', async () => {
    const res = await request(app)
      .post('/api/v1/payments/invoices/generate')
      .set('Authorization', `Bearer ${bendaharaToken}`)
      .send({
        semesterId,
        amount: 500000,
        dueDate: new Date(Date.now() + 30 * 86400000).toISOString(),
      });
    expect(res.status).toBe(201);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/payments/invoices - should list invoices', async () => {
    const res = await request(app)
      .get('/api/v1/payments/invoices')
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
    expect(res.body.meta).toHaveProperty('total');
  });

  it('GET /api/v1/payments/financial-summary - should get financial summary', async () => {
    const res = await request(app)
      .get('/api/v1/payments/financial-summary')
      .set('Authorization', `Bearer ${bendaharaToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data).toHaveProperty('totalBilled');
    expect(res.body.data).toHaveProperty('totalPaid');
  });
});
