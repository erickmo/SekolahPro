# Developer Documentation

Go + React Boilerplate — CQRS + Vernon Hybrid Architecture

## Index

| # | File | Description |
|---|------|-------------|
| 01 | [Getting Started](./01-getting-started.md) | Prerequisites, clone, setup, run API dan web dashboard |
| 02 | [Architecture Overview](./02-architecture-overview.md) | Diagram arsitektur, layer diagram, decision matrix CQRS vs Vernon |
| 03 | [Adding a CQRS Domain](./03-adding-cqrs-domain.md) | Step-by-step menambah domain dengan pola CQRS |
| 04 | [Adding a Vernon Domain](./04-adding-vernon-domain.md) | Step-by-step menambah domain dengan pola Vernon _rels/_data |
| 05 | [Multi-Tenant Guide](./05-multi-tenant-guide.md) | Single mode, multi mode, Two-Phase JWT, middleware chain |
| 06 | [Testing Guide](./06-testing-guide.md) | Unit tests, integration tests Testcontainers, Vitest, Playwright |
| 07 | [Environment Configuration](./07-environment-configuration.md) | Tabel env vars, .env.example walkthrough, Docker services |

## Quick Reference

```
Boilerplate root
├── api/                 Go API (Chi + Uber FX + CQRS + Vernon)
│   ├── cmd/api/         Entry point
│   ├── internal/        Domain, command, query, delivery
│   ├── infrastructure/  Config, DB, cache, telemetry
│   ├── pkg/             Shared packages
│   ├── migrations/      SQL migrations
│   └── tests/           Integration tests
└── web-dashboard/       React 18 + TypeScript + Vite
    └── src/             Aplikasi frontend
```

## Prinsip Utama

- **Hybrid pattern**: CQRS untuk domain dengan business logic kompleks, Vernon untuk domain read-heavy dengan banyak relasi
- **Multi-tenant by design**: semua tabel punya `tenant_id` + `company_id` sebagai leading column index
- **Dependency injection**: semua dependency lewat Uber FX, zero manual wiring
- **Scope sebagai parameter eksplisit**: repository tidak pernah baca scope dari context — selalu dari parameter

## Docs ini dibaca oleh

- Developer baru yang onboarding ke project
- Developer senior yang menambah domain atau fitur baru
- Reviewer yang ingin memahami arsitektur sebelum review PR
