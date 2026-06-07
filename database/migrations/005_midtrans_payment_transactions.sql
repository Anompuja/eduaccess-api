CREATE TABLE IF NOT EXISTS "payment_transactions" (
    "id"                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "school_id"                 UUID NOT NULL REFERENCES "schools"("id"),
    "plan_id"                   UUID NOT NULL REFERENCES "plans"("id"),
    "created_by_user_id"        UUID NOT NULL REFERENCES "users"("id"),
    "activated_subscription_id" UUID REFERENCES "subscriptions"("id"),
    "status"                    VARCHAR(50) NOT NULL CHECK ("status" IN ('pending','paid','failed','expired','cancelled')),
    "cycle"                     VARCHAR(50) NOT NULL CHECK ("cycle" IN ('month','year')),
    "amount"                    BIGINT NOT NULL DEFAULT 0,
    "currency"                  VARCHAR(10) NOT NULL DEFAULT 'IDR',
    "provider"                  VARCHAR(50) NOT NULL DEFAULT 'midtrans',
    "provider_order_id"         VARCHAR(191) NOT NULL,
    "provider_transaction_id"   VARCHAR(191),
    "provider_snap_token"       VARCHAR(191),
    "provider_redirect_url"     TEXT,
    "payment_type"              VARCHAR(100),
    "transaction_status"        VARCHAR(100),
    "fraud_status"              VARCHAR(100),
    "raw_notification"          JSONB NOT NULL DEFAULT '{}'::jsonb,
    "paid_at"                   TIMESTAMPTZ,
    "expires_at"                TIMESTAMPTZ,
    "created_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT "payment_transactions_provider_order_id_unique" UNIQUE ("provider_order_id")
);

CREATE INDEX IF NOT EXISTS "idx_payment_transactions_school_id" ON "payment_transactions"("school_id");
CREATE INDEX IF NOT EXISTS "idx_payment_transactions_status" ON "payment_transactions"("status");
CREATE INDEX IF NOT EXISTS "idx_payment_transactions_provider" ON "payment_transactions"("provider");
