-- Seed data untuk domain Example.
-- Gunakan UUID yang tetap agar seed idempotent (bisa dijalankan berulang).
-- Ganti TENANT_ID dan COMPANY_ID dengan nilai yang sesuai di .env sebelum menjalankan.

-- Jalankan dengan:
--   psql $DATABASE_URL -f seeds/001_example.sql

DO $$
DECLARE
  v_tenant_id  UUID := '00000000-0000-0000-0000-000000000001'::UUID;
  v_company_id UUID := '00000000-0000-0000-0000-000000000002'::UUID;
BEGIN
  INSERT INTO examples (id, tenant_id, company_id, name, description, is_active, created_at, updated_at)
  VALUES
    ('10000000-0000-0000-0000-000000000001', v_tenant_id, v_company_id,
     'Example Alpha', 'Contoh pertama untuk development dan testing', TRUE, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000002', v_tenant_id, v_company_id,
     'Example Beta',  'Contoh kedua dengan is_active=false',        FALSE, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000003', v_tenant_id, v_company_id,
     'Example Gamma', 'Contoh ketiga untuk pagination testing',     TRUE, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000004', v_tenant_id, v_company_id,
     'Example Delta', 'Contoh keempat — akan diupdate saat test',   TRUE, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000005', v_tenant_id, v_company_id,
     'Example Epsilon', 'Contoh kelima — akan dihapus saat test',   TRUE, NOW(), NOW())
  ON CONFLICT (id) DO NOTHING;

  RAISE NOTICE 'Seed 001_example: % rows inserted', (SELECT COUNT(*) FROM examples WHERE tenant_id = v_tenant_id);
END $$;
