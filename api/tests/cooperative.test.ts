import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-coop-001';
let cashierToken: string;
let studentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'coop-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Coop Test School',
      npsn: `COP${Date.now()}`,
      subdomain: 'coop-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `cashier_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Cashier', email, password: 'Password123!', role: 'KASIR_KOPERASI', schoolId: SCHOOL_ID });
  cashierToken = regRes.body.data.accessToken;

  const student = await prisma.student.create({
    data: {
      nisn: `COP${Date.now()}`.slice(-10),
      name: 'Coop Student',
      birthDate: new Date('2008-01-01'),
      gender: 'FEMALE',
      birthPlace: 'Surabaya',
      religion: 'ISLAM',
      address: 'Test Address',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.savingsTransaction.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.savingsAccount.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.$disconnect();
});

describe('Cooperative Savings', () => {
  it('POST /api/v1/cooperative/:studentId/account - should create savings account', async () => {
    const res = await request(app)
      .post(`/api/v1/cooperative/${studentId}/account`)
      .set('Authorization', `Bearer ${cashierToken}`);
    expect(res.status).toBe(201);
    expect(res.body.data.studentId).toBe(studentId);
    expect(res.body.data.balance).toBe(0);
  });

  it('GET /api/v1/cooperative/:studentId/account - should get account', async () => {
    const res = await request(app)
      .get(`/api/v1/cooperative/${studentId}/account`)
      .set('Authorization', `Bearer ${cashierToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.studentId).toBe(studentId);
  });

  it('POST /api/v1/cooperative/:studentId/deposit - should deposit money', async () => {
    const res = await request(app)
      .post(`/api/v1/cooperative/${studentId}/deposit`)
      .set('Authorization', `Bearer ${cashierToken}`)
      .send({ amount: 50000, note: 'Setoran awal' });
    expect(res.status).toBe(200);
    expect(res.body.data.account.balance).toBe(50000);
  });

  it('POST /api/v1/cooperative/:studentId/withdraw - should withdraw money', async () => {
    const res = await request(app)
      .post(`/api/v1/cooperative/${studentId}/withdraw`)
      .set('Authorization', `Bearer ${cashierToken}`)
      .send({ amount: 10000, note: 'Penarikan' });
    expect(res.status).toBe(200);
    expect(res.body.data.account.balance).toBe(40000);
  });

  it('POST /api/v1/cooperative/:studentId/withdraw - should reject insufficient balance', async () => {
    const res = await request(app)
      .post(`/api/v1/cooperative/${studentId}/withdraw`)
      .set('Authorization', `Bearer ${cashierToken}`)
      .send({ amount: 50000, note: 'Too much' });
    expect(res.status).toBe(400);
  });

  it('GET /api/v1/cooperative/:studentId/transactions - should list transactions', async () => {
    const res = await request(app)
      .get(`/api/v1/cooperative/${studentId}/transactions`)
      .set('Authorization', `Bearer ${cashierToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
    expect(res.body.data.length).toBeGreaterThan(0);
  });
});

describe('Cooperative Products', () => {
  it('POST /api/v1/cooperative/products - should create product', async () => {
    const res = await request(app)
      .post('/api/v1/cooperative/products')
      .set('Authorization', `Bearer ${cashierToken}`)
      .send({ name: 'Pensil HB', price: 2000, stock: 100, category: 'ALAT_TULIS' });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Pensil HB');
  });

  it('GET /api/v1/cooperative/products - should list products', async () => {
    const res = await request(app)
      .get('/api/v1/cooperative/products')
      .set('Authorization', `Bearer ${cashierToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
