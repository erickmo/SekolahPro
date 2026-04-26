# Sekolah Domain

School management platform. Organized into subdomains:

| Subdomain | Description | ADRs |
|-----------|-------------|------|
| [student](student/) | Student data, admission, finance, health, discipline, counseling | ADR-S001–S018 |
| [academic](academic/) | Curriculum, subjects, teaching schedules, exams, rapor | ADR-S019–S025 |
| [teacher](teacher/) | Teacher attendance, workload, evaluation, leave, payroll | ADR-S026–S032 |
| [dormitory](dormitory/) | Dormitory management, activities, discipline | ADR-S033–S035 |
| [canteen](canteen/) | Canteen management and transactions | ADR-S036–S037 |
| [facility](facility/) | Library, laboratory, asset inventory, room booking | ADR-S038–S041 |
| [communication](communication/) | Parent portal, messaging, notifications, announcements | ADR-S042–S046 |
| [admin](admin/) | Approval workflows, school profile, committee, budget | ADR-S047–S050 |
| [finance](finance/) | Payment gateway | ADR-S051 |
| [integration](integration/) | Transportation, e-learning, reporting, Dapodik | ADR-S052–S055 |
| [alumni](alumni/) | Alumni, scholarships, school events | ADR-S056–S058 |

## Cross-Domain Events

See [docs/shared/events/](../../shared/events/) for canonical event schemas.
