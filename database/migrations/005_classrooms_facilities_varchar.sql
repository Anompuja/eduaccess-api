-- Change facilities from JSONB to VARCHAR so plain string values are accepted.
-- Existing rows all have the default empty array, so USING '' is safe.
ALTER TABLE "school_classrooms"
  ALTER COLUMN "facilities" TYPE VARCHAR(500) USING '';

ALTER TABLE "school_classrooms"
  ALTER COLUMN "facilities" SET DEFAULT '';

ALTER TABLE "school_classrooms"
  ALTER COLUMN "facilities" SET NOT NULL;
