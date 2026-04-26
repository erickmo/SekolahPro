# 02 — Capability Map

**Status**: Draft  
**Perspective**: System Analyst + Architecture

## Shared / Core Capabilities

### Platform Foundation
- clean architecture + CQRS
- Vernon read-cache untuk domain tertentu
- event bus
- dependency injection
- UUID v7
- frontend multi-app foundation

### Cross-Cutting Shared Domains
- multi-tenant 4-level hierarchy
- users & roles
- academic years
- teachers & staff
- class rooms
- dual-mode institution type
- regulatory compliance baseline
- deployment, migration, sync engine strategy

## Management Sekolah Capability Groups

### Student Lifecycle
- student core
- relationships/autoload
- guardian
- address/previous school
- documents
- class placement
- PPDB

### Academic Operations
- curriculum
- subjects
- academic calendar
- teaching schedule
- lesson plan
- teaching journal

### Attendance & Assessment
- daily attendance
- academic record
- subject grade detail
- exams/assessment
- rapor generation

### Student Services
- health
- discipline
- achievement
- extracurricular
- counseling
- SPP

### HR & School Operations
- teacher attendance
- workload
- performance
- development
- leave
- payroll
- substitution

### Facilities & Services
- dormitory
- canteen
- library
- laboratory
- asset & inventory
- room booking
- transportation

### Communication & Administration
- parent portal
- messaging
- notifications
- announcement
- correspondence
- approval workflow
- profile & accreditation
- school committee

### Finance, Integration & Growth
- RKAS
- payment gateway
- e-learning integration
- analytics
- Dapodik
- alumni
- scholarship
- events

## Management Koperasi Capability Groups

### Cooperative Core
- nasabah/member
- rekening
- produk & akad

### Savings & Financing
- simpanan pokok/wajib
- tabungan
- deposito
- pinjaman/pembiayaan
- angsuran
- denda
- agunan

### Operations & Transactions
- transaksi rekening/non-rekening
- teller session
- denomination
- kas & cash flow

### Accounting & Compliance
- jurnal & COA
- SHU
- laporan regulasi
- zakat & infaq
- AML/CFT
- UU PDP
- reserve fund
- health indicators
- takaful integration
- dissolution

## Ownership Model

- `core`: shared foundational domains
- `sekolah`: academic and school operations
- `koperasi`: cooperative finance and operations
- `shared`: actor model, compliance, billing, analytics, migration, governance
