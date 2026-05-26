DROP TABLE IF EXISTS "account_restaurant";
DROP TABLE IF EXISTS "account_achievement";
DROP TABLE IF EXISTS "achievement";

ALTER TABLE "client_profile"
    DROP COLUMN IF EXISTS paid_orders_count;
