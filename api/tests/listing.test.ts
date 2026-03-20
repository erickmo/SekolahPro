import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-lst-001';
let userToken: string;
let superadminToken: string;
let vendorId: string;
let placementId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'lst-test' },
    create: {
      id: SCHOOL_ID,
      name: 'Listing Test School',
      npsn: `LST${Date.now()}`,
      subdomain: 'lst-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  // Regular authenticated user for vendor registration
  const userEmail = `user_lst_${Date.now()}@school.test`;
  const userRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'User Listing', email: userEmail, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  userToken = userRes.body.data.accessToken;

  // EDS_SUPERADMIN for approval and placement creation
  const adminEmail = `superadmin_lst_${Date.now()}@school.test`;
  const adminRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'EDS Superadmin Listing', email: adminEmail, password: 'Password123!', role: 'EDS_SUPERADMIN', schoolId: SCHOOL_ID });
  superadminToken = adminRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.listingReview.deleteMany({ where: { vendorId } }).catch(() => null);
  if (placementId) {
    await prisma.listingPlacement.deleteMany({ where: { id: placementId } }).catch(() => null);
  }
  await prisma.listingBilling.deleteMany({ where: { vendorId } }).catch(() => null);
  if (vendorId) {
    await prisma.listingVendor.deleteMany({ where: { id: vendorId } }).catch(() => null);
  }
  await prisma.user.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('Listing Marketplace', () => {
  it('GET /api/v1/listing/active — list active placements (public, no auth)', async () => {
    const res = await request(app).get('/api/v1/listing/active');
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('POST /api/v1/listing/vendors — register vendor as authenticated user', async () => {
    const res = await request(app)
      .post('/api/v1/listing/vendors')
      .set('Authorization', `Bearer ${userToken}`)
      .send({
        businessName: 'Bimbel Pintar',
        category: 'GURU_LES',
        description: 'Bimbel terpercaya',
        contactPhone: '08123456789',
        contactEmail: 'bimbel@test.com',
        photos: [],
        coverageArea: ['Jakarta Selatan'],
      });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('id');
    expect(res.body.data.businessName).toBe('Bimbel Pintar');
    vendorId = res.body.data.id;
  });

  it('GET /api/v1/listing/vendors — list vendors', async () => {
    const res = await request(app)
      .get('/api/v1/listing/vendors')
      .set('Authorization', `Bearer ${userToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/listing/vendors/:id — get vendor detail', async () => {
    if (!vendorId) return;
    const res = await request(app)
      .get(`/api/v1/listing/vendors/${vendorId}`)
      .set('Authorization', `Bearer ${userToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.id).toBe(vendorId);
  });

  it('PATCH /api/v1/listing/vendors/:id/approve — approve vendor as EDS_SUPERADMIN', async () => {
    if (!vendorId) return;
    const res = await request(app)
      .patch(`/api/v1/listing/vendors/${vendorId}/approve`)
      .set('Authorization', `Bearer ${superadminToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.status).toBe('ACTIVE');
  });

  it('POST /api/v1/listing/placements — create placement as EDS_SUPERADMIN', async () => {
    if (!vendorId) return;
    const res = await request(app)
      .post('/api/v1/listing/placements')
      .set('Authorization', `Bearer ${superadminToken}`)
      .send({
        vendorId,
        moduleId: 'M11',
        position: 1,
      });
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveProperty('id');
    expect(res.body.data.vendorId).toBe(vendorId);
    placementId = res.body.data.id;
  });

  it('GET /api/v1/listing/placements — list placements', async () => {
    const res = await request(app)
      .get('/api/v1/listing/placements')
      .set('Authorization', `Bearer ${userToken}`);
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
