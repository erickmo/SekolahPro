import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-canteen-001';
let adminToken: string;
let studentId: string;
let menuItemId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'canteen-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Canteen Test School',
      npsn: `CAN${Date.now()}`,
      subdomain: 'canteen-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `admin_can_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Canteen', email, password: 'Password123!', role: 'KASIR_KOPERASI', schoolId: SCHOOL_ID });
  adminToken = regRes.body.data.accessToken;

  // Create a student for ordering
  const student = await prisma.student.create({
    data: {
      nisn: `CAN${Date.now()}`.slice(-10),
      name: 'Canteen Student',
      birthDate: new Date('2008-01-01'),
      gender: 'MALE',
      birthPlace: 'Jakarta',
      religion: 'ISLAM',
      address: 'Test',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.menuOrderItem.deleteMany({ where: { order: { schoolId: SCHOOL_ID } } });
  await prisma.menuOrder.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.cashlessTransaction.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.cashlessWallet.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.menuItem.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } });
  await prisma.$disconnect();
});

describe('Canteen Menu', () => {
  it('POST /api/v1/canteen/menu - should add menu item', async () => {
    const res = await request(app)
      .post('/api/v1/canteen/menu')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        name: 'Nasi Goreng',
        price: 15000,
        category: 'MAKANAN',
        stock: 50,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Nasi Goreng');
    menuItemId = res.body.data.id;
  });

  it('GET /api/v1/canteen/menu - should list menu items', async () => {
    const res = await request(app)
      .get('/api/v1/canteen/menu')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
    expect(res.body.data.length).toBeGreaterThan(0);
  });
});

describe('Canteen Wallet', () => {
  it('POST /api/v1/canteen/wallet/topup - should top up wallet', async () => {
    const res = await request(app)
      .post('/api/v1/canteen/wallet/topup')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({ ownerId: studentId, amount: 100000 });
    expect(res.status).toBe(200);
    expect(res.body.data.balance).toBeGreaterThanOrEqual(100000);
  });

  it('GET /api/v1/canteen/wallet/:ownerId - should get wallet balance', async () => {
    const res = await request(app)
      .get(`/api/v1/canteen/wallet/${studentId}`)
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data).toHaveProperty('balance');
  });
});

describe('Canteen Orders', () => {
  it('POST /api/v1/canteen/orders - should place order with wallet', async () => {
    if (!menuItemId) return;
    const res = await request(app)
      .post('/api/v1/canteen/orders')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        studentId,
        items: [{ menuItemId, quantity: 1 }],
        paymentMethod: 'WALLET',
      });
    expect(res.status).toBe(201);
    expect(res.body.data).toHaveProperty('id');
    expect(res.body.data.status).toBe('PENDING');
  });

  it('GET /api/v1/canteen/orders - should list orders', async () => {
    const res = await request(app)
      .get('/api/v1/canteen/orders')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
