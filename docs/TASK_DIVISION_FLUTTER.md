# EduAccess Flutter — Pembagian Tugas (3 Developer)

> Migrasi dari: React + TypeScript (kitagiat-admin-portal)  
> Target: Flutter (mobile/desktop admin portal)  
> Total screens: 18 halaman + foundation

---

## Gambaran Pembagian

| | Dev 1 — Foundation & Auth | Dev 2 — User & Siswa | Dev 3 — Akademik & Sekolah |
|---|---|---|---|
| **Focus** | Setup proyek + Auth + Core | Manajemen pengguna | Akademik + Operasional sekolah |
| **Screens** | 5 screens | 6 screens | 7 screens |
| **Dependency** | Harus selesai duluan | Butuh foundation dari Dev 1 | Butuh foundation dari Dev 1 |
| **Kompleksitas** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

> **PENTING:** Dev 1 harus menyelesaikan foundation (setup proyek, routing, tema, API client, auth) sebelum Dev 2 dan Dev 3 mulai coding screen.

---

## DEV 1 — Foundation, Auth & Core Screens

### Tanggung Jawab Utama
> Setup arsitektur Flutter dari nol. Semua dev lain bergantung pada hasil kerja Dev 1.

---

### BAGIAN A — Project Setup & Architecture

**Tugas:**
- [ ] Init Flutter project (nama package, struktur folder)
- [ ] Setup folder structure:
  ```
  lib/
  ├── core/
  │   ├── api/         ← Dio client, interceptors, token refresh
  │   ├── auth/        ← Auth logic, token storage
  │   ├── router/      ← GoRouter routing setup
  │   ├── theme/       ← Warna, typography, ThemeData
  │   └── widgets/     ← Shared widgets (Button, Input, Card, Table, dll)
  ├── features/
  │   ├── auth/
  │   ├── dashboard/
  │   ├── users/
  │   ├── students/
  │   ├── parents/
  │   ├── academic/
  │   ├── school/
  │   └── ...
  ```
- [ ] Setup packages utama di `pubspec.yaml`:
  ```yaml
  dio: ^5.x           # HTTP client (ganti Axios)
  go_router: ^14.x    # Routing (ganti React Router)
  riverpod: ^2.x      # State management (ganti React Query + Context)
  flutter_secure_storage # Simpan token (ganti localStorage)
  shared_preferences  # Settings & preferences
  ```
- [ ] Setup `ThemeData` dengan warna brand Kitagiat:
  - Primary: `#1D3557` (dark blue)
  - Secondary: `#4E89AE` (medium blue)
  - Background: `#F5F6FA`
  - Text: `#2C2C2C`
- [ ] Setup GoRouter dengan Protected & Public routes
- [ ] Setup Dio interceptor untuk:
  - Attach `Authorization: Bearer <token>` ke semua request
  - Auto-refresh token jika 401
  - Handle 403, 404, 500 secara global

---

### BAGIAN B — Auth Module

**Reference screen lama:** `Login.tsx`, `Register.tsx`  
**Backend endpoints:** `POST /auth/login`, `POST /auth/register`, `POST /auth/refresh`, `POST /auth/logout`

**Screens:**
- [ ] **Login Screen**
  - Input: email, password
  - Validasi: required, format email, min 8 char password
  - Tombol: Login
  - Handle error: 401 (salah credential), 403 (akun nonaktif)
  - Simpan `access_token` + `refresh_token` ke secure storage
  - Redirect ke Dashboard setelah berhasil

- [ ] **Register Screen**
  - Input: name, email, password, role (dropdown)
  - Role pilihan: admin_sekolah, kepala_sekolah, guru, staff, orangtua, siswa
  - Handle error: 409 (email sudah ada)

- [ ] **Auth Provider / State**
  - Riverpod provider untuk user session
  - Logic: login, logout, refresh token, cek isAuthenticated
  - Simpan data user (id, role, school_id, name, email, avatar)

- [ ] **Role-based Navigation Guard**
  - Redirect ke login jika belum auth
  - Redirect ke dashboard jika sudah auth (akses `/login`)
  - Sembunyikan menu berdasarkan role

---

### BAGIAN C — Layout & Navigation

**Reference screen lama:** Layout components (Sidebar, Header)

- [ ] **App Shell / Main Layout**
  - Sidebar navigasi (collapsible)
  - Header (nama user, avatar, logout button)
  - Content area
  - Responsive: sidebar drawer di mobile, permanent di desktop/tablet

- [ ] **Sidebar Menu** (tampil/sembunyi per role)
  ```
  Dashboard
  Manajemen Siswa      ← (Dev 2)
  Guru & Staff         ← (Dev 2)
  Orang Tua            ← (Dev 2)
  Struktur Akademik    ← (Dev 3)
  Naik Kelas           ← (Dev 3)
  Tracking Siswa       ← (Dev 3)
  Profil Sekolah       ← (Dev 3)
  Subscription         ← (Dev 3)
  Pengaturan
  Notifikasi
  ```

---

### BAGIAN D — Core Screens

**Reference screen lama:** `Dashboard.tsx` (414 baris), `Profile.tsx`, `Settings.tsx`, `Notifications.tsx`

- [ ] **Dashboard Screen**
  - Statistik: total siswa, guru, langganan aktif
  - Quick action cards
  - Recent activity feed
  - Responsive grid layout

- [ ] **Profile Screen**
  - Tampil data user (nama, email, avatar, role)
  - Form update nama & avatar
  - Form ganti password (old + new password)
  - API: `GET /profile`, `PUT /profile`, `PUT /users/:id/password`

- [ ] **Settings Screen**
  - Preferensi aplikasi (tema, bahasa, notifikasi)
  - (Data lebih ke lokal, tidak semua perlu API)

- [ ] **Notifications Screen**
  - List notifikasi sistem
  - Mark as read

- [ ] **404 / Error Screen**

---

### Shared Widgets yang Harus Dibuat Dev 1 (dipakai semua dev)

> Buat di `lib/core/widgets/` agar bisa dipakai Dev 2 & Dev 3

| Widget | Deskripsi |
|--------|-----------|
| `AppButton` | Primary, secondary, danger button |
| `AppTextField` | Input dengan label + error state |
| `AppDropdown` | Dropdown/select field |
| `AppDatePicker` | Date picker |
| `AppCard` | Card container |
| `AppDataTable` | Tabel dengan pagination, sort, search |
| `AppPagination` | Pagination widget |
| `AppBadge` | Status badge (active, inactive, dll) |
| `AppDialog` | Modal/dialog wrapper |
| `AppConfirmDialog` | Dialog konfirmasi hapus/nonaktifkan |
| `AppSearchBar` | Search input dengan debounce |
| `AppEmptyState` | Tampilan kosong (no data) |
| `AppLoadingIndicator` | Loading spinner |
| `AppErrorState` | Tampilan error dengan retry |
| `AppToast` | Notifikasi toast (sukses/gagal) |

---

---

## DEV 2 — Manajemen Pengguna (User, Siswa, Guru, Orang Tua)

### Tanggung Jawab Utama
> Semua hal yang berkaitan dengan data pengguna di sekolah.

**Reference screen lama:** `Students.tsx` (253 baris), `ParentManagement.tsx` (2091 baris!), `TeachersStaff.tsx` (186 baris), `Users.tsx` (179 baris)  
**Backend endpoints:** `/users`, `/students`, `/parents`

---

### BAGIAN A — User Management

**Screens:**
- [ ] **Daftar Pengguna** (`/users`)
  - Tabel: nama, email, role, status, aksi
  - Filter: by role (dropdown)
  - Search: nama / email / username
  - Pagination
  - Aksi per baris: Edit, Nonaktifkan
  - API: `GET /users?role=&search=&page=&per_page=`

- [ ] **Detail Pengguna** (`/users/:id`)
  - Tampil semua info user
  - Tombol edit, ganti password, nonaktifkan
  - API: `GET /users/:id`

- [ ] **Edit Pengguna** (modal/sheet)
  - Form: nama, avatar
  - API: `PUT /users/:id`

- [ ] **Ganti Password User** (modal)
  - Input: password baru (admin/superadmin tidak perlu password lama)
  - API: `PUT /users/:id/password`

- [ ] **Nonaktifkan User** (confirm dialog)
  - API: `DELETE /users/:id`

---

### BAGIAN B — Manajemen Siswa (Fitur Terbesar)

**Screens:**
- [ ] **Daftar Siswa** (`/students`)
  - Tabel: foto, nama, NIS, kelas, status, aksi
  - Filter: education_level_id, class_id, sub_class_id
  - Search: nama / email / NIS / NISN
  - Pagination (default 20, max 100)
  - Aksi: Detail, Edit, Nonaktifkan
  - Tombol: Tambah Siswa
  - API: `GET /students?search=&education_level_id=&class_id=&sub_class_id=&page=&per_page=`

- [ ] **Detail Siswa** (`/students/:id`)
  - Info akun: nama, email, username, avatar
  - Info profil: NIS, NISN, gender, agama, TTL, alamat, tahun masuk, jalur masuk
  - Info kelas: level, kelas, sub-kelas
  - Daftar orang tua yang terhubung (nama, hubungan, primary)
  - Tombol: Edit, Link Orang Tua, Nonaktifkan
  - API: `GET /students/:id`

- [ ] **Form Tambah Siswa** (full-screen form / modal besar)
  - Section 1 — Info Akun: nama, email, username, password
  - Section 2 — Profil: NIS, NISN, gender, agama, TTL, alamat, tahun masuk, jalur masuk
  - Section 3 — Penempatan Kelas: level → kelas → sub-kelas (cascading dropdown)
  - Default password: "Siswa@12345" (tampilkan info ke user)
  - API: `POST /students`

- [ ] **Form Edit Siswa** (full-screen / bottom sheet)
  - Field sama dengan form tambah
  - API: `PUT /students/:id`

- [ ] **Link Orang Tua ke Siswa** (dialog)
  - Cari orang tua yang ada (search by nama/email)
  - Pilih relationship: Ayah / Ibu / Wali / Lainnya
  - Tandai sebagai primary (checkbox)
  - API: `POST /students/:id/parents`

- [ ] **Unlink Orang Tua** (confirm dialog)
  - API: `DELETE /students/:id/parents/:parent_id`

- [ ] **Nonaktifkan Siswa** (confirm dialog)
  - API: `DELETE /students/:id`

---

### BAGIAN C — Manajemen Orang Tua

> Ini screen terbesar di frontend lama (2091 baris). Prioritas tinggi.

**Screens:**
- [ ] **Daftar Orang Tua** (`/parents`)
  - Tabel: nama, email, no. HP, anak yang terhubung, aksi
  - Search: nama / email
  - Pagination
  - Tombol: Tambah Orang Tua
  - API: `GET /parents?search=&page=&per_page=`

- [ ] **Detail Orang Tua** (`/parents/:id`)
  - Info akun: nama, email, username
  - Info profil: nama ayah, nama ibu, agama ayah/ibu, no. HP, alamat
  - Daftar siswa yang terhubung
  - API: `GET /parents/:id`

- [ ] **Form Tambah Orang Tua**
  - Section 1 — Info Akun: nama, email, username, password
  - Section 2 — Profil: nama ayah, ibu, agama, no. HP, alamat
  - API: `POST /parents`

- [ ] **Form Edit Orang Tua**
  - Field profil (nama ayah, ibu, agama, HP, alamat)
  - API: `PUT /parents/:id`

- [ ] **Nonaktifkan Orang Tua** (confirm dialog)
  - API: `DELETE /parents/:id`

---

### BAGIAN D — Manajemen Guru & Staff

**Screens:**
- [ ] **Daftar Guru & Staff** (`/teachers-staff`)
  - Tabel: nama, email, role (guru/staff), status, aksi
  - Filter: by role (guru / staff)
  - Search: nama / email
  - Pagination
  - API: `GET /users?role=guru` dan `GET /users?role=staff`

  > Note: Di backend baru, guru dan staff adalah User biasa. Tampilkan dua tab (Guru / Staff).

- [ ] **Detail, Edit, Ganti Password, Nonaktifkan**
  - Sama dengan alur User Management di atas
  - Reuse components yang sama

---

### Field Reference Lengkap

**Siswa:**
```
Akun:  name, email, username, password
Profil: nis, nisn, phone_number, address, gender (L/P),
        religion, birth_place, birth_date (YYYY-MM-DD),
        tahun_masuk, jalur_masuk_sekolah (reguler/beasiswa/mutasi/lainnya),
        education_level_id, class_id, sub_class_id
```

**Orang Tua:**
```
Akun:  name, email, username, password
Profil: father_name, mother_name, father_religion, mother_religion,
        phone_number, address
```

**Link Orang Tua-Siswa:**
```
parent_id, relationship (father/mother/guardian/other), is_primary (bool)
```

---

---

## DEV 3 — Akademik, Sekolah & Operasional

### Tanggung Jawab Utama
> Struktur akademik sekolah, naik kelas, tracking siswa, profil sekolah, dan subscription.

**Reference screen lama:** `Academic.tsx` (251 baris), `ClassPromotion.tsx` (958 baris), `StudentTracking.tsx` (677 baris), `School.tsx` (501 baris), `Subscription.tsx` (452 baris), `Payment.tsx` (842 baris), `Reports.tsx` (276 baris)

---

### BAGIAN A — Struktur Akademik

**Reference screen lama:** `Academic.tsx`  
**Backend endpoints:** `/academic/levels`, `/academic/classes`, `/academic/sub-classes`

> Di frontend lama ada tab-tab: Level Pendidikan, Kelas, Sub-Kelas, dll. Backend baru mendukung 3 level hierarki.

**Screens:**
- [ ] **Halaman Akademik** (`/academic`) — Tabbed layout:

  **Tab 1 — Level Pendidikan**
  - List: nama level (SD, SMP, SMA, dll)
  - Tambah level (input nama)
  - Edit nama level (inline edit / dialog)
  - Hapus level (confirm dialog)
  - API: `GET/POST/PUT/DELETE /academic/levels`

  **Tab 2 — Kelas**
  - Filter by level (dropdown)
  - List: nama kelas, level induk
  - Tambah kelas (pilih level → isi nama)
  - Edit, hapus
  - API: `GET/POST/PUT/DELETE /academic/classes?level_id=`

  **Tab 3 — Sub-Kelas**
  - Filter by kelas (dropdown)
  - List: nama sub-kelas, kelas induk
  - Tambah sub-kelas (pilih kelas → isi nama)
  - Edit, hapus
  - API: `GET/POST/PUT/DELETE /academic/sub-classes?class_id=`

---

### BAGIAN B — Naik Kelas (Class Promotion)

**Reference screen lama:** `ClassPromotion.tsx` (958 baris — sangat kompleks)

> **Catatan:** Endpoint naik kelas belum tersedia di backend baru. Dev 3 koordinasi dengan backend dev dulu atau buat UI terlebih dahulu dengan data dummy.

**Screens:**
- [ ] **Halaman Naik Kelas** (`/class-promotion`)
  - Pilih tahun ajaran
  - Pilih level/kelas sumber
  - Tampil daftar siswa
  - Tandai siswa yang naik kelas / tinggal kelas
  - Pilih kelas tujuan
  - Bulk select + confirm promosi
  - Tampilkan hasil / summary

---

### BAGIAN C — Tracking Siswa

**Reference screen lama:** `StudentTracking.tsx` (677 baris)

> **Catatan:** Sama seperti class promotion, endpoint tracking belum tentu ada di backend baru. Konfirmasi dengan backend dev, atau build UI dengan data mock.

**Screens:**
- [ ] **Halaman Tracking Siswa** (`/student-tracking`)
  - Search/filter siswa
  - Tampilkan riwayat akademik siswa
  - Progress tracking per semester/tahun
  - Riwayat kelas, absensi, nilai (sesuai data yang tersedia)

---

### BAGIAN D — Profil Sekolah

**Reference screen lama:** `School.tsx` (501 baris)  
**Backend endpoints:** `/schools/:id`, `/schools/:id/rules`

**Screens:**
- [ ] **Halaman Profil Sekolah** (`/school`)
  - Tampil info sekolah: nama, alamat, no. HP, email, deskripsi, logo, timezone, status
  - Tombol Edit Sekolah
  - API: `GET /schools/:id`

- [ ] **Form Edit Sekolah** (modal/full-screen)
  - Field: nama, alamat, phone, email, deskripsi, image_path, time_zone
  - Untuk superadmin: tambahan field status (active/nonactive)
  - API: `PUT /schools/:id`

- [ ] **Halaman Rules Sekolah** (sub-section di School)
  - List key-value rules
  - Form upsert rules (tambah/edit banyak sekaligus)
  - Tombol Simpan Semua
  - API: `GET /schools/:id/rules`, `PUT /schools/:id/rules`

---

### BAGIAN E — Subscription

**Reference screen lama:** `Subscription.tsx` (452 baris) — hanya untuk Manager role  
**Backend endpoints:** `/schools/:id/subscription`

**Screens:**
- [ ] **Halaman Subscription** (`/subscription`)
  - Tampil status langganan aktif:
    - Nama plan, deskripsi, list fitur
    - Status: active / inactive / trial / expired / cancelled
    - Siklus: bulanan / tahunan / sekali bayar
    - Harga, tanggal berakhir
  - Info limit pengguna
  - Tombol Upgrade / Perpanjang (jika expired)
  - API: `GET /schools/:id/subscription`

  > **Catatan:** Fitur payment/checkout mungkin belum ada di backend baru. Koordinasi terlebih dahulu.

---

### BAGIAN F — Payment & Reports

**Reference screen lama:** `Payment.tsx` (842 baris), `Reports.tsx` (276 baris)

> **Catatan:** Backend belum memiliki endpoint payment dan reports. Buat UI placeholder dulu, atau skip sementara dan prioritaskan yang sudah ada backend-nya.

**Screens (jika sudah ada backend):**
- [ ] **Halaman Payment** (`/payment`)
  - Riwayat pembayaran
  - Status invoice
  - Proses pembayaran baru

- [ ] **Halaman Reports** (`/reports`)
  - Statistik sekolah
  - Export laporan

---

---

## Dependency & Urutan Pengerjaan

```
Minggu 1:
  Dev 1 → Foundation (setup project, theme, API client, routing, shared widgets)
  Dev 2 → Siapkan desain/wireframe User & Siswa
  Dev 3 → Siapkan desain/wireframe Akademik & Sekolah

Minggu 2:
  Dev 1 → Auth screens + Layout + Dashboard
  Dev 2 → Mulai Manajemen Siswa (butuh shared widgets dari Dev 1)
  Dev 3 → Mulai Struktur Akademik (butuh shared widgets dari Dev 1)

Minggu 3+:
  Dev 1 → Profile, Settings, Notifications + support tim
  Dev 2 → Parent Management + Guru & Staff
  Dev 3 → Sekolah, Subscription, Class Promotion
```

---

## Konvensi Kode (Wajib Disepakati Sebelum Mulai)

| Hal | Keputusan |
|-----|-----------|
| State management | Riverpod (Notifier + AsyncNotifier) |
| HTTP client | Dio + Retrofit (opsional) |
| Routing | GoRouter |
| Naming file | `snake_case.dart` |
| Naming class | `PascalCase` |
| Folder per feature | `feature/screen/`, `feature/controller/`, `feature/model/` |
| API base URL | Di `core/api/api_client.dart`, bisa swap dev/prod |
| Token storage | `flutter_secure_storage` |
| Error handling | Global via Dio interceptor + per-screen error state |
| Form validation | `flutter_form_builder` atau manual validator |
| Image/avatar upload | Koordinasi dengan backend (endpoint belum ada) |

---

## Catatan Penting

1. **Import/Export Excel** (ada di frontend lama) — backend baru belum ada endpoint ini. Skip dulu atau buat dummy.
2. **QR Code siswa** — backend baru belum ada. Skip atau pakai package `qr_flutter` saja di client.
3. **Keycloak SSO** — tidak perlu dimigrate, backend baru pakai JWT langsung.
4. **Academic Year** — belum ada di backend baru. Koordinasi dengan backend dev.
5. **Subjects & Classrooms** — ada di frontend lama tapi belum ada di backend baru. Skip dulu.
6. **Image upload (avatar/logo sekolah)** — backend menerima `image_path` (string URL/path), bukan file upload langsung. Koordinasi.

---

*Generated for EduAccess Flutter Migration — `docs/TASK_DIVISION_FLUTTER.md`*
