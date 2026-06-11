# EduAccess API

Multi-tenant School Management SaaS backend — Go, Echo, GORM, Supabase Auth, PostgreSQL.

---

## Quick Start (5 menit)

```bash
# 1. Clone & masuk ke folder
git clone https://github.com/Anompuja/eduaccess-api.git
cd eduaccess-api

# 2. Install dependencies
go mod tidy

# 3. Salin file environment dan isi kredensial Supabase
cp .env.example .env
# Buka .env dan isi: DATABASE_URL, SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY, SUPABASE_JWT_SECRET

# 4. Generate Swagger docs
swag init -g cmd/main.go --output docs

# 5. Jalankan server
go run ./cmd/main.go
```

Server berjalan di `http://localhost:8080`
Swagger UI tersedia di `http://localhost:8080/swagger/index.html`

---

## Test Accounts

> Gunakan akun berikut untuk mencoba API. Kredensial ini sudah terseed di Supabase project yang digunakan.

| Role           | Email                          | Password   | Catatan                            |
| -------------- | ------------------------------ | ---------- | ---------------------------------- |
| `superadmin`   | `superadmin@eduaccess.com`     | `Test1234!`| Akses penuh platform, tanpa school |
| `admin_sekolah`| `adminsekolah@eduaccess.com`   | `Test1234!`| Akses penuh dalam satu sekolah     |

**Login:**
```bash
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "superadmin@eduaccess.com",
  "password": "Test1234!"
}
```

Salin `access_token` dari response, gunakan sebagai `Authorization: Bearer <token>` di semua request berikutnya.

> Di Swagger UI: klik tombol **Authorize** (kanan atas) → masukkan `Bearer <token>`.

**Catatan penting:** `school_id` otomatis dibaca dari JWT — tidak perlu dikirim manual di request body (kecuali untuk superadmin pada endpoint tertentu).

---

## Prerequisites

- Go 1.22+
- `swag` CLI — install sekali:
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```
- Akses ke Supabase project (credentials diberikan via LMS)

---

## Environment Variables

Salin `.env.example` ke `.env`:

| Variable                   | Wajib | Keterangan                                                         |
| -------------------------- | ----- | ------------------------------------------------------------------ |
| `DATABASE_URL`             | Ya    | Supabase PostgreSQL connection string (session mode, port 5432)    |
| `SUPABASE_URL`             | Ya    | URL project Supabase (`https://[ref].supabase.co`)                 |
| `SUPABASE_SERVICE_ROLE_KEY`| Ya    | Service role key untuk Admin API (buat/hapus user)                 |
| `SUPABASE_JWT_SECRET`      | Ya    | JWT secret dari Supabase dashboard — untuk validasi token          |
| `APP_PORT`                 | Tidak | Port server, default `8080`                                        |
| `CORS_ALLOW_ORIGINS`       | Tidak | Origins yang diizinkan, default `*`                                |
| `SWAGGER_HOST`             | Tidak | Hostname untuk Swagger UI, default `localhost:8080`                |
| `SWAGGER_SCHEME`           | Tidak | `http` untuk lokal, `https` untuk deployment                       |

---

## Arsitektur

Setiap modul di bawah `internal/{feature}/` mengikuti pola **DDD Layered**:

```
internal/{feature}/
├── domain/           # Entity struct + Repository interface (tanpa framework)
├── application/      # Use-case handler (satu file per operasi: create, list, get, dst)
├── delivery/http/    # Echo handler + DTO (translate HTTP → domain → HTTP)
└── infrastructure/   # Implementasi repository via GORM
```

```
internal/
├── admin/            # Admin sekolah CRUD
├── auth/             # Login, register, refresh (via Supabase Auth)
├── headmaster/       # Kepala sekolah CRUD
├── notification/     # Notifikasi (REST + WebSocket)
├── parent/           # Orang tua CRUD
├── school/           # Sekolah, langganan, rules
├── student/          # Siswa CRUD + parent linking + in-memory cache
├── teacher/          # Guru CRUD
├── staff/            # Staff CRUD
├── academic/         # Level, kelas, sub-kelas, jadwal
├── attendance/       # Absensi via QR code
└── shared/
    ├── middleware/   # JWT auth (ES256 via Supabase JWKS)
    ├── httpcache/    # HTTP caching middleware (ETag + Cache-Control)
    └── response/     # Standar JSON response helper
```

---

## WebSocket — Notifikasi Real-time

**Endpoint:** `GET /ws/notifications?token=<JWT>`

WebSocket Hub mengelola semua koneksi aktif per user. Saat event terjadi (contoh: siswa absen), backend langsung *broadcast* payload JSON ke semua koneksi orang tua yang sedang online — tanpa client perlu polling.

**Contoh koneksi dari Flutter:**
```dart
final channel = WebSocketChannel.connect(
  Uri.parse('ws://localhost:8080/ws/notifications?token=$accessToken'),
);
await for (final msg in channel.stream) {
  final data = jsonDecode(msg);
  // handle notifikasi masuk
}
```

**Payload yang dikirim server:**
```json
{
  "id": "uuid",
  "type": "attendance",
  "title": "Kehadiran Budi Santoso",
  "body": "Budi Santoso telah hadir di kelas Matematika (Ruang 1) pukul 07:30",
  "data": { "student_name": "...", "attendance_status": "present" },
  "createdAt": "2026-01-01T07:30:00Z"
}
```

---

## Caching

Dua strategi caching diterapkan:

### 1. In-Memory Cache — Student List

`GET /students` menggunakan `go-cache` (in-process). Cache key dibentuk dari kombinasi role + school_id + filter parameter + pagination. TTL: 30 detik.

Ketika ada student dibuat/diupdate/dihapus, cache diinvalidasi otomatis via `InvalidatePrefix("student:list:")`.

### 2. HTTP ETag Cache — Endpoint Read-Heavy

Middleware `httpcache` di `internal/shared/httpcache/cache.go` menangani endpoint lain:

| Preset               | Cache-Control          | Digunakan untuk          |
| -------------------- | ---------------------- | ------------------------ |
| `Profile`            | `private, max-age=300` | Data profil user         |
| `Stats`              | `private, max-age=60`  | Dashboard stats          |
| `Reference`          | `private, max-age=120` | Data akademik (level/kelas) |
| `ShortLived`         | `private, max-age=30`  | CRUD list (parent, dll)  |
| `AlwaysRevalidate`   | `private, no-cache`    | Data operasional live    |

Setiap response mendapat header `ETag` (SHA-256 dari body). Request berikutnya dengan `If-None-Match` yang cocok mendapat `304 Not Modified` — hemat bandwidth tanpa data stale.

---

## Auth Flow

Token dikeluarkan oleh **Supabase Auth** dan ditandatangani dengan **ES256 (ECDSA)**. Validasi di backend menggunakan public key yang di-fetch dari JWKS endpoint Supabase — bukan JWT_SECRET biasa.

Custom hook Supabase (`public.custom_access_token_hook`) menyisipkan `app_role` dan `school_id` ke dalam setiap token, sehingga middleware bisa membaca role dan tenant tanpa query database.

```
Client                          EduAccess API
  |                                  |
  |-- POST /auth/login ------------->|
  |   (email + password)             |--- Supabase.SignIn() -->[ Supabase Auth ]
  |                                  |<-- access_token (ES256) ---
  |<-- { access_token, refresh_token }
  |                                  |
  |-- GET /students ---------------->|
  |   Authorization: Bearer <token>  |-- validate ECDSA JWKS
  |                                  |-- extract app_role, school_id
  |<-- 200 { data: [...] } ----------|
```

---

## API Reference

Base path: `/api/v1` — semua endpoint butuh `Authorization: Bearer <token>` kecuali `/auth/login` dan `/auth/register`.

Lihat dokumentasi lengkap di Swagger: `http://localhost:8080/swagger/index.html`

### Ringkasan Endpoint

| Modul           | Endpoint                        | Keterangan                        |
| --------------- | ------------------------------- | --------------------------------- |
| **Auth**        | `POST /auth/login`              | Login, dapat access + refresh token |
|                 | `POST /auth/register`           | Daftar user baru                  |
|                 | `POST /auth/refresh`            | Refresh access token              |
|                 | `GET  /auth/me`                 | Identitas user dari JWT           |
| **Students**    | `GET /students`                 | List siswa (paginated, cached)    |
|                 | `POST /students`                | Buat siswa baru                   |
|                 | `GET /students/:id`             | Detail siswa                      |
|                 | `PUT /students/:id`             | Update siswa                      |
|                 | `DELETE /students/:id`          | Nonaktifkan siswa                 |
|                 | `POST /students/:id/parents`    | Link orang tua ke siswa           |
| **Notifications**| `GET /notifications`           | List notifikasi (paginated)       |
|                 | `PATCH /notifications/:id/read` | Tandai notifikasi terbaca         |
|                 | `WS /ws/notifications`          | Stream notifikasi real-time       |
| **Schools**     | `GET /schools`                  | List sekolah (superadmin)         |
|                 | `POST /schools`                 | Buat sekolah baru                 |
| **Academic**    | `GET /academic/levels`          | Level pendidikan                  |
|                 | `GET /academic/classes`         | Kelas                             |
|                 | `GET /academic/subclasses`      | Sub-kelas                         |
| **Attendance**  | `GET /class-schedules/:id/qr`   | Generate QR token absensi         |
|                 | `POST /attendance/scan`         | Scan QR, catat kehadiran          |

---

## Response Format

Semua endpoint mengembalikan envelope JSON yang konsisten:

```json
// Success (single)
{ "success": true,  "message": "student retrieved", "data": { ... } }

// Success (paginated)
{ "success": true,  "message": "students retrieved", "data": [...], "page": 1, "per_page": 20, "total": 150 }

// Error
{ "success": false, "message": "access denied" }
```

---

## Regenerate Swagger

Jalankan ini setiap kali mengubah anotasi handler:

```bash
swag init -g cmd/main.go --output docs
```

---

## Health Check

```
GET /health → 200 { "status": "ok" }
```
