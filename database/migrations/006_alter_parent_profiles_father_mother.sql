-- Migration: Alter parent_profiles table to separate father and mother details
-- Date: 2026-06-08

ALTER TABLE "parent_profiles" 
    ADD COLUMN "father_name" VARCHAR(191),
    ADD COLUMN "mother_name" VARCHAR(191),
    ADD COLUMN "father_religion" VARCHAR(100),
    ADD COLUMN "mother_religion" VARCHAR(100);

-- Drop the old religion column which is replaced by father_religion and mother_religion
ALTER TABLE "parent_profiles" 
    DROP COLUMN IF EXISTS "religion";
