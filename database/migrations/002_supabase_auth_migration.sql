-- ============================================================
-- EduAccess API — Supabase Auth Migration
-- File    : 002_supabase_auth_migration.sql
-- Purpose : Migrate authentication ownership to Supabase Auth,
--           add Storage bucket policies, Realtime notifications table,
--           and custom JWT claims hook.
-- ============================================================

-- ── 1. Remove columns now managed by Supabase Auth ──────────────────────────
ALTER TABLE "users"
    DROP COLUMN IF EXISTS "password",
    DROP COLUMN IF EXISTS "email_verified_at",
    DROP COLUMN IF EXISTS "verification_code",
    DROP COLUMN IF EXISTS "two_factor_secret",
    DROP COLUMN IF EXISTS "two_factor_recovery_codes",
    DROP COLUMN IF EXISTS "two_factor_confirmed_at",
    DROP COLUMN IF EXISTS "trial_ends_at";

-- ── 2. Drop refresh_tokens (Supabase Auth manages sessions) ─────────────────
DROP TABLE IF EXISTS "refresh_tokens";

-- ── 3. Clear existing user data and link users table to auth.users ───────────
-- NOTE: This truncates all user-derived data. Run only on a fresh/dev DB.
-- For production with existing users, migrate accounts to Supabase Auth first.
TRUNCATE TABLE
    "model_has_permissions",
    "model_has_roles",
    "school_users"
    RESTART IDENTITY CASCADE;

DELETE FROM "users";

-- Add FK from public.users → auth.users so Supabase owns the identity record.
-- The backend creates the auth.users row first (via Admin API), then inserts
-- the public.users profile row with the same UUID.
ALTER TABLE "users"
    ADD CONSTRAINT "users_id_auth_fk"
    FOREIGN KEY ("id") REFERENCES auth.users("id") ON DELETE CASCADE;

-- ── 4. Notifications table (Supabase Realtime enabled) ───────────────────────
-- Enable Realtime on this table via Supabase Dashboard → Database → Replication.
CREATE TABLE IF NOT EXISTS "notifications" (
    "id"         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"    UUID        NOT NULL REFERENCES auth.users("id") ON DELETE CASCADE,
    "school_id"  UUID        REFERENCES "schools"("id") ON DELETE CASCADE,
    "type"       VARCHAR(50) NOT NULL,   -- e.g. 'grade', 'attendance', 'announcement'
    "title"      VARCHAR(255) NOT NULL,
    "body"       TEXT,
    "read"       BOOLEAN     NOT NULL DEFAULT FALSE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_notifications_user_id"  ON "notifications" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_notifications_school_id" ON "notifications" ("school_id");

-- ── 5. Custom JWT claims hook ─────────────────────────────────────────────────
-- This function runs every time Supabase issues a JWT.
-- It injects school_id and app_role into the token payload so the backend
-- can authorise requests without extra DB lookups.
--
-- After running this migration, activate the hook in:
-- Supabase Dashboard → Authentication → Hooks → Custom Access Token
-- and point it to auth.custom_access_token_hook
--
CREATE OR REPLACE FUNCTION auth.custom_access_token_hook(event jsonb)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
DECLARE
    v_user_id   uuid  := (event->>'user_id')::uuid;
    v_school_id uuid;
    v_role      text;
BEGIN
    -- Prefer the most recently linked active school + role pair.
    SELECT su.school_id, r.name
    INTO   v_school_id, v_role
    FROM   school_users su
    JOIN   model_has_roles mhr ON mhr.user_id = su.user_id
    JOIN   roles r             ON r.id = mhr.role_id
    LEFT JOIN schools s        ON s.id = su.school_id
    WHERE  su.user_id   = v_user_id
      AND  su.deleted_at IS NULL
      AND  (s.id IS NULL OR s.deleted_at IS NULL)
    ORDER BY
        CASE WHEN s.status = 'active' THEN 0 ELSE 1 END,
        su.created_at DESC
    LIMIT 1;

    -- Inject custom claims. app_role avoids collision with Supabase's built-in
    -- 'role' claim which is used for Postgres RLS.
    RETURN jsonb_set(
        jsonb_set(
            event,
            '{claims,school_id}',
            COALESCE(to_jsonb(v_school_id::text), 'null'::jsonb)
        ),
        '{claims,app_role}',
        COALESCE(to_jsonb(v_role), 'null'::jsonb)
    );
END;
$$;

-- Grant execute permission to the Supabase auth service.
GRANT EXECUTE ON FUNCTION auth.custom_access_token_hook TO supabase_auth_admin;
