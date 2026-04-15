-- +migrate Up
-- Vernon base: UUID v7 generator (time-sortable) + sync queue untuk at-least-once delivery.
-- Dipakai oleh semua tabel domain Vernon (_rels/_data pattern).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- uuid_generate_v7: UUID v7 — time-sortable, monotonically increasing.
-- Lebih baik dari gen_random_uuid() (v4) untuk primary key karena clustering index B-tree.
CREATE OR REPLACE FUNCTION uuid_generate_v7() RETURNS uuid AS $$
DECLARE
  unix_ts_ms bigint;
  uuid_bytes bytea;
BEGIN
  unix_ts_ms := (EXTRACT(EPOCH FROM clock_timestamp()) * 1000)::bigint;
  uuid_bytes := set_byte(
    set_byte(
      overlay(gen_random_bytes(16) placing substring(int8send(unix_ts_ms) from 3) from 1 for 6),
      6, (get_byte(gen_random_bytes(16), 6) & 15) | 112
    ),
    8, (get_byte(gen_random_bytes(16), 8) & 63) | 128
  );
  RETURN encode(uuid_bytes, 'hex')::uuid;
END
$$ LANGUAGE plpgsql VOLATILE;

-- _sync_queue: Transactional outbox untuk Vernon sync events.
-- SyncEngine menulis ke sini di dalam TX yang sama dengan domain write.
-- Background worker membaca dan propagate ke consumer domains.
-- Memberikan at-least-once delivery tanpa bergantung pada pg_notify.
CREATE TABLE IF NOT EXISTS _sync_queue (
    id           BIGSERIAL    PRIMARY KEY,
    domain       TEXT         NOT NULL,
    entity_id    UUID         NOT NULL,
    tenant_id    UUID         NOT NULL,
    company_id   UUID         NOT NULL,
    action       TEXT         NOT NULL,  -- create | update | patch | delete
    payload      JSONB        NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    failed_at    TIMESTAMPTZ,
    retry_count  INT          NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_sync_queue_pending
    ON _sync_queue (created_at ASC)
    WHERE processed_at IS NULL AND failed_at IS NULL;

-- Template untuk domain baru yang pakai Vernon pattern.
-- Salin dan ganti {domain_name} dengan nama tabel.
--
-- CREATE TABLE {domain_name} (
--     id              UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
--     tenant_id       UUID        NOT NULL,
--     company_id      UUID        NOT NULL,
--     _rels           JSONB       NOT NULL DEFAULT '{}',
--     _data           JSONB       NOT NULL DEFAULT '{}',
--     _sync_status    TEXT        NOT NULL DEFAULT 'synced',
--     _sync_version   BIGINT      NOT NULL DEFAULT 0,
--     created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
--     updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
--     deleted_at      TIMESTAMPTZ
-- );
--
-- CREATE INDEX idx_{domain_name}_scope
--     ON {domain_name} (tenant_id, company_id, created_at DESC);
-- CREATE INDEX idx_{domain_name}_data
--     ON {domain_name} USING GIN (_data) WHERE deleted_at IS NULL;
-- CREATE INDEX idx_{domain_name}_sync
--     ON {domain_name} (_sync_status)
--     WHERE _sync_status != 'synced' AND deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS _sync_queue;
DROP FUNCTION IF EXISTS uuid_generate_v7();
