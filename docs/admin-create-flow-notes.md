# Admin Create API Notes

Tujuan endpoint ini: membuat akun admin sekolah dalam satu alur bisnis.

## 1. Alur End-to-End (POST /api/v1/admins)

1. Client kirim JSON request.
2. Handler bind + validasi request (DTO).
3. Handler ubah request jadi command untuk use case.
4. Use case validasi otorisasi dan konteks sekolah.
5. Use case cek unique email dan username.
6. Use case hash password.
7. Use case buat data user (role admin_sekolah).
8. Use case buat data admin_profiles yang mengacu ke user_id tadi.
9. Handler map hasil ke response JSON.

Catatan penting:

- user_id tidak diminta dari client, dibuat oleh server (uuid.New) supaya aman dan konsisten.
- school_id:
  - admin_sekolah: dari token/middleware.
  - superadmin: wajib kirim school_id di request.

## 2. Fungsi Tiap File (yang sudah dibuat)

### Delivery (HTTP)

- internal/admin/delivery/http/dto.go
  - Kontrak input/output HTTP.
  - Tempat field request, response, dan rule validasi.

- internal/admin/delivery/http/handler.go
  - Pintu masuk HTTP untuk endpoint admins.
  - Bind + validate request.
  - Mapping request DTO -> CreateAdminCommand.
  - Panggil use case.
  - Mapping domain -> response JSON.

### Application (Use Case)

- internal/admin/application/create_admin.go
  - Inti aturan bisnis create admin.
  - Cek role requester.
  - Resolve school context.
  - Cek email/username.
  - Hash password.
  - Buat user lalu buat admin profile.

- internal/admin/application/user_creator.go
  - Interface dependency untuk operasi user (Create, ExistsByEmail, ExistsByUsername).
  - Tujuan: use case tidak tergantung langsung ke implementasi auth/GORM.

### Domain (Core)

- internal/admin/domain/admin.go
  - Entity inti AdminProfile (model bisnis, bukan model HTTP).

- internal/admin/domain/repository.go
  - Kontrak repository untuk kebutuhan domain/use case admin.
  - Use case bergantung ke kontrak ini, bukan ke GORM langsung.

### Infrastructure (Data Access)

- internal/admin/infrastructure/admin_repository.go
  - Implementasi repository pakai GORM.
  - Mapping entity domain -> tabel admin_profiles.

### Bootstrap / Wiring

- cmd/main.go
  - Daftarkan repository dan handler admin.
  - Hubungkan dependency: userRepo + adminRepo -> createAdmin handler -> route /admins.

## 3. Kenapa repository contract tidak ditaruh di handler?

Karena handler adalah layer delivery (HTTP), bukan layer bisnis.
Kontrak repository ditaruh di core (domain/application) agar:

1. Use case tetap independen dari framework/DB.
2. Mudah di-test (bisa pakai mock/fake).
3. Mudah ganti implementasi data source tanpa ubah aturan bisnis.

## 4. Mental Model Singkat

- DTO: bahasa API
- Handler: controller HTTP
- Use Case: aturan bisnis
- Domain: model inti + kontrak
- Infrastructure: adapter ke database (GORM)

## 5. Hal yang bisa ditingkatkan nanti

Saat ini create user dan create admin_profile adalah dua langkah berurutan.
Agar lebih kuat, bisa dibuat transaction lintas keduanya supaya benar-benar atomic end-to-end.
