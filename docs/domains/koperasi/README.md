# Koperasi Domain

Cooperative financial management platform. Organized into subdomains:

| Subdomain | Description | ADRs |
|-----------|-------------|------|
| [member](member/) | Nasabah, rekening, membership lifecycle | ADR-K001–K002, K030 |
| [produk](produk/) | Produk akad, simpanan pokok/wajib, tabungan, deposito | ADR-K003–K006 |
| [pinjaman](pinjaman/) | Pinjaman, angsuran, denda, jaminan, loan collection | ADR-K007–K010, K033 |
| [transaksi](transaksi/) | Transaksi, teller session, money denomination | ADR-K011–K013 |
| [keuangan](keuangan/) | Kas/cashflow, jurnal COA, SHU, reserve fund | ADR-K014–K016, K034 |
| [kepatuhan](kepatuhan/) | Laporan regulasi, AML/CFT, data privacy, BCP | ADR-K017, K027–K029 |
| [zakat](zakat/) | Zakat dan infaq | ADR-K018 |
| [toko](toko/) | Toko dan kantin | ADR-K019 |
| [hr](hr/) | Payroll deduction | ADR-K020 |
| [digital](digital/) | E-wallet, biometric, mobile app | ADR-K021, K032, K036 |
| [komunikasi](komunikasi/) | Notifikasi, dashboard portal | ADR-K022–K023 |
| [integrasi](integrasi/) | API integration | ADR-K024 |
| [governance](governance/) | Governance, RAT, health indicators, dissolution | ADR-K025–K026, K031, K037–K040 |
| [asuransi](asuransi/) | Insurance/takaful integration | ADR-K035 |

## Cross-Domain Events

See [docs/shared/events/](../../shared/events/) for canonical event schemas.
