# EduAccess — Feature Specification
> Dokumen ini ditujukan untuk tim Designer, UI/UX, dan Frontend.  
> Berisi seluruh fitur yang tersedia di backend, lengkap dengan data model, akses per role, dan perilaku khusus.

---

## Ringkasan Sistem

**EduAccess** adalah platform manajemen sekolah multi-tenant (SaaS).  
Setiap sekolah adalah "tenant" yang terisolasi — data satu sekolah tidak bisa diakses sekolah lain.

**Base URL API:** `http://localhost:8080/api/v1`  
**Auth:** Bearer token (JWT) di header `Authorization: Bearer <token>`  
**Swagger UI:** `http://localhost:8080/swagger/index.html`

---

## Roles Pengguna

| Role | Deskripsi |
|------|-----------|
| `superadmin` | Admin sistem, tidak terikat ke sekolah manapun |
| `admin_sekolah` | Admin per sekolah, kelola semua data sekolah |
| `kepala_sekolah` | Kepala sekolah, akses baca penuh di sekolah |
| `guru` | Guru, akses terbatas |
| `staff` | Staf sekolah, akses terbatas |
| `orangtua` | Orang tua siswa |
| `siswa` | Siswa |

---

## Matriks Akses (RBAC)

| Fitur | superadmin | admin_sekolah | kepala_sekolah | guru | staff | orangtua | siswa |
|-------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Buat sekolah | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Lihat sekolah | Semua | Sendiri | Sendiri | Sendiri | Sendiri | Sendiri | Sendiri |
| Edit sekolah | ✅ (+ status) | ✅ (sendiri) | ❌ | ❌ | ❌ | ❌ | ❌ |
| Hapus sekolah | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Kelola Rules | ✅ | ✅ (sendiri) | ✅ (sendiri) | ❌ | ❌ | ❌ | ❌ |
| Lihat subscription | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Lihat user | Semua | Sekolah sendiri | Sekolah sendiri | Sekolah sendiri | Sekolah sendiri | Sekolah sendiri | Sendiri |
| Edit user | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Sendiri |
| Deactivate user | ✅ | ✅ (bukan admin lain) | ✅ | ✅ | ✅ | ✅ | Sendiri |
| Ganti password | ✅ (tanpa old pwd) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Buat siswa | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Kelola siswa | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Buat orang tua | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Kelola orang tua | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Kelola akademik | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |

---

## Modul 1 — Autentikasi

### Screens yang Dibutuhkan
- Halaman Login
- Halaman Register
- (Tidak ada halaman profile password di auth — ada di User Management)

### Endpoints

| Method | Path | Deskripsi | Auth |
|--------|------|-----------|------|
| POST | `/auth/register` | Daftar akun baru | Public |
| POST | `/auth/login` | Login, dapat access + refresh token | Public |
| POST | `/auth/refresh` | Perbarui token | Token |
| POST | `/auth/logout` | Logout, revoke refresh token | Token |

### Form: Register
```
name          — string, wajib, 2–100 karakter
email         — string, wajib, format email
password      — string, wajib, min 8 karakter
role          — pilihan: admin_sekolah | kepala_sekolah | guru | staff | orangtua | siswa
username      — string, opsional, 3–50 karakter alfanumerik
```

### Form: Login
```
email         — string, wajib
password      — string, wajib
```

### Token System
- **Access token** — pendek (15 menit), untuk semua request API
- **Refresh token** — 7 hari, untuk minta access token baru tanpa login ulang
- Jika akun dinonaktifkan → login return `403 Forbidden`

---

## Modul 2 — Manajemen Sekolah

### Screens yang Dibutuhkan
- Daftar sekolah (superadmin)
- Detail / profil sekolah
- Form buat / edit sekolah
- Halaman rules sekolah
- Halaman subscription

### Endpoints

| Method | Path | Deskripsi | Roles |
|--------|------|-----------|-------|
| POST | `/schools` | Buat sekolah baru | superadmin |
| GET | `/schools` | Daftar semua sekolah | Semua (scope sendiri) |
| GET | `/schools/:id` | Detail sekolah | Semua (scope sendiri) |
| PUT | `/schools/:id` | Update sekolah | superadmin, admin_sekolah |
| DELETE | `/schools/:id` | Nonaktifkan sekolah | superadmin |
| GET | `/schools/:id/rules` | Lihat rules sekolah | Tenant |
| PUT | `/schools/:id/rules` | Upsert rules sekolah | admin_sekolah, superadmin |
| GET | `/schools/:id/subscription` | Lihat subscription | superadmin, admin_sekolah, kepala_sekolah |

### Form: Buat / Edit Sekolah
```
name          — string, wajib, 2–191 karakter
address       — string, opsional, max 191
phone         — string, opsional, max 50
email         — string, opsional, format email
description   — string, opsional
image_path    — string, opsional (URL/path logo sekolah)
time_zone     — string, opsional (default: "Asia/Jakarta")
status        — "active" | "nonactive" (hanya superadmin)
```

### Data: School Rules
Rules adalah key-value bebas per sekolah (konfigurasi khusus sekolah).
```
key           — string, max 191
value         — string, max 191
note          — string, opsional
```

### Data: Subscription
```
status        — active | inactive | trial | expired | cancelled
cycle         — month | year | onetime
quantity      — integer
price         — integer (Rupiah)
ends_at       — tanggal berakhir (nullable)
plan.name     — nama paket
plan.features — list fitur paket
plan.monthly_price, plan.yearly_price
```

### Perilaku Khusus
- Sekolah baru dibuat dengan status `nonactive`
- Soft delete: sekolah tidak benar-benar dihapus, hanya dinonaktifkan
- Superadmin melihat semua sekolah; role lain hanya melihat sekolah sendiri

---

## Modul 3 — Manajemen Pengguna

### Screens yang Dibutuhkan
- Daftar pengguna (dengan filter role & search)
- Detail pengguna
- Form edit pengguna
- Form ganti password
- Halaman profil sendiri (My Profile)

### Endpoints

| Method | Path | Deskripsi | Roles |
|--------|------|-----------|-------|
| GET | `/users` | Daftar pengguna | Semua |
| GET | `/users/:id` | Detail pengguna | Semua |
| PUT | `/users/:id` | Update nama & avatar | Semua |
| DELETE | `/users/:id` | Nonaktifkan user | Semua (dengan batasan) |
| PUT | `/users/:id/password` | Ganti password | Semua |
| GET | `/profile` | Profil sendiri | Semua |
| PUT | `/profile` | Update profil sendiri | Semua |

### Query Params: List Users
```
role          — filter role (opsional)
search        — cari nama / email / username (opsional)
page          — halaman (default: 1)
per_page      — jumlah per halaman (default: 20, max: 100)
```

### Form: Update User
```
name          — string, opsional, 2–100 karakter
avatar        — string, opsional, max 255 (URL/path foto)
```

### Form: Ganti Password
```
old_password  — string, wajib kecuali superadmin, min 8
new_password  — string, wajib, min 8
```

### Data: User
```
id            — UUID
school_id     — UUID (null untuk superadmin)
role          — string
name          — string
username      — string
email         — string
avatar        — string (default: "default.png")
verified      — boolean (default: false)
created_at    — timestamp
updated_at    — timestamp
```

### Perilaku Khusus
- Admin sekolah tidak bisa nonaktifkan admin sekolah lain di sekolah yang sama
- Superadmin bisa ganti password user lain tanpa tahu password lama
- Soft delete: user tidak benar-benar dihapus

---

## Modul 4 — Manajemen Siswa

### Screens yang Dibutuhkan
- Daftar siswa (dengan filter kelas, level, search)
- Detail siswa (termasuk info orang tua)
- Form tambah siswa
- Form edit siswa
- Kelola orang tua siswa (link/unlink)

### Endpoints

| Method | Path | Deskripsi | Roles |
|--------|------|-----------|-------|
| POST | `/students` | Tambah siswa baru | superadmin, admin_sekolah |
| GET | `/students` | Daftar siswa | Semua |
| GET | `/students/:id` | Detail siswa + orang tua | Semua |
| PUT | `/students/:id` | Update profil siswa | Semua |
| DELETE | `/students/:id` | Nonaktifkan siswa | Semua |
| POST | `/students/:id/parents` | Link orang tua ke siswa | Semua |
| DELETE | `/students/:id/parents/:parent_id` | Unlink orang tua | Semua |

### Query Params: List Students
```
search              — cari nama / email / NIS / NISN (opsional)
education_level_id  — filter level pendidikan (opsional)
class_id            — filter kelas (opsional)
sub_class_id        — filter sub-kelas (opsional)
page                — halaman (default: 1)
per_page            — per halaman (default: 20, max: 100)
```

### Form: Tambah / Edit Siswa
**Info Akun (untuk buat akun login siswa):**
```
name                — string, wajib, 2–191
email               — string, wajib, format email
username            — string, opsional, 3–50
password            — string, opsional, min 8 (default: "Siswa@12345")
```
**Profil Siswa:**
```
nis                 — string, opsional (Nomor Induk Siswa)
nisn                — string, opsional (Nomor Induk Siswa Nasional)
phone_number        — string, opsional, max 50
address             — string, opsional
gender              — "L" (Laki-laki) | "P" (Perempuan)
religion            — string, opsional
birth_place         — string, opsional
birth_date          — date, opsional, format YYYY-MM-DD
tahun_masuk         — string, opsional (tahun masuk sekolah)
jalur_masuk_sekolah — "reguler" | "beasiswa" | "mutasi" | "lainnya"
education_level_id  — UUID, opsional
class_id            — UUID, opsional
sub_class_id        — UUID, opsional
```

### Form: Link Orang Tua
```
parent_id     — UUID, wajib (ID profil orang tua yang sudah ada)
relationship  — "father" | "mother" | "guardian" | "other"
is_primary    — boolean
```

### Data: Siswa (Response)
```
id, user_id, school_id
name, email, username, avatar        — dari akun user
nis, nisn, phone_number, address
gender, religion, birth_place, birth_date
tahun_masuk, jalur_masuk_sekolah
education_level_id, class_id, sub_class_id
parents []                           — list orang tua yang terhubung
created_at, updated_at
```

---

## Modul 5 — Manajemen Orang Tua

### Screens yang Dibutuhkan
- Daftar orang tua
- Detail orang tua
- Form tambah / edit orang tua

### Endpoints

| Method | Path | Deskripsi | Roles |
|--------|------|-----------|-------|
| POST | `/parents` | Tambah orang tua baru | superadmin, admin_sekolah |
| GET | `/parents` | Daftar orang tua | Semua |
| GET | `/parents/:id` | Detail orang tua | Semua |
| PUT | `/parents/:id` | Update profil orang tua | Semua |
| DELETE | `/parents/:id` | Nonaktifkan orang tua | Semua |

### Query Params: List Parents
```
search        — cari nama / email (opsional)
page          — halaman (default: 1)
per_page      — per halaman (default: 20, max: 100)
```

### Form: Tambah / Edit Orang Tua
**Info Akun:**
```
name          — string, wajib, 2–191
email         — string, wajib, format email
username      — string, opsional, 3–50
password      — string, opsional, min 8
```
**Profil Orang Tua:**
```
father_name     — string, opsional, max 191
mother_name     — string, opsional, max 191
father_religion — string, opsional
mother_religion — string, opsional
phone_number    — string, opsional, max 50
address         — string, opsional
```

### Data: Orang Tua (Response)
```
id, user_id, school_id
name, email, username, avatar        — dari akun user
father_name, mother_name
father_religion, mother_religion
phone_number, address
created_at, updated_at
```

---

## Modul 6 — Struktur Akademik

Hierarki: **Level Pendidikan → Kelas → Sub-Kelas**  
Contoh: `SMP → Kelas 7 → 7A`

### Screens yang Dibutuhkan
- Halaman kelola Level Pendidikan
- Halaman kelola Kelas (per level)
- Halaman kelola Sub-Kelas (per kelas)
- Dropdown/select di form siswa

### 6.1 Level Pendidikan

| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/academic/levels` | Buat level (misal: SMP, SMA) |
| GET | `/academic/levels` | Daftar semua level |
| PUT | `/academic/levels/:id` | Update nama level |
| DELETE | `/academic/levels/:id` | Hapus level |

**Form:**
```
name    — string, wajib, 1–191
```

### 6.2 Kelas

| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/academic/classes` | Buat kelas (misal: Kelas 7) |
| GET | `/academic/classes` | Daftar kelas, filter by `?level_id=` |
| PUT | `/academic/classes/:id` | Update nama kelas |
| DELETE | `/academic/classes/:id` | Hapus kelas |

**Form:**
```
level_id  — UUID, wajib (parent level)
name      — string, wajib, 1–191
```

### 6.3 Sub-Kelas

| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/academic/sub-classes` | Buat sub-kelas (misal: 7A, 7B) |
| GET | `/academic/sub-classes` | Daftar sub-kelas, filter by `?class_id=` |
| PUT | `/academic/sub-classes/:id` | Update nama sub-kelas |
| DELETE | `/academic/sub-classes/:id` | Hapus sub-kelas |

**Form:**
```
class_id  — UUID, wajib (parent kelas)
name      — string, wajib, 1–191
```

---

## Format Response API

Semua response mengikuti envelope yang konsisten:

### Success (data tunggal)
```json
{
  "success": true,
  "message": "deskripsi pesan",
  "data": { ... }
}
```

### Success (list/paginated)
```json
{
  "success": true,
  "message": "deskripsi pesan",
  "data": [ ... ],
  "page": 1,
  "per_page": 20,
  "total": 100
}
```

### Error
```json
{
  "success": false,
  "message": "deskripsi error",
  "errors": { "field": "pesan validasi" }
}
```

### HTTP Status Codes
| Code | Arti |
|------|------|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request (data tidak valid) |
| 401 | Unauthorized (belum login) |
| 403 | Forbidden (tidak punya akses) |
| 404 | Not Found |
| 409 | Conflict (data sudah ada, misal email duplikat) |
| 422 | Unprocessable Entity (validasi gagal) |
| 500 | Internal Server Error |

---

## Perilaku Sistem Penting

| Perilaku | Detail |
|----------|--------|
| **Soft Delete** | Data tidak benar-benar dihapus, hanya ditandai `deleted_at`. Tidak tampil di list. |
| **Tenant Scoping** | Setiap sekolah hanya bisa akses datanya sendiri. Superadmin bisa akses semua. |
| **Pagination** | Default 20 item/halaman, max 100. |
| **Email Unik** | Email harus unik di seluruh sistem (lintas sekolah). |
| **Password Default Siswa** | Jika tidak diisi, password siswa default: `Siswa@12345` |
| **Username Default** | Jika tidak diisi, username diambil dari bagian depan email (sebelum @) |
| **Avatar Default** | `default.png` |
| **Status Sekolah Baru** | Baru dibuat = `nonactive`, harus diaktifkan superadmin |
| **Token Rotation** | Setiap refresh mengeluarkan token baru (access + refresh) |
| **bcrypt Password** | Password di-hash dengan bcrypt sebelum disimpan |

---

## Halaman / Flow yang Perlu Dibuat (Rekomendasi untuk Designer)

### Alur Superadmin
1. Login
2. Dashboard → statistik semua sekolah
3. Daftar sekolah → buat sekolah baru → aktifkan sekolah
4. Kelola user semua sekolah

### Alur Admin Sekolah
1. Login
2. Dashboard sekolah
3. Setup akademik: buat Level → Kelas → Sub-Kelas
4. Kelola siswa: tambah, edit, assign ke kelas, link orang tua
5. Kelola orang tua
6. Kelola user sekolah (guru, staff)
7. Settings: edit profil sekolah, rules, lihat subscription

### Alur Guru / Staff
1. Login
2. Lihat daftar siswa & kelas
3. Lihat profil siswa & orang tua

### Alur Orangtua
1. Login
2. Lihat profil anak
3. Edit profil sendiri

### Alur Siswa
1. Login
2. Lihat profil sendiri
3. Ganti password

---

*Generated from EduAccess API codebase — `docs/FEATURE_SPEC.md`*
