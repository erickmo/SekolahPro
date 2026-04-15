# 04 — API Usage Guide

## Base URL dan Format Dasar

Semua request API menggunakan format:

```
https://api.domain-anda.com
```

- **Format request body:** `application/json`
- **Format response:** `application/json`
- **Autentikasi:** JWT Bearer Token (di header `Authorization`)

---

## Endpoint Overview

| Kategori | Method | Path | Deskripsi |
|----------|--------|------|-----------|
| System | GET | `/health` | Status API (tidak butuh token) |
| System | GET | `/metrics` | Prometheus metrics (tidak butuh token) |
| Docs | GET | `/swagger/*` | Dokumentasi API interaktif |
| Examples | GET, POST | `/api/v1/examples` | CRUD resource contoh (butuh token) |
| Examples | GET, PUT, PATCH, DELETE | `/api/v1/examples/{id}` | Operasi per item (butuh token) |
| Vernon Domain | GET, POST | `/api/v1/{domain}/` | CRUD domain Vernon auto-generated |
| Vernon Domain | GET, PUT, PATCH, DELETE | `/api/v1/{domain}/{id}` | Operasi per item domain Vernon |

---

## Authentication Flow

### Single-Tenant — Alur Login

```
┌─────────────┐          ┌──────────────────┐
│   Client    │          │   API Server     │
└──────┬──────┘          └────────┬─────────┘
       │                          │
       │  POST /api/v1/auth/login │
       │  { username, password }  │
       │ ─────────────────────── ▶│
       │                          │
       │  200 OK                  │
       │  { access_token,         │
       │    refresh_token }       │
       │ ◀─────────────────────── │
       │                          │
       │  GET /api/v1/examples    │
       │  Authorization: Bearer   │
       │  {access_token}          │
       │ ─────────────────────── ▶│
       │                          │
       │  200 OK { data: [...] }  │
       │ ◀─────────────────────── │
```

### Multi-Tenant — Alur Login Dua Fase

```
┌─────────────┐          ┌──────────────────┐
│   Client    │          │   API Server     │
└──────┬──────┘          └────────┬─────────┘
       │                          │
       │  FASE 1: Login           │
       │  POST /api/v1/auth/login │
       │  { username, password }  │
       │ ─────────────────────── ▶│
       │                          │
       │  200 OK                  │
       │  { access_token }        │
       │  (hanya berisi tenant_id)│
       │ ◀─────────────────────── │
       │                          │
       │  GET /api/v1/companies   │
       │  Authorization: Bearer   │
       │  {access_token_fase_1}   │
       │ ─────────────────────── ▶│
       │                          │
       │  200 OK { companies: [   │
       │    { id, code, name }    │
       │  ]}                      │
       │ ◀─────────────────────── │
       │                          │
       │  FASE 2: Pilih Company   │
       │  POST /api/v1/auth/scope │
       │  { company_id: "..." }   │
       │ ─────────────────────── ▶│
       │                          │
       │  200 OK                  │
       │  { access_token }        │
       │  (berisi tenant_id +     │
       │   company_id + branch_id)│
       │ ◀─────────────────────── │
       │                          │
       │  GET /api/v1/orders/     │
       │  Authorization: Bearer   │
       │  {access_token_fase_2}   │
       │ ─────────────────────── ▶│
       │                          │
       │  200 OK { data: [...] }  │
       │ ◀─────────────────────── │
```

---

## Cara Mendapatkan JWT Token

### Login (Single-Tenant dan Multi-Tenant)

```bash
curl -X POST https://api.domain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user@company.com",
    "password": "password123"
  }'
```

**Response berhasil:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

### Pilih Company (Multi-Tenant Fase 2)

```bash
curl -X POST https://api.domain.com/api/v1/auth/scope \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {access_token_fase_1}" \
  -d '{
    "company_id": "550e8400-e29b-41d4-a716-446655440002"
  }'
```

**Response berhasil:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

### Refresh Token

```bash
curl -X POST https://api.domain.com/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Menyimpan Token di Header

Untuk semua request yang membutuhkan autentikasi, sertakan token di header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Request dan Response Format

### Request Body (untuk POST/PUT/PATCH)

```json
{
  "field1": "nilai",
  "field2": 123,
  "field3": true
}
```

### Vernon Domain — Format Body

Untuk domain yang menggunakan Vernon Pattern, data dimasukkan ke dalam objek `_data`:

```json
{
  "_data": {
    "name": "Nama Produk",
    "price": 150000,
    "stock": 50,
    "category": "electronics"
  },
  "_rels": {
    "category_id": "uuid-kategori"
  }
}
```

### Response Format Standar

**Satu item:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "name": "Contoh Item",
  "created_at": "2026-04-14T08:00:00Z",
  "updated_at": "2026-04-14T08:00:00Z"
}
```

**Koleksi dengan pagination:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Item Pertama"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "Item Kedua"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

## Pagination

Semua endpoint yang mengembalikan koleksi mendukung pagination melalui query parameter:

| Parameter | Default | Deskripsi |
|-----------|---------|-----------|
| `page` | `1` | Nomor halaman (dimulai dari 1) |
| `limit` | `20` | Jumlah item per halaman. Maksimum: `100` |

### Contoh

```bash
# Halaman pertama, 20 item
GET /api/v1/products/?page=1&limit=20

# Halaman ketiga, 10 item
GET /api/v1/products/?page=3&limit=10
```

### Membaca Response Pagination

```json
{
  "data": [ ... ],
  "meta": {
    "page": 2,
    "limit": 10,
    "total": 150,
    "total_pages": 15
  }
}
```

- `total` — total semua item (sebelum pagination)
- `total_pages` — total halaman yang tersedia
- Jika `page` lebih besar dari `total_pages`, `data` akan berisi array kosong `[]`

---

## Filter dan Sort (Vernon Domains)

Endpoint domain Vernon mendukung filter dan sort melalui query parameter.

### Filter

**Format:** `filter[nama_field]=nilai`

Filter bekerja pada field yang ada di dalam kolom `_data` (JSON).

```bash
# Filter produk berdasarkan kategori
GET /api/v1/products/?filter[category]=electronics

# Filter berdasarkan status
GET /api/v1/orders/?filter[status]=pending

# Kombinasi filter (AND logic)
GET /api/v1/products/?filter[category]=electronics&filter[status]=active
```

> **Catatan:** Filter menggunakan pencocokan exact (sama persis). Untuk pencarian
> yang lebih kompleks, gunakan endpoint search khusus jika tersedia.

### Sort

**Format:** `sort=nama_field` (ascending) atau `sort=-nama_field` (descending dengan prefix `-`)

```bash
# Sort ascending berdasarkan nama
GET /api/v1/products/?sort=name

# Sort descending berdasarkan tanggal dibuat (terbaru duluan)
GET /api/v1/products/?sort=-created_at

# Sort berdasarkan field dalam _data
GET /api/v1/products/?sort=-price
```

### Kombinasi Filter, Sort, dan Pagination

```bash
GET /api/v1/products/?filter[category]=electronics&sort=-price&page=1&limit=10
```

---

## Relationship Expansion (expand)

Vernon domains mendukung autoloading relasi menggunakan parameter `expand`.

**Format:** `expand=nama_relasi` atau `expand=relasi1,relasi2` (multiple)

Tanpa expand (relasi berupa ID saja):
```json
{
  "id": "uuid-produk",
  "_data": {
    "name": "Laptop Gaming",
    "price": 15000000
  },
  "_rels": {
    "category_id": "uuid-kategori"
  }
}
```

Dengan `expand=category`:
```json
{
  "id": "uuid-produk",
  "_data": {
    "name": "Laptop Gaming",
    "price": 15000000
  },
  "_rels": {
    "category_id": "uuid-kategori",
    "category": {
      "id": "uuid-kategori",
      "_data": {
        "name": "Electronics",
        "slug": "electronics"
      }
    }
  }
}
```

### Expand Berantai (Dot Notation)

```bash
# Expand relasi category, dan dari category expand relasi parent_category
GET /api/v1/products/?expand=category.parent_category

# Multiple expand
GET /api/v1/orders/?expand=customer,product,branch
```

---

## Error Response Format

Semua error dikembalikan dalam format JSON yang konsisten:

```json
{
  "error": "pesan error yang mudah dibaca",
  "code": "ERROR_CODE_CONSTANT",
  "details": {
    "field": "deskripsi error per field (untuk validation error)"
  }
}
```

### HTTP Status Codes

| Status Code | Arti | Kapan Terjadi |
|-------------|------|---------------|
| `200 OK` | Berhasil | GET berhasil, PUT/PATCH berhasil |
| `201 Created` | Berhasil dibuat | POST berhasil membuat resource baru |
| `204 No Content` | Berhasil, tanpa body | DELETE berhasil |
| `400 Bad Request` | Request tidak valid | Body JSON tidak dapat diparsing |
| `401 Unauthorized` | Tidak terautentikasi | Token tidak ada, expired, atau invalid |
| `403 Forbidden` | Tidak punya izin | Token valid tapi tidak punya akses ke resource ini |
| `404 Not Found` | Resource tidak ditemukan | ID tidak ada atau sudah dihapus |
| `422 Unprocessable Entity` | Validasi gagal | Data tidak lengkap atau format salah |
| `500 Internal Server Error` | Error server | Error tidak terduga di server |

### Contoh Error Response

**401 Unauthorized (token tidak ada):**
```json
{
  "error": "missing or invalid authorization header",
  "code": "UNAUTHORIZED"
}
```

**401 Unauthorized (token expired):**
```json
{
  "error": "token has expired",
  "code": "TOKEN_EXPIRED"
}
```

**403 Forbidden:**
```json
{
  "error": "insufficient scope: company_id required",
  "code": "FORBIDDEN_SCOPE"
}
```

**422 Validation Error:**
```json
{
  "error": "validation failed",
  "code": "VALIDATION_ERROR",
  "details": {
    "name": "name is required",
    "price": "price must be greater than 0"
  }
}
```

**404 Not Found:**
```json
{
  "error": "record not found",
  "code": "NOT_FOUND"
}
```

---

## Contoh Curl Commands

Simpan token ke variabel untuk kemudahan:
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
BASE="https://api.domain.com/api/v1"
```

### Health Check

```bash
curl -s $BASE/../health
```

### Login

```bash
curl -s -X POST $BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin@company.com", "password": "password123"}' \
  | python3 -m json.tool
```

### List Resource (GET Collection)

```bash
curl -s "$BASE/products/?page=1&limit=10&sort=-created_at" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -m json.tool
```

### Get Single Resource

```bash
curl -s "$BASE/products/550e8400-e29b-41d4-a716-446655440001" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -m json.tool
```

### Create Resource (POST)

```bash
curl -s -X POST "$BASE/products/" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "_data": {
      "name": "Laptop Gaming XYZ",
      "price": 15000000,
      "stock": 10,
      "status": "active"
    },
    "_rels": {
      "category_id": "uuid-kategori-electronics"
    }
  }' \
  | python3 -m json.tool
```

### Update Resource (PUT — full replace)

```bash
curl -s -X PUT "$BASE/products/550e8400-e29b-41d4-a716-446655440001" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "_data": {
      "name": "Laptop Gaming XYZ v2",
      "price": 16000000,
      "stock": 8,
      "status": "active"
    }
  }' \
  | python3 -m json.tool
```

### Partial Update (PATCH — hanya field yang berubah)

```bash
curl -s -X PATCH "$BASE/products/550e8400-e29b-41d4-a716-446655440001" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "_data": {
      "price": 14500000
    }
  }' \
  | python3 -m json.tool
```

### Delete Resource

```bash
curl -s -X DELETE "$BASE/products/550e8400-e29b-41d4-a716-446655440001" \
  -H "Authorization: Bearer $TOKEN"
# Response: 204 No Content (body kosong)
```

### Filter dan Expand

```bash
# Filter + sort + pagination + expand relasi
curl -s "$BASE/products/?filter[status]=active&sort=-price&page=1&limit=5&expand=category" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -m json.tool
```

### Refresh Token

```bash
curl -s -X POST "$BASE/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}' \
  | python3 -m json.tool
```

---

## Swagger UI — Dokumentasi Interaktif

API menyediakan dokumentasi interaktif yang memungkinkan Anda mencoba langsung
setiap endpoint dari browser:

1. Buka: `http://localhost:8080/swagger/index.html`
2. Klik tombol **Authorize** di kanan atas
3. Masukkan token dalam format: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
4. Klik **Authorize**, lalu **Close**
5. Sekarang Anda bisa mencoba setiap endpoint langsung dari UI

> Di production, Swagger UI hanya diaktifkan jika `ENV != production` atau
> jika eksplisit diaktifkan oleh konfigurasi operator.
