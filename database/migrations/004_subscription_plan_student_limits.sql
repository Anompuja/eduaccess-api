ALTER TABLE "plans"
    ADD COLUMN IF NOT EXISTS "max_students" INT NOT NULL DEFAULT 0;

UPDATE "plans"
SET
    "description" = 'Masa percobaan untuk sekolah baru sebelum berlangganan paket berbayar.',
    "features" = '["Maks 100 siswa","Durasi trial 14 hari","1 sekolah","Absensi dan akademik dasar"]'::jsonb,
    "max_students" = 100,
    "monthly_price" = 0,
    "yearly_price" = 0,
    "is_default" = TRUE,
    "updated_at" = NOW()
WHERE LOWER("name") = 'trial';

UPDATE "plans"
SET
    "description" = 'Paket awal untuk sekolah kecil yang sudah menjalankan operasional harian.',
    "features" = '["Maks 500 siswa","1 sekolah","Dashboard, akademik, absensi, dan pelacakan siswa"]'::jsonb,
    "max_students" = 500,
    "monthly_price" = 499000,
    "yearly_price" = 4990000,
    "active" = TRUE,
    "is_default" = FALSE,
    "updated_at" = NOW()
WHERE LOWER("name") = 'basic';

UPDATE "plans"
SET
    "name" = 'Pro',
    "description" = 'Paket untuk sekolah berkembang yang butuh kapasitas lebih besar.',
    "features" = '["Maks 1500 siswa","1 sekolah","Semua fitur Basic","Laporan operasional lebih besar"]'::jsonb,
    "max_students" = 1500,
    "monthly_price" = 1299000,
    "yearly_price" = 12990000,
    "active" = TRUE,
    "is_default" = FALSE,
    "updated_at" = NOW()
WHERE LOWER("name") IN ('standard', 'pro');

UPDATE "plans"
SET
    "name" = 'Enterprise',
    "description" = 'Paket untuk sekolah besar dengan kebutuhan kapasitas tinggi.',
    "features" = '["Maks 5000 siswa","1 sekolah","Semua fitur Pro","Prioritas dukungan implementasi"]'::jsonb,
    "max_students" = 5000,
    "monthly_price" = 2999000,
    "yearly_price" = 29990000,
    "active" = TRUE,
    "is_default" = FALSE,
    "updated_at" = NOW()
WHERE LOWER("name") IN ('premium', 'enterprise');

INSERT INTO "plans" ("name", "description", "features", "max_students", "monthly_price", "yearly_price", "active", "is_default")
SELECT
    'Pro',
    'Paket untuk sekolah berkembang yang butuh kapasitas lebih besar.',
    '["Maks 1500 siswa","1 sekolah","Semua fitur Basic","Laporan operasional lebih besar"]'::jsonb,
    1500,
    1299000,
    12990000,
    TRUE,
    FALSE
WHERE NOT EXISTS (
    SELECT 1 FROM "plans" WHERE LOWER("name") = 'pro'
);

INSERT INTO "plans" ("name", "description", "features", "max_students", "monthly_price", "yearly_price", "active", "is_default")
SELECT
    'Enterprise',
    'Paket untuk sekolah besar dengan kebutuhan kapasitas tinggi.',
    '["Maks 5000 siswa","1 sekolah","Semua fitur Pro","Prioritas dukungan implementasi"]'::jsonb,
    5000,
    2999000,
    29990000,
    TRUE,
    FALSE
WHERE NOT EXISTS (
    SELECT 1 FROM "plans" WHERE LOWER("name") = 'enterprise'
);
