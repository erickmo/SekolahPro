-- +migrate Up
-- Domain CQRS: users, roles, user_roles (ADR-013)
-- Authentication & Authorization — bukan Vernon, pakai typed columns.

-- ── Users: akun login ────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS users (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- Identitas login
    email           VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    password_hash   TEXT         NOT NULL,
    full_name       VARCHAR(255) NOT NULL,

    -- Status
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    is_superadmin   BOOLEAN      NOT NULL DEFAULT false,
    email_verified  BOOLEAN      NOT NULL DEFAULT false,
    last_login_at   TIMESTAMPTZ,

    -- Timestamps
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_user_email UNIQUE (email)
);

CREATE INDEX IF NOT EXISTS idx_user_email
    ON users (email);
CREATE INDEX IF NOT EXISTS idx_user_active
    ON users (is_active)
    WHERE is_active = true;

-- ── Roles: definisi role per company ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS roles (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID         NOT NULL,
    company_id      UUID         NOT NULL,

    -- Identitas
    name            VARCHAR(50)  NOT NULL,
    code            VARCHAR(30)  NOT NULL,
    description     TEXT,

    -- Role type
    role_type       VARCHAR(20)  NOT NULL,
    is_system       BOOLEAN      NOT NULL DEFAULT false,

    -- Permissions (JSONB array)
    permissions     JSONB        NOT NULL DEFAULT '[]',

    -- Timestamps
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_role_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_role_type CHECK (role_type IN (
        'admin', 'principal', 'vice_principal',
        'teacher', 'homeroom_teacher', 'counselor',
        'finance', 'staff', 'parent'
    ))
);

CREATE INDEX IF NOT EXISTS idx_role_tenant_company
    ON roles (tenant_id, company_id);

-- ── User Roles: assignment user ↔ role per company ──────────────────────────

CREATE TABLE IF NOT EXISTS user_roles (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id         UUID         NOT NULL,
    role_id         UUID         NOT NULL,
    tenant_id       UUID         NOT NULL,
    company_id      UUID         NOT NULL,

    -- Optional: link ke teacher/guardian
    teacher_id      UUID,
    guardian_id     UUID,

    -- Status
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    assigned_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_user_role UNIQUE (user_id, role_id, company_id)
);

CREATE INDEX IF NOT EXISTS idx_user_role_user
    ON user_roles (user_id);
CREATE INDEX IF NOT EXISTS idx_user_role_role
    ON user_roles (role_id);
CREATE INDEX IF NOT EXISTS idx_user_role_company
    ON user_roles (tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_user_role_teacher
    ON user_roles (teacher_id)
    WHERE teacher_id IS NOT NULL;

-- +migrate Down
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
