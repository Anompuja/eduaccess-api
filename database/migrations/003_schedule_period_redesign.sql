-- Migration 003: Redesign school_schedules as lesson periods per day
-- Run manually in Supabase SQL editor

-- ── school_schedules: drop old columns, add period columns ────────────────────

ALTER TABLE "school_schedules"
    DROP COLUMN IF EXISTS "school_education_level_id",
    DROP COLUMN IF EXISTS "school_classes_id",
    DROP COLUMN IF EXISTS "school_sub_classes_id",
    DROP COLUMN IF EXISTS "role_id",
    DROP COLUMN IF EXISTS "shift_type",
    DROP COLUMN IF EXISTS "late_tolerance",
    DROP COLUMN IF EXISTS "is_active";

-- Ensure start_time/end_time are VARCHAR so GORM string mapping works correctly
ALTER TABLE "school_schedules"
    ALTER COLUMN "start_time" TYPE VARCHAR(10) USING start_time::text,
    ALTER COLUMN "end_time"   TYPE VARCHAR(10) USING end_time::text;

ALTER TABLE "school_schedules"
    ADD COLUMN "day_of_week"    VARCHAR(20)  NOT NULL DEFAULT 'monday'
        CHECK ("day_of_week" IN ('monday','tuesday','wednesday','thursday','friday','saturday','sunday')),
    ADD COLUMN "period_number"  INT          NOT NULL DEFAULT 1,
    ADD COLUMN "label"          VARCHAR(50)  NOT NULL DEFAULT '',
    ADD COLUMN "is_break"       BOOLEAN      NOT NULL DEFAULT FALSE;

ALTER TABLE "school_schedules"
    ALTER COLUMN "day_of_week" DROP DEFAULT,
    ALTER COLUMN "period_number" DROP DEFAULT,
    ALTER COLUMN "label" DROP DEFAULT;

-- Unique constraint: one period number per day per school
ALTER TABLE "school_schedules"
    ADD CONSTRAINT "school_schedules_school_day_period_unique"
    UNIQUE ("school_id", "day_of_week", "period_number");

-- Drop the old school_schedule_days table (no longer needed)
DROP TABLE IF EXISTS "school_schedule_days";

-- ── class_schedules: add optional period FK references ────────────────────────

ALTER TABLE "class_schedules"
    ADD COLUMN "start_period_id" UUID REFERENCES "school_schedules"("id") ON DELETE SET NULL,
    ADD COLUMN "end_period_id"   UUID REFERENCES "school_schedules"("id") ON DELETE SET NULL;

CREATE INDEX "idx_class_schedules_start_period" ON "class_schedules"("start_period_id");
CREATE INDEX "idx_class_schedules_end_period" ON "class_schedules"("end_period_id");
