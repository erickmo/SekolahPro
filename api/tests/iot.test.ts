import request from 'supertest';
import { app } from '../src/index';
import { prisma } from '../src/lib/prisma';

const SCHOOL_ID = 'test-school-iot-001';
let token: string;
let deviceId: string;

beforeAll(async () => {
  await prisma.school.upsert({
    where: { subdomain: 'iot-test' },
    create: {
      id: SCHOOL_ID,
      name: 'IoT Test School',
      npsn: `IOT${Date.now()}`,
      subdomain: 'iot-test',
      address: 'Test Address',
      config: {},
    },
    update: {},
  });

  const email = `iot_admin_${Date.now()}@school.test`;
  const regRes = await request(app)
    .post('/api/v1/auth/register')
    .send({ name: 'Admin IoT', email, password: 'Password123!', role: 'ADMIN_SEKOLAH', schoolId: SCHOOL_ID });
  token = regRes.body.data.accessToken;
});

afterAll(async () => {
  await prisma.ioTReading.deleteMany({ where: { device: { schoolId: SCHOOL_ID } } }).catch(() => null);
  await prisma.ioTDevice.deleteMany({ where: { schoolId: SCHOOL_ID } }).catch(() => null);
  await prisma.school.deleteMany({ where: { id: SCHOOL_ID } }).catch(() => null);
  await prisma.$disconnect();
});

describe('IoT Module', () => {
  it('POST /api/v1/iot/devices - should register device', async () => {
    const res = await request(app)
      .post('/api/v1/iot/devices')
      .set('Authorization', `Bearer ${token}`)
      .send({
        name: 'Sensor Suhu R.101',
        type: 'TEMPERATURE_HUMIDITY',
        location: 'Ruang Kelas 101',
        mqttTopic: `school/test/temp/${Date.now()}`,
      });
    expect(res.status).toBe(201);
    expect(res.body.data.name).toBe('Sensor Suhu R.101');
    expect(res.body.data.type).toBe('TEMPERATURE_HUMIDITY');
    deviceId = res.body.data.id;
  });

  it('GET /api/v1/iot/devices - should list devices', async () => {
    const res = await request(app)
      .get('/api/v1/iot/devices')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/iot/devices/:id - should get device by id', async () => {
    if (!deviceId) return;
    const res = await request(app)
      .get(`/api/v1/iot/devices/${deviceId}`)
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body.data.id).toBe(deviceId);
  });

  it('PATCH /api/v1/iot/devices/:id/status - should update device status', async () => {
    if (!deviceId) return;
    const res = await request(app)
      .patch(`/api/v1/iot/devices/${deviceId}/status`)
      .set('Authorization', `Bearer ${token}`)
      .send({ isActive: false });
    expect(res.status).toBe(200);
    expect(res.body.data.isActive).toBe(false);
  });

  it('POST /api/v1/iot/readings - should record reading (no auth)', async () => {
    if (!deviceId) return;
    const res = await request(app)
      .post('/api/v1/iot/readings')
      .send({
        deviceId,
        value: 28.5,
        unit: 'celsius',
        metadata: { humidity: 70 },
      });
    expect(res.status).toBe(201);
    expect(res.body.data.deviceId).toBe(deviceId);
    expect(res.body.data.value).toBe(28.5);
  });

  it('GET /api/v1/iot/readings - should list readings (with auth)', async () => {
    const res = await request(app)
      .get('/api/v1/iot/readings')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(Array.isArray(res.body.data)).toBe(true);
  });

  it('GET /api/v1/iot/devices/:id/latest - should get latest reading', async () => {
    if (!deviceId) return;
    const res = await request(app)
      .get(`/api/v1/iot/devices/${deviceId}/latest`)
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    // Latest reading may or may not be present; both are valid
  });
});
