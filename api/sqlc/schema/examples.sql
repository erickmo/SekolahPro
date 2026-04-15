-- Schema untuk sqlc code generation.
-- tenant_id dan company_id wajib ada di semua tabel — digunakan pada mode single maupun multi tenant.
CREATE TABLE examples (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID         NOT NULL,
    company_id  UUID         NOT NULL,
    branch_id   UUID,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
