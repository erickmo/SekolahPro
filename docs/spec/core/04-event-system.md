# 04 — Sistem Event SekolahPro

Dokumen ini menjabarkan arsitektur event-driven SekolahPro: event bus abstraction (InMemory untuk development, NATS JetStream untuk production), dependency injection via Uber FX, pola CQRS command/query, dan alur event di dalam sistem.

---

## ADR References

| ADR | Judul | Relevansi |
|-----|-------|-----------|
| ADR-005 | Event Bus Abstraction — InMemory / NATS JetStream | Interface event bus dan dua implementasi |
| ADR-006 | Uber FX sebagai Dependency Injection Container | Lifecycle management dan wiring |
| ADR-001 | Go Clean Architecture + CQRS | Command/query separation |
| ADR-014 | Vernon Sync Engine Strategy | Event-driven sync untuk Vernon _data |

---

## 1. Event Bus Abstraction

### Interface

Semua application code hanya bergantung pada satu interface:

```go
// domain/port/eventbus.go
type Event struct {
    ID          string
    Type        string
    AggregateID string
    OccurredAt  time.Time
    Payload     []byte  // JSON-encoded
}

type Handler func(ctx context.Context, event Event) error

type EventBus interface {
    Publish(ctx context.Context, subject string, event Event) error
    Subscribe(subject string, handler Handler) error
    Close() error
}
```

`domain/` layer mendefinisikan interface ini. `infrastructure/` layer menyediakan implementasi. Application code tidak pernah langsung bergantung pada NATS atau InMemory — hanya pada `EventBus`.

### Implementasi 1: InMemory (Development & Test)

```
Karakteristik:
- In-process, zero network latency
- Events hilang jika process restart (acceptable di dev/test)
- Synchronous publish, async dispatch per goroutine
- Zero configuration — langsung jalan tanpa Docker NATS
```

Aktivasi: `USE_NATS=false` di environment variable.

### Implementasi 2: NATS JetStream (Production)

```
Karakteristik:
- Durable — event tidak hilang jika consumer restart
- At-least-once delivery dengan deduplikasi via nats.MsgId
- Ordered delivery per subject
- Replay capability untuk recovery SyncEngine
- Horizontal scaling consumer
- Retry dengan backoff: NakWithDelay(5 * time.Second)
```

Aktivasi: `USE_NATS=true`, `NATS_URL=nats://nats-1:4222,...`

### JetStream Stream Configuration

| Stream | Subjects | Retensi | Replicas | Penggunaan |
|--------|----------|---------|----------|------------|
| `DOMAIN_EVENTS` | `events.>` | 7 hari | 3 | Semua domain events |
| `SYNC_ENGINE` | `sync.>` | 24 jam | 3 | Vernon SyncEngine events |

---

## 2. Subject Naming Convention

Semua subject mengikuti format:

```
{stream_prefix}.{domain}.{event_type}
```

### Domain Events

```
events.{domain}.{event_type}

Contoh:
events.students.student_created
events.students.student_updated
events.academic_years.academic_year_activated
events.academic_years.academic_year_closed
events.class_rooms.class_room_updated
events.teachers.teacher_updated
events.teachers.teacher_signature_uploaded
```

### Vernon Sync Events

```
sync.{target_domain}.{source_entity}_updated

Contoh:
sync.students.academic_year_updated
sync.class_rooms.teacher_updated
sync.student_grades.class_room_updated
```

**Penting**: Subject harus didefinisikan sebagai Go constants, bukan string literal. Typo di subject name menyebabkan event tidak terdeliver ke consumer yang tepat.

---

## 3. Alur Event dalam Sistem

### Alur 1: Command yang Menghasilkan Domain Event

```
HTTP Request
     │
     ▼
handler/student.go
  (parse request, tidak ada business logic)
     │
     ▼
usecase/command/create_student/handler.go
  ┌─────────────────────────────────────────────┐
  │ 1. Validasi business rule                   │
  │ 2. Generate UUID v7 (app-layer, sebelum DB) │
  │ 3. Build _rels & _data (jika Vernon domain) │
  │ 4. repo.SaveWithVernon(ctx, student)        │
  │ 5. eventBus.Publish(                        │
  │      "events.students.student_created",     │
  │      Event{ID: newUUID, Payload: ...}       │
  │    )                                        │
  └─────────────────────────────────────────────┘
     │
     ▼
202 Accepted (atau 201 dengan ID)
```

### Alur 2: Vernon Sync (Async)

```
Parent entity diupdate (misal: nama guru berubah)
     │
     ▼
eventBus.Publish("events.teachers.teacher_updated", event)
     │
     │  (async — tidak menunggu)
     ▼
NATS JetStream (prod) / InMemory goroutine (dev)
     │
     ▼
SyncEngine Worker picks up event
     │
     ▼
Lookup SyncRegistry["teachers"]
→ [{Table: "class_rooms", Priority: "normal"}, ...]
     │
     ▼
Selective field check: apakah "full_name" ada di Fields?
  JA → lanjut sync
  TIDAK → skip
     │
     ▼
Batch UPDATE class_rooms
SET _data = jsonb_set(_data, '{homeroom_teacher}', $1::jsonb),
    _sync_status = 'synced'
WHERE _rels->>'homeroom_teacher_id' = $2
     │
     ▼
Publish "sync.class_rooms.teacher_updated" (SyncCompleted)
```

### Alur 3: Vernon Sync Failure

```
SyncEngine Worker picks up event
     │
     ▼
Execute batch UPDATE
     │ ERROR (DB timeout, connection issue)
     ▼
Attempt 1: retry setelah 1 detik
Attempt 2: retry setelah 5 detik
Attempt 3: retry setelah 30 detik
     │ MASIH GAGAL
     ▼
Set _sync_status = 'error' pada rows yang gagal
     │
     ▼
Send to Dead Letter Queue (DLQ)
     │
     ▼
Alert monitoring (jika DLQ depth > threshold)
```

---

## 4. CQRS Event Patterns

### Command Handler Pattern

Command handler bertanggung jawab atas:
1. Validasi business rule
2. Operasi write ke repository
3. Publish domain event

```
CreateStudentCommand
  → CreateStudentHandler.Handle()
    → validate()
    → studentRepo.Save()
    → eventBus.Publish("events.students.student_created", ...)
    → return studentID, nil
```

### Query Handler Pattern

Query handler bertanggung jawab atas:
1. Validasi input query (scope, pagination)
2. Fetch dari read repository (atau Vernon reader)
3. Map ke Result DTO

```
ListStudentsQuery
  → ListStudentsHandler.Handle()
    → scope := ScopeFromContext(ctx)
    → reader.ListByTenant(tenantID, filters, pagination)
    → map rows → []StudentListItem
    → return StudentListResult
```

**Penting**: Query handler tidak boleh memanggil Command handler dan sebaliknya.

---

## 5. Uber FX: Dependency Injection

### Module Structure

FX module mengelompokkan dependency per layer/domain:

```go
// Infrastructure modules
var DatabaseModule = fx.Module("database",
    fx.Provide(database.NewPostgres),       // *sqlx.DB
    fx.Provide(database.NewSQLCQueries),    // *sqlc.Queries
    fx.Provide(database.NewMigrator),
)

var MessagingModule = fx.Module("messaging",
    fx.Provide(newEventBus),  // switch InMemory vs NATS via config
)

// Domain module — contoh untuk teachers
var TeacherModule = fx.Module("teacher",
    fx.Provide(
        repository.NewTeacherPostgres,         // write repo
        repository.NewTeacherVernonReader,     // read repo (Vernon)
        command.NewCreateTeacherHandler,
        command.NewUpdateTeacherHandler,
        query.NewGetTeacherHandler,
        query.NewListTeachersHandler,
        handler.NewTeacherHTTPHandler,
        sync.NewTeacherSyncHandler,            // Vernon sync
    ),
)
```

### Application Bootstrap

```go
// cmd/api/main.go
func main() {
    app := fx.New(
        fx.Provide(config.Load),
        fx.Provide(logging.NewZerolog),
        DatabaseModule,
        CacheModule,
        MessagingModule,
        TelemetryModule,

        // Domain modules
        AuthModule,
        UserModule,
        AcademicYearModule,
        ClassRoomModule,
        TeacherModule,
        StudentModule,
        // ... domain lainnya

        // HTTP server
        fx.Provide(router.NewChi),
        fx.Invoke(router.RegisterRoutes),
        fx.Invoke(startHTTPServer),
    )

    app.Run()  // blocks; SIGTERM → graceful shutdown
}
```

### Lifecycle Management (OnStart / OnStop)

FX mengelola urutan startup dan shutdown secara otomatis:

```
Startup Order:
Config → Logger → Database → Redis → NATS → Domain Modules → HTTP Server

Shutdown Order (reverse):
HTTP Server → NATS drain → Redis close → Database close
```

Setiap dependency yang memiliki resource eksternal (DB, NATS, Redis) wajib mendaftarkan `fx.Hook`:

```go
lc.Append(fx.Hook{
    OnStart: func(ctx context.Context) error { /* connect & verify */ },
    OnStop:  func(ctx context.Context) error { /* graceful close */ },
})
```

### Dependency Declaration Pattern

```go
type CreateStudentDeps struct {
    fx.In  // FX marker

    Repo         port.StudentRepository
    TeacherRepo  port.TeacherRepository
    AcademicRepo port.AcademicYearRepository
    EventBus     port.EventBus
    Logger       zerolog.Logger
    Tracer       trace.Tracer
}

func NewCreateStudentHandler(deps CreateStudentDeps) *CreateStudentHandler {
    return &CreateStudentHandler{...}
}
```

### Swappable Implementations (FX Annotate)

```go
// Dua implementasi untuk interface yang sama
fx.Provide(
    fx.Annotate(
        repository.NewStudentPostgres,
        fx.As(new(port.StudentRepository)),   // write
    ),
    fx.Annotate(
        repository.NewStudentVernonReader,
        fx.As(new(port.StudentReader)),        // read (Vernon)
    ),
)
```

---

## 6. Observability

### Tracing (OpenTelemetry)

Setiap command/query handler membuat span:

```go
ctx, span := tracer.Start(ctx, "CreateStudent")
defer span.End()
```

### Metrics (Prometheus)

Setiap service wajib expose `/metrics`:

```
http_request_duration_seconds{path, method, status}
sync_lag_seconds{entity_type}
sync_error_total{entity_type}
event_publish_total{subject}
event_handler_error_total{subject}
```

### Health Endpoints

Setiap service wajib expose:
- `/healthz` — liveness probe (apakah proses berjalan?)
- `/readyz` — readiness probe (apakah siap menerima traffic?)

---

## Key Decisions

1. **EventBus sebagai interface, bukan implementation** — application code tidak pernah bergantung langsung pada NATS atau InMemory

2. **InMemory untuk development** — zero friction: `go run ./cmd/api` langsung jalan tanpa Docker NATS

3. **NATS JetStream untuk production** — at-least-once delivery, persistence, replay, deduplikasi via MsgId

4. **Uber FX untuk lifecycle** — graceful startup/shutdown otomatis; fail-fast jika dependency graph invalid sebelum server menerima request pertama

5. **Domain event di-publish setelah commit** — event hanya di-publish setelah write ke DB berhasil; tidak ada event untuk failed operations

6. **Subject sebagai Go constants** — mencegah typo yang menyebabkan event tidak terdeliver

7. **Sync berjalan async, terpisah dari write transaction** — write performance tidak terpengaruh oleh jumlah dependent domain Vernon

---

## Constraints & Implications

### Constraints

- `domain/` layer mendefinisikan `EventBus` interface — `infrastructure/` layer menyediakan implementasi
- Semua subject event harus didefinisikan sebagai Go constants di `domain/event/subjects.go` atau serupa
- Command handler wajib publish domain event setelah write berhasil
- Query handler tidak boleh mempublikasikan event
- Setiap FX module yang menyediakan resource eksternal wajib mendaftarkan `OnStop` hook
- InMemory dispatch berjalan dalam goroutine tanpa ordering guarantee — test yang bergantung pada ordering harus menggunakan NATS nyata

### Implications untuk Feature Baru

- Setiap domain event baru harus: (1) didefinisikan di `domain/event/`, (2) di-publish di command handler, (3) di-subscribe di consumer yang relevan
- Vernon SyncEngine event wajib menggunakan prefix `sync.` dan didaftarkan di `SyncRegistry`
- Jika feature memerlukan guaranteed event ordering, wajib digunakan NATS JetStream (tidak bisa InMemory)
- Dead Letter Queue events harus di-monitor — jika DLQ menumpuk, ada masalah di SyncEngine yang perlu investigasi

### Behavioral Difference InMemory vs NATS

| Aspek | InMemory | NATS JetStream |
|-------|----------|----------------|
| Ordering | Tidak guaranteed | Guaranteed per subject |
| Durability | Hilang saat restart | Persistent |
| Deduplikasi | Tidak ada | Via MsgId (1 menit window) |
| Retry | Manual (goroutine crash = lost) | Built-in Nak/NakWithDelay |
| Test suitability | Unit + integration test | Behavioral test dengan ordering |
