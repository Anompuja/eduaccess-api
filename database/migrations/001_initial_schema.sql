-- ============================================================
-- EduAccess API — Initial Schema Migration
-- File    : 001_initial_schema.sql
-- Target  : PostgreSQL 15 / Supabase

-- ============================================================
-- AUTH / IAM
-- ============================================================

CREATE TABLE "users" (
    "id"                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"                      VARCHAR(191),
    "username"                  VARCHAR(191) NOT NULL,
    "email"                     VARCHAR(191) NOT NULL,
    "password"                  VARCHAR(191),
    "avatar"                    VARCHAR(191) NOT NULL DEFAULT 'default.png',
    "qr_code"                   VARCHAR(191),
    "email_verified_at"         TIMESTAMPTZ,
    "verification_code"         VARCHAR(191),
    "verified"                  BOOLEAN NOT NULL DEFAULT FALSE,
    "two_factor_secret"         TEXT,
    "two_factor_recovery_codes" TEXT,
    "two_factor_confirmed_at"   TIMESTAMPTZ,
    "trial_ends_at"             TIMESTAMPTZ,
    "deleted_at"                TIMESTAMPTZ,
    "created_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "users_email_unique"    UNIQUE ("email"),
    CONSTRAINT "users_username_unique" UNIQUE ("username")
);

CREATE TABLE "roles" (
    "id"           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"         VARCHAR(50)  NOT NULL,
    "guard_name"   VARCHAR(50)  NOT NULL DEFAULT 'web',
    "display_name" VARCHAR(191),
    "description"  TEXT,
    "created_at"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "roles_name_guard_unique" UNIQUE ("name", "guard_name")
);

CREATE TABLE "permissions" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"       VARCHAR(191) NOT NULL,
    "guard_name" VARCHAR(50)  NOT NULL DEFAULT 'web',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "permissions_name_guard_unique" UNIQUE ("name", "guard_name")
);

CREATE TABLE "model_has_roles" (
    "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "role_id" UUID NOT NULL REFERENCES "roles"("id") ON DELETE CASCADE,

    PRIMARY KEY ("user_id", "role_id")
);

CREATE TABLE "model_has_permissions" (
    "user_id"       UUID NOT NULL REFERENCES "users"("id")       ON DELETE CASCADE,
    "permission_id" UUID NOT NULL REFERENCES "permissions"("id") ON DELETE CASCADE,

    PRIMARY KEY ("user_id", "permission_id")
);

CREATE TABLE "role_has_permissions" (
    "role_id"       UUID NOT NULL REFERENCES "roles"("id")       ON DELETE CASCADE,
    "permission_id" UUID NOT NULL REFERENCES "permissions"("id") ON DELETE CASCADE,

    PRIMARY KEY ("role_id", "permission_id")
);

-- JWT refresh tokens (new — KitaGiat used Laravel Passport OAuth)
CREATE TABLE "refresh_tokens" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"    UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "token"      TEXT        NOT NULL,
    "expires_at" TIMESTAMPTZ NOT NULL,
    "revoked_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "refresh_tokens_token_unique" UNIQUE ("token")
);

CREATE TABLE "api_keys" (
    "id"           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"      UUID         NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "name"         VARCHAR(191) NOT NULL,
    "key"          VARCHAR(64)  NOT NULL,
    "last_used_at" TIMESTAMPTZ,
    "created_at"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "api_keys_key_unique" UNIQUE ("key")
);

-- ============================================================
-- BILLING / SUBSCRIPTION
-- ============================================================

CREATE TABLE "plans" (
    "id"            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"          VARCHAR(191) NOT NULL,
    "description"   TEXT,
    "features"      JSONB        NOT NULL DEFAULT '[]',
    "monthly_price" BIGINT       NOT NULL DEFAULT 0,
    "yearly_price"  BIGINT       NOT NULL DEFAULT 0,
    "onetime_price" BIGINT,
    "active"        BOOLEAN NOT NULL DEFAULT TRUE,
    "is_default"    BOOLEAN NOT NULL DEFAULT FALSE,
    "deleted_at"    TIMESTAMPTZ,
    "created_at"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- subscriptions.school_id FK added after schools table
CREATE TABLE "subscriptions" (
    "id"                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"              UUID NOT NULL,
    "plan_id"                UUID NOT NULL REFERENCES "plans"("id"),
    "status"                 VARCHAR(50) NOT NULL CHECK ("status" IN ('active','inactive','trial','expired','cancelled')),
    "cycle"                  VARCHAR(50) NOT NULL DEFAULT 'month' CHECK ("cycle" IN ('month','year','onetime')),
    "quantity"               INT    NOT NULL DEFAULT 1,
    "price"                  BIGINT NOT NULL DEFAULT 0,
    "vendor_slug"            VARCHAR(191),
    "vendor_product_id"      VARCHAR(191),
    "vendor_transaction_id"  VARCHAR(191),
    "vendor_customer_id"     VARCHAR(191),
    "vendor_subscription_id" VARCHAR(191),
    "trial_ends_at"          TIMESTAMPTZ,
    "ends_at"                TIMESTAMPTZ,
    "created_at"             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- SCHOOL (TENANT CORE)
-- ============================================================

CREATE TABLE "schools" (
    "id"            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "headmaster_id" UUID,
    "name"          VARCHAR(191) NOT NULL,
    "address"       VARCHAR(191),
    "phone"         VARCHAR(50),
    "email"         VARCHAR(191),
    "description"   TEXT,
    "image_path"    VARCHAR(191),
    "time_zone"     VARCHAR(100) NOT NULL DEFAULT 'Asia/Jakarta',
    "status"        VARCHAR(50)  NOT NULL DEFAULT 'nonactive' CHECK ("status" IN ('active','nonactive')),
    "deleted_at"    TIMESTAMPTZ,
    "created_at"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "schools_headmaster_id_fkey" FOREIGN KEY ("headmaster_id")
        REFERENCES "users"("id") ON DELETE SET NULL
);

-- Deferred FK now that schools exists
ALTER TABLE "subscriptions"
    ADD CONSTRAINT "subscriptions_school_id_fkey"
    FOREIGN KEY ("school_id") REFERENCES "schools"("id") ON DELETE CASCADE;

CREATE TABLE "school_users" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"    UUID NOT NULL REFERENCES "users"("id")   ON DELETE CASCADE,
    "school_id"  UUID NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "deleted_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "school_users_user_school_unique" UNIQUE ("user_id", "school_id")
);

CREATE TABLE "school_rules" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"  UUID         NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "key"        VARCHAR(191) NOT NULL,
    "value"      VARCHAR(191) NOT NULL,
    "note"       TEXT,
    "deleted_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "school_rules_school_key_unique" UNIQUE ("school_id", "key")
);

CREATE TABLE "whatsapp_accounts" (
    "id"                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"            UUID         NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "account_name"         VARCHAR(191) NOT NULL,
    "service_whatsapp_id"  VARCHAR(191),
    "phone_number"         VARCHAR(50),
    "access_token"         TEXT,
    "created_at"           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ACADEMIC STRUCTURE
-- ============================================================

CREATE TABLE "school_academic_years" (
    "id"          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"   UUID         NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "name"        VARCHAR(191),
    "description" TEXT,
    "start_date"  DATE        NOT NULL,
    "end_date"    DATE        NOT NULL,
    "is_active"   BOOLEAN     NOT NULL DEFAULT FALSE,
    "deleted_at"  TIMESTAMPTZ,
    "created_at"  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "school_education_levels" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"  UUID         NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "name"       VARCHAR(191) NOT NULL,
    "deleted_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "school_classes" (
    "id"                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"                 UUID         NOT NULL REFERENCES "schools"("id")                ON DELETE CASCADE,
    "school_education_level_id" UUID         NOT NULL REFERENCES "school_education_levels"("id") ON DELETE CASCADE,
    "name"                      VARCHAR(191) NOT NULL,
    "deleted_at"                TIMESTAMPTZ,
    "created_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "school_sub_classes" (
    "id"              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"       UUID         NOT NULL REFERENCES "schools"("id")       ON DELETE CASCADE,
    "school_class_id" UUID         NOT NULL REFERENCES "school_classes"("id") ON DELETE CASCADE,
    "name"            VARCHAR(191) NOT NULL,
    "deleted_at"      TIMESTAMPTZ,
    "created_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "school_classrooms" (
    "id"                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"               UUID NOT NULL REFERENCES "schools"("id")               ON DELETE CASCADE,
    "school_class_id"         UUID REFERENCES "school_classes"("id")                 ON DELETE SET NULL,
    "school_sub_class_id"     UUID REFERENCES "school_sub_classes"("id")             ON DELETE SET NULL,
    "school_academic_year_id" UUID REFERENCES "school_academic_years"("id")          ON DELETE SET NULL,
    "homeroom_teacher_id"     UUID REFERENCES "users"("id")                          ON DELETE SET NULL,
    "name"                    VARCHAR(191),
    "code_room"               VARCHAR(50),
    "floor"                   VARCHAR(50),
    "building"                VARCHAR(191),
    "capacity"                INT,
    "facilities"              JSONB        NOT NULL DEFAULT '[]',
    "room_type"               VARCHAR(100),
    "status"                  VARCHAR(50)  NOT NULL DEFAULT 'unknown'
                                  CHECK ("status" IN ('unknown','available','occupied','maintenance')),
    "deleted_at"              TIMESTAMPTZ,
    "created_at"              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- PEOPLE PROFILES
-- ============================================================

CREATE TABLE "teacher_profiles" (
    "id"                                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"                             UUID NOT NULL REFERENCES "users"("id")   ON DELETE CASCADE,
    "school_id"                           UUID NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "nip"                                 VARCHAR(191),
    "nuptk"                               VARCHAR(191),
    "phone_number"                        VARCHAR(50),
    "address"                             TEXT,
    "gender"                              VARCHAR(50),
    "religion"                            VARCHAR(100),
    "birth_place"                         VARCHAR(191),
    "birth_date"                          DATE,
    "nik"                                 VARCHAR(50),
    "ktp_image_path"                      VARCHAR(191),
    "kewarganegaraan"                     VARCHAR(100),
    "golongan_darah"                      VARCHAR(10),
    "berat_badan"                         VARCHAR(20),
    "tinggi_badan"                        VARCHAR(20),
    "penyakit_yang_sering_kambuh"         TEXT,
    "kelainan_jasmani"                    TEXT,
    "penyakit_kronis_yang_pernah_diderita" TEXT,
    "rt_rw"                               VARCHAR(50),
    "kode_pos"                            VARCHAR(20),
    "pendidikan_terakhir"                 VARCHAR(100),
    "jurusan"                             VARCHAR(191),
    "tahun_lulus"                         VARCHAR(10),
    "tahun_masuk"                         VARCHAR(10),
    "deleted_at"                          TIMESTAMPTZ,
    "created_at"                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Kept separate from teacher_profiles per user decision
CREATE TABLE "headmaster_profiles" (
    "id"             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"        UUID NOT NULL REFERENCES "users"("id")   ON DELETE CASCADE,
    "school_id"      UUID NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "phone_number"   VARCHAR(50),
    "address"        TEXT,
    "gender"         VARCHAR(50),
    "religion"       VARCHAR(100),
    "birth_place"    VARCHAR(191),
    "birth_date"     DATE,
    "nik"            VARCHAR(50),
    "ktp_image_path" VARCHAR(191),
    "deleted_at"     TIMESTAMPTZ,
    "created_at"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "staff_profiles" (
    "id"             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"        UUID NOT NULL REFERENCES "users"("id")   ON DELETE CASCADE,
    "school_id"      UUID NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "phone_number"   VARCHAR(50),
    "address"        TEXT,
    "gender"         VARCHAR(50),
    "religion"       VARCHAR(100),
    "birth_place"    VARCHAR(191),
    "birth_date"     DATE,
    "nik"            VARCHAR(50),
    "ktp_image_path" VARCHAR(191),
    "deleted_at"     TIMESTAMPTZ,
    "created_at"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "parent_profiles" (
    "id"              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"         UUID NOT NULL REFERENCES "users"("id")   ON DELETE CASCADE,
    "school_id"       UUID NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "father_name"     VARCHAR(191),
    "mother_name"     VARCHAR(191),
    "father_religion" VARCHAR(100),
    "mother_religion" VARCHAR(100),
    "phone_number"    VARCHAR(50),
    "address"         TEXT,
    "deleted_at"      TIMESTAMPTZ,
    "created_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "student_profiles" (
    "id"                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"                   UUID NOT NULL REFERENCES "users"("id")   ON DELETE CASCADE,
    "school_id"                 UUID NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "nis"                       VARCHAR(191),
    "nisn"                      VARCHAR(191),
    "phone_number"              VARCHAR(50),
    "address"                   TEXT,
    "gender"                    VARCHAR(50),
    "religion"                  VARCHAR(100),
    "birth_place"               VARCHAR(191),
    "birth_date"                DATE,
    "tahun_masuk"               VARCHAR(10),
    "jalur_masuk_sekolah"       VARCHAR(50)
                                    CHECK ("jalur_masuk_sekolah" IN ('reguler','beasiswa','mutasi','lainnya')),
    "school_education_level_id" UUID REFERENCES "school_education_levels"("id") ON DELETE SET NULL,
    "school_class_id"           UUID REFERENCES "school_classes"("id")           ON DELETE SET NULL,
    "school_sub_class_id"       UUID REFERENCES "school_sub_classes"("id")       ON DELETE SET NULL,
    "deleted_at"                TIMESTAMPTZ,
    "created_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- NEW: replaces student_profiles.parent_id (1:1 → proper many-to-many)
CREATE TABLE "student_parent_links" (
    "id"           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"    UUID NOT NULL REFERENCES "schools"("id")          ON DELETE CASCADE,
    "student_id"   UUID NOT NULL REFERENCES "student_profiles"("id") ON DELETE CASCADE,
    "parent_id"    UUID NOT NULL REFERENCES "parent_profiles"("id")  ON DELETE CASCADE,
    "relationship" VARCHAR(50) NOT NULL DEFAULT 'parent'
                       CHECK ("relationship" IN ('father','mother','guardian','other')),
    "is_primary"   BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "student_parent_links_unique" UNIQUE ("student_id", "parent_id")
);

-- ============================================================
-- SUBJECTS
-- ============================================================

CREATE TABLE "school_subjects" (
    "id"                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"                 UUID         NOT NULL REFERENCES "schools"("id")                ON DELETE CASCADE,
    "school_education_level_id" UUID         REFERENCES "school_education_levels"("id")          ON DELETE SET NULL,
    "name"                      VARCHAR(191) NOT NULL,
    "code"                      VARCHAR(50),
    "category"                  VARCHAR(50)  NOT NULL DEFAULT 'core'
                                    CHECK ("category" IN ('core','elective','extracurricular','specialized','vocational')),
    "deleted_at"                TIMESTAMPTZ,
    "created_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- SCHEDULING
-- ============================================================

-- Shift-level schedule templates (e.g. "morning shift for students of Kelas 1")
CREATE TABLE "school_schedules" (
    "id"                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"                 UUID NOT NULL REFERENCES "schools"("id")               ON DELETE CASCADE,
    "school_education_level_id" UUID REFERENCES "school_education_levels"("id")         ON DELETE SET NULL,
    "school_classes_id"         UUID REFERENCES "school_classes"("id")                  ON DELETE SET NULL,
    "school_sub_classes_id"     UUID REFERENCES "school_sub_classes"("id")              ON DELETE SET NULL,
    "role_id"                   UUID REFERENCES "roles"("id")                           ON DELETE SET NULL,
    "shift_type"                VARCHAR(50) CHECK ("shift_type" IN ('morning','afternoon','full_day')),
    "start_time"                TIME,
    "end_time"                  TIME,
    "late_tolerance"            TIME,
    "is_active"                 BOOLEAN     NOT NULL DEFAULT TRUE,
    "deleted_at"                TIMESTAMPTZ,
    "created_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "school_schedule_days" (
    "id"                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_schedule_id" UUID        NOT NULL REFERENCES "school_schedules"("id") ON DELETE CASCADE,
    "day"                VARCHAR(20) NOT NULL
                             CHECK ("day" IN ('monday','tuesday','wednesday','thursday','friday','saturday','sunday')),
    "created_at"         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT "school_schedule_days_unique" UNIQUE ("school_schedule_id", "day")
);

-- Individual class sessions (teacher + classroom + subject + date)
CREATE TABLE "class_schedules" (
    "id"                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"              UUID NOT NULL REFERENCES "schools"("id")           ON DELETE CASCADE,
    "school_classroom_id"    UUID NOT NULL REFERENCES "school_classrooms"("id") ON DELETE CASCADE,
    "school_subject_id"      UUID NOT NULL REFERENCES "school_subjects"("id")   ON DELETE CASCADE,
    "teacher_id"             UUID NOT NULL REFERENCES "users"("id")             ON DELETE CASCADE,
    "date"                   DATE        NOT NULL,
    "start_time"             TIME        NOT NULL,
    "end_time"               TIME        NOT NULL,
    "teacher_attendance_time" TIMESTAMPTZ,
    "status"                 VARCHAR(50) NOT NULL DEFAULT 'scheduled'
                                 CHECK ("status" IN ('scheduled','ongoing','completed','cancelled')),
    "deleted_at"             TIMESTAMPTZ,
    "created_at"             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Per-student attendance within a class session
CREATE TABLE "class_schedule_students" (
    "id"                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"              UUID NOT NULL REFERENCES "schools"("id")          ON DELETE CASCADE,
    "class_schedule_id"      UUID NOT NULL REFERENCES "class_schedules"("id") ON DELETE CASCADE,
    "student_id"             UUID NOT NULL REFERENCES "users"("id")           ON DELETE CASCADE,
    "type"                   VARCHAR(20) CHECK ("type" IN ('check_in','check_out')),
    "photo_path"             VARCHAR(191),
    "note"                   TEXT,
    "student_attendance_time" TIMESTAMPTZ,
    "status"                 VARCHAR(50) NOT NULL DEFAULT 'absent'
                                 CHECK ("status" IN ('present','absent','late','excused')),
    "deleted_at"             TIMESTAMPTZ,
    "created_at"             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ATTENDANCE (general check-in/out via QR or face recognition)
-- ============================================================

CREATE TABLE "school_attendances" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"  UUID        NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "user_id"    UUID        NOT NULL REFERENCES "users"("id")   ON DELETE CASCADE,
    "type"       VARCHAR(20) CHECK ("type" IN ('check_in','check_out')),
    "photo_path" VARCHAR(191),
    "date"       DATE        NOT NULL,
    "time"       TIME,
    "status"     VARCHAR(50) NOT NULL
                     CHECK ("status" IN ('present','absent','late','excused','early','sick')),
    "note"       TEXT,
    "deleted_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ACADEMIC RECORDS
-- ============================================================

-- Student enrollment per classroom per academic year
CREATE TABLE "student_studies" (
    "id"                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"               UUID NOT NULL REFERENCES "schools"("id")               ON DELETE CASCADE,
    "student_id"              UUID NOT NULL REFERENCES "users"("id")                 ON DELETE CASCADE,
    "school_classroom_id"     UUID NOT NULL REFERENCES "school_classrooms"("id")     ON DELETE CASCADE,
    "school_academic_year_id" UUID NOT NULL REFERENCES "school_academic_years"("id") ON DELETE CASCADE,
    "school_class_id"         UUID REFERENCES "school_classes"("id")                 ON DELETE SET NULL,
    "school_sub_class_id"     UUID REFERENCES "school_sub_classes"("id")             ON DELETE SET NULL,
    "homeroom_teacher_id"     UUID REFERENCES "users"("id")                          ON DELETE SET NULL,
    "status"                  VARCHAR(50) NOT NULL
                                  CHECK ("status" IN ('active','inactive','graduated','transferred')),
    "enrollment_date"         DATE        NOT NULL,
    "deleted_at"              TIMESTAMPTZ,
    "created_at"              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Student grade promotion records
CREATE TABLE "student_promotions" (
    "id"                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"               UUID NOT NULL REFERENCES "schools"("id")               ON DELETE CASCADE,
    "student_id"              UUID NOT NULL REFERENCES "users"("id")                 ON DELETE CASCADE,
    "from_classroom_id"       UUID NOT NULL REFERENCES "school_classrooms"("id")     ON DELETE CASCADE,
    "to_classroom_id"         UUID NOT NULL REFERENCES "school_classrooms"("id")     ON DELETE CASCADE,
    "school_academic_year_id" UUID NOT NULL REFERENCES "school_academic_years"("id") ON DELETE CASCADE,
    "promotion_date"          DATE        NOT NULL,
    "status"                  VARCHAR(50) NOT NULL
                                  CHECK ("status" IN ('promoted','retained','transferred','rejected','eligible')),
    "notes"                   TEXT,
    "deleted_at"              TIMESTAMPTZ,
    "created_at"              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- CBT (ONLINE EXAM)
-- ============================================================

CREATE TABLE "cbt_categories" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"  UUID         NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "value"      VARCHAR(191) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "cbt_question_categories" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"  UUID         NOT NULL REFERENCES "schools"("id") ON DELETE CASCADE,
    "value"      VARCHAR(191) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "cbt" (
    "id"               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"        UUID NOT NULL REFERENCES "schools"("id")           ON DELETE CASCADE,
    "creator_id"       UUID REFERENCES "users"("id")                       ON DELETE SET NULL,
    "cbt_category_id"  UUID REFERENCES "cbt_categories"("id")              ON DELETE SET NULL,
    "subject_id"       UUID REFERENCES "school_subjects"("id")             ON DELETE SET NULL,
    "class_id"         UUID REFERENCES "school_classes"("id")              ON DELETE SET NULL,
    "title"            VARCHAR(191) NOT NULL,
    "description"      TEXT,
    "status"           VARCHAR(50)  NOT NULL DEFAULT 'draft'
                           CHECK ("status" IN ('published','pending','draft','completed','archived','deleted')),
    "start_time"       TIMESTAMPTZ,
    "end_time"         TIMESTAMPTZ,
    "duration_minutes" INT,
    "deleted_at"       TIMESTAMPTZ,
    "created_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "cbt_question" (
    "id"                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"                 UUID NOT NULL REFERENCES "schools"("id")                 ON DELETE CASCADE,
    "cbt_id"                    UUID REFERENCES "cbt"("id")                               ON DELETE CASCADE,
    "cbt_question_category_id"  UUID NOT NULL REFERENCES "cbt_question_categories"("id") ON DELETE CASCADE,
    "value"                     TEXT               NOT NULL,
    "score_weight"              DOUBLE PRECISION   NOT NULL DEFAULT 1,
    "created_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "cbt_options" (
    "id"              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"       UUID NOT NULL REFERENCES "schools"("id")       ON DELETE CASCADE,
    "cbt_question_id" UUID NOT NULL REFERENCES "cbt_question"("id") ON DELETE CASCADE,
    "value"           TEXT    NOT NULL,
    "is_correct"      BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Correct answer key for a question (essay or option reference)
CREATE TABLE "cbt_answers" (
    "id"              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"       UUID NOT NULL REFERENCES "schools"("id")       ON DELETE CASCADE,
    "cbt_question_id" UUID NOT NULL REFERENCES "cbt_question"("id") ON DELETE CASCADE,
    "cbt_option_id"   UUID REFERENCES "cbt_options"("id")            ON DELETE SET NULL,
    "value"           TEXT NOT NULL,
    "created_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One campaign = one student's exam attempt
-- status values translated from Indonesian originals:
--   'Akan dimulai' → 'upcoming'
--   'Berlangsung'  → 'ongoing'
--   'Selesai'      → 'completed'
--   'Dibatalkan'   → 'cancelled'
CREATE TABLE "cbt_campaigns" (
    "id"               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"        UUID NOT NULL REFERENCES "schools"("id")      ON DELETE CASCADE,
    "cbt_id"           UUID REFERENCES "cbt"("id")                    ON DELETE SET NULL,
    "title"            VARCHAR(191) NOT NULL,
    "category"         VARCHAR(191) NOT NULL,
    "subject_id"       UUID REFERENCES "school_subjects"("id")        ON DELETE SET NULL,
    "class_id"         UUID REFERENCES "school_classes"("id")         ON DELETE SET NULL,
    "teacher_id"       UUID REFERENCES "users"("id")                  ON DELETE SET NULL,
    "student_id"       UUID REFERENCES "users"("id")                  ON DELETE SET NULL,
    "start_time"       TIMESTAMPTZ,
    "end_time"         TIMESTAMPTZ,
    "status"           VARCHAR(50)  NOT NULL DEFAULT 'upcoming'
                           CHECK ("status" IN ('upcoming','ongoing','completed','cancelled')),
    "exam_score"       DOUBLE PRECISION,
    "duration_minutes" INT,
    "created_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Snapshot of questions at the time the campaign was started
CREATE TABLE "cbt_campaign_questions" (
    "id"              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"       UUID NOT NULL REFERENCES "schools"("id")       ON DELETE CASCADE,
    "campaign_id"     UUID NOT NULL REFERENCES "cbt_campaigns"("id") ON DELETE CASCADE,
    "cbt_question_id" UUID NOT NULL REFERENCES "cbt_question"("id") ON DELETE CASCADE,
    "category"        VARCHAR(191)     NOT NULL,
    "value"           TEXT             NOT NULL,
    "score_weight"    DOUBLE PRECISION NOT NULL DEFAULT 1,
    "created_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Snapshot of options at the time the campaign was started
CREATE TABLE "cbt_campaign_options" (
    "id"                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"                UUID NOT NULL REFERENCES "schools"("id")                   ON DELETE CASCADE,
    "cbt_campaign_question_id" UUID NOT NULL REFERENCES "cbt_campaign_questions"("id")    ON DELETE CASCADE,
    "cbt_option_id"            UUID REFERENCES "cbt_options"("id")                         ON DELETE SET NULL,
    "value"                    TEXT    NOT NULL,
    "is_correct"               BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at"               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Student's submitted answers
CREATE TABLE "cbt_campaign_answers" (
    "id"                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"                UUID NOT NULL REFERENCES "schools"("id")                ON DELETE CASCADE,
    "cbt_campaign_question_id" UUID NOT NULL REFERENCES "cbt_campaign_questions"("id") ON DELETE CASCADE,
    "cbt_campaign_option_id"   UUID REFERENCES "cbt_campaign_options"("id")             ON DELETE SET NULL,
    "cbt_answer_id"            UUID REFERENCES "cbt_answers"("id")                      ON DELETE SET NULL,
    "value"                    TEXT,
    "source_answer_value"      TEXT,
    "score"                    DOUBLE PRECISION,
    "created_at"               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================

CREATE TABLE "notifications" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"  UUID REFERENCES "schools"("id") ON DELETE CASCADE,
    "user_id"    UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "type"       VARCHAR(191) NOT NULL,
    "title"      VARCHAR(191),
    "body"       TEXT,
    "data"       JSONB        NOT NULL DEFAULT '{}',
    "read_at"    TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- users
CREATE INDEX "idx_users_email"    ON "users"("email");
CREATE INDEX "idx_users_username" ON "users"("username");

-- refresh_tokens
CREATE INDEX "idx_refresh_tokens_user_id"   ON "refresh_tokens"("user_id");
CREATE INDEX "idx_refresh_tokens_expires_at" ON "refresh_tokens"("expires_at");

-- api_keys
CREATE INDEX "idx_api_keys_user_id" ON "api_keys"("user_id");

-- schools
CREATE INDEX "idx_schools_status" ON "schools"("status");

-- school_users
CREATE INDEX "idx_school_users_school_id" ON "school_users"("school_id");
CREATE INDEX "idx_school_users_user_id"   ON "school_users"("user_id");

-- subscriptions
CREATE INDEX "idx_subscriptions_school_id" ON "subscriptions"("school_id");
CREATE INDEX "idx_subscriptions_status"    ON "subscriptions"("status");

-- academic structure
CREATE INDEX "idx_school_academic_years_school_id"      ON "school_academic_years"("school_id");
CREATE INDEX "idx_school_education_levels_school_id"    ON "school_education_levels"("school_id");
CREATE INDEX "idx_school_classes_school_id"             ON "school_classes"("school_id");
CREATE INDEX "idx_school_classes_education_level_id"    ON "school_classes"("school_education_level_id");
CREATE INDEX "idx_school_sub_classes_school_id"         ON "school_sub_classes"("school_id");
CREATE INDEX "idx_school_sub_classes_class_id"          ON "school_sub_classes"("school_class_id");
CREATE INDEX "idx_school_classrooms_school_id"          ON "school_classrooms"("school_id");
CREATE INDEX "idx_school_classrooms_academic_year_id"   ON "school_classrooms"("school_academic_year_id");
CREATE INDEX "idx_school_classrooms_class_id"           ON "school_classrooms"("school_class_id");

-- profiles
CREATE INDEX "idx_teacher_profiles_school_id"     ON "teacher_profiles"("school_id");
CREATE INDEX "idx_teacher_profiles_user_id"       ON "teacher_profiles"("user_id");
CREATE INDEX "idx_headmaster_profiles_school_id"  ON "headmaster_profiles"("school_id");
CREATE INDEX "idx_headmaster_profiles_user_id"    ON "headmaster_profiles"("user_id");
CREATE INDEX "idx_staff_profiles_school_id"       ON "staff_profiles"("school_id");
CREATE INDEX "idx_parent_profiles_school_id"      ON "parent_profiles"("school_id");
CREATE INDEX "idx_student_profiles_school_id"     ON "student_profiles"("school_id");
CREATE INDEX "idx_student_profiles_user_id"       ON "student_profiles"("user_id");
CREATE INDEX "idx_student_profiles_nis"           ON "student_profiles"("nis");
CREATE INDEX "idx_student_profiles_nisn"          ON "student_profiles"("nisn");
CREATE INDEX "idx_student_parent_links_student"   ON "student_parent_links"("student_id");
CREATE INDEX "idx_student_parent_links_parent"    ON "student_parent_links"("parent_id");
CREATE INDEX "idx_student_parent_links_school_id" ON "student_parent_links"("school_id");

-- subjects
CREATE INDEX "idx_school_subjects_school_id"    ON "school_subjects"("school_id");
CREATE INDEX "idx_school_subjects_level_id"     ON "school_subjects"("school_education_level_id");

-- schedules
CREATE INDEX "idx_school_schedules_school_id"      ON "school_schedules"("school_id");
CREATE INDEX "idx_class_schedules_school_id"       ON "class_schedules"("school_id");
CREATE INDEX "idx_class_schedules_classroom_id"    ON "class_schedules"("school_classroom_id");
CREATE INDEX "idx_class_schedules_teacher_id"      ON "class_schedules"("teacher_id");
CREATE INDEX "idx_class_schedules_date"            ON "class_schedules"("date");
CREATE INDEX "idx_class_schedule_students_schedule" ON "class_schedule_students"("class_schedule_id");
CREATE INDEX "idx_class_schedule_students_student"  ON "class_schedule_students"("student_id");
CREATE INDEX "idx_class_schedule_students_school"   ON "class_schedule_students"("school_id");

-- attendance
CREATE INDEX "idx_school_attendances_school_id" ON "school_attendances"("school_id");
CREATE INDEX "idx_school_attendances_user_id"   ON "school_attendances"("user_id");
CREATE INDEX "idx_school_attendances_date"      ON "school_attendances"("date");
CREATE INDEX "idx_school_attendances_school_date" ON "school_attendances"("school_id", "date");

-- academic records
CREATE INDEX "idx_student_studies_school_id"       ON "student_studies"("school_id");
CREATE INDEX "idx_student_studies_student_id"      ON "student_studies"("student_id");
CREATE INDEX "idx_student_studies_classroom_id"    ON "student_studies"("school_classroom_id");
CREATE INDEX "idx_student_studies_academic_year"   ON "student_studies"("school_academic_year_id");
CREATE INDEX "idx_student_promotions_school_id"    ON "student_promotions"("school_id");
CREATE INDEX "idx_student_promotions_student_id"   ON "student_promotions"("student_id");

-- CBT
CREATE INDEX "idx_cbt_school_id"                      ON "cbt"("school_id");
CREATE INDEX "idx_cbt_creator_id"                     ON "cbt"("creator_id");
CREATE INDEX "idx_cbt_status"                         ON "cbt"("status");
CREATE INDEX "idx_cbt_question_school_id"             ON "cbt_question"("school_id");
CREATE INDEX "idx_cbt_question_cbt_id"                ON "cbt_question"("cbt_id");
CREATE INDEX "idx_cbt_campaigns_school_id"            ON "cbt_campaigns"("school_id");
CREATE INDEX "idx_cbt_campaigns_student_id"           ON "cbt_campaigns"("student_id");
CREATE INDEX "idx_cbt_campaigns_cbt_id"               ON "cbt_campaigns"("cbt_id");
CREATE INDEX "idx_cbt_campaigns_status"               ON "cbt_campaigns"("status");
CREATE INDEX "idx_cbt_campaign_questions_campaign_id" ON "cbt_campaign_questions"("campaign_id");
CREATE INDEX "idx_cbt_campaign_answers_campaign_q_id" ON "cbt_campaign_answers"("cbt_campaign_question_id");

-- notifications
CREATE INDEX "idx_notifications_user_id"  ON "notifications"("user_id");
CREATE INDEX "idx_notifications_school_id" ON "notifications"("school_id");
CREATE INDEX "idx_notifications_read_at"  ON "notifications"("read_at");

-- ============================================================
-- SEED DATA — Roles
-- ============================================================

INSERT INTO "roles" ("name", "guard_name", "display_name", "description") VALUES
    ('superadmin',    'web', 'Super Admin',    'Full system access, bypasses tenant scoping'),
    ('admin_sekolah', 'web', 'Admin Sekolah',  'Platform-level school manager (renamed from manager)'),
    ('kepala_sekolah','web', 'Kepala Sekolah', 'School principal, school-scoped'),
    ('guru',          'web', 'Guru',           'Teacher, school-scoped'),
    ('staff',         'web', 'Staff',          'Administrative staff, school-scoped'),
    ('orangtua',      'web', 'Orang Tua',      'Parent or guardian, school-scoped'),
    ('siswa',         'web', 'Siswa',          'Student, school-scoped');

-- ============================================================
-- SEED DATA — Default Plans
-- ============================================================

INSERT INTO "plans" ("name", "description", "features", "monthly_price", "yearly_price", "active", "is_default") VALUES
    ('Trial',    'Plan percobaan gratis untuk semua sekolah baru.',
     '["Maks 50 siswa","1 sekolah","Absensi dasar","CBT dasar"]',
     0, 0, TRUE, TRUE),
    ('Basic',    'Paket dasar untuk sekolah kecil.',
     '["Maks 200 siswa","1 sekolah","Absensi lengkap","CBT lengkap","Notifikasi WhatsApp"]',
     500000, 5000000, TRUE, FALSE),
    ('Standard', 'Paket standar untuk sekolah menengah.',
     '["Maks 500 siswa","1 sekolah","Semua fitur Basic","Laporan lanjutan"]',
     1000000, 10000000, TRUE, FALSE),
    ('Premium',  'Paket premium tanpa batas untuk sekolah besar.',
     '["Siswa tidak terbatas","1 sekolah","Semua fitur","Prioritas support"]',
     1500000, 15000000, TRUE, FALSE);
