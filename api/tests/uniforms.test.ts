import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-uni-001';
let adminToken: string;
let tatausahaToken: string;
let itemId: string;
let orderId: string;
let studentId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'uni-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Uniforms Test School',
      npsn: `UNI${Date.now()}`,
      subdomain: 'uni-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const adminEmail = `uni_admin_${Date.now()}@school.test`;
  const adminRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin Seragam', email: adminEmail, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  adminToken = adminRes.body.data.accessToken;

  const tuEmail = `uni_tu_${Date.now()}@school.test`;
  const tuRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Tata Usaha Seragam', email: tuEmail, password: 'Password123!', role: 'TATA_USAHA', schoolId: SCHOOL_ID });
  tatausahaToken = tuRes.body.data.accessToken;

  const student = await prisma.student.create({
    data: {
      nisn: `UNI${Date.now()}`.slice(-10),
      name: 'Siswa Seragam',
      birthDate: new Date('2009-08-20'),
      gender: 'MALE',
      birthPlace: 'Surabaya',
      religion: 'ISLAM',
      address: 'Test Address',
      schoolId: SCHOOL_ID,
    },
  });
  studentId = student.id;
});

afterAll(async () => {
  await prisma.uniformOrder.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.student.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.uniformItem.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Uniforms Module', () => {
  it('POST /api/v1/uniforms/items - should create uniform item', async () => {
    const res = await request(app)
      .post('/api/v1/uniforms/items')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        name: 'Kemeja Putih Pria',
        description: 'Kemeja seragam',
        price: 85000,
        sizes: ['S', 'M', 'L', 'XL'],
        imageUrl: null,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Kemeja Putih Pria');
    expect(Number(res.body.data.price)).toBe(85000);
    itemId = res.body.data.id;
  });

  it('GET /api/v1/uniforms/items - should list uniform items', async () => {
    const res = await request(app)
      .get('/api/v1/uniforms/items')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('PATCH /api/v1/uniforms/items/:id/stock - should update stock', async () => {
    if (!itemId) return;
    const res = await request(app)
      .patch(`/api/v1/uniforms/items/${itemId}/stock`)
      .set('Authorization', `Bearer ${tatausahaToken}`)
      .send({ stock: { S: 50, M: 100, L: 75, XL: 30 } });
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(itemId);
  });

  it('POST /api/v1/uniforms/orders - should create order', async () => {
    if (!itemId) return;
    const res = await request(app)
      .post('/api/v1/uniforms/orders')
      .set('Authorization', `Bearer ${adminToken}`)
      .send({
        itemId,
        studentId,
        size: 'M',
        quantity: 2,
        totalPrice: 170000,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.itemId).toBe(itemId);
    expect(res.body.data.studentId).toBe(studentId);
    expect(res.body.data.quantity).toBe(2);
    orderId = res.body.data.id;
  });

  it('GET /api/v1/uniforms/orders - should list orders', async () => {
    const res = await request(app)
      .get('/api/v1/uniforms/orders')
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/uniforms/orders/:id - should get order by id', async () => {
    if (!orderId) return;
    const res = await request(app)
      .get(`/api/v1/uniforms/orders/${orderId}`)
      .set('Authorization', `Bearer ${adminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(orderId);
  });

  it('PATCH /api/v1/uniforms/orders/:id/status - should update order status', async () => {
    if (!orderId) return;
    const res = await request(app)
      .patch(`/api/v1/uniforms/orders/${orderId}/status`)
      .set('Authorization', `Bearer ${tatausahaToken}`)
      .send({ status: 'PAID' });
    expect(res.status).toBe(200);
    expect(res.body.data.status).toBe('PAID');
  });
});
