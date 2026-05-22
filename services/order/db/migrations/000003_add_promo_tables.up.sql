-- Promo code tables: codes, scoping, usage tracking.
-- Adapted from the monolith schema (db/migrations/000001_init_schema.up.sql).

CREATE TABLE IF NOT EXISTS "promocode" (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code              TEXT NOT NULL UNIQUE,
    title             TEXT NOT NULL DEFAULT '',
    discount_percent  INT,
    discount_amount   BIGINT,
    max_uses          INT DEFAULT 1,
    min_order_amount  BIGINT DEFAULT 0,
    is_global         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at        TIMESTAMPTZ,
    CONSTRAINT chk_discount CHECK (
        (discount_percent IS NOT NULL AND discount_amount IS NULL) OR
        (discount_percent IS NULL AND discount_amount IS NOT NULL)
    ),
    CONSTRAINT chk_discount_percent CHECK (discount_percent IS NULL OR (discount_percent > 0 AND discount_percent <= 100)),
    CONSTRAINT chk_discount_amount  CHECK (discount_amount IS NULL OR discount_amount > 0),
    CONSTRAINT chk_code_length      CHECK (length(code) BETWEEN 2 AND 50)
);

CREATE TABLE IF NOT EXISTS "promocode_restaurant_brand" (
    promocode_id        BIGINT NOT NULL REFERENCES "promocode"(id) ON DELETE CASCADE,
    restaurant_brand_id BIGINT NOT NULL,
    PRIMARY KEY (promocode_id, restaurant_brand_id)
);

CREATE TABLE IF NOT EXISTS "promocode_category" (
    promocode_id BIGINT NOT NULL REFERENCES "promocode"(id) ON DELETE CASCADE,
    category_id  BIGINT NOT NULL,
    PRIMARY KEY (promocode_id, category_id)
);

CREATE TABLE IF NOT EXISTS "promocode_usage" (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    promocode_id      BIGINT NOT NULL REFERENCES "promocode"(id) ON DELETE CASCADE,
    client_account_id BIGINT NOT NULL,
    order_id          BIGINT REFERENCES "order"(id) ON DELETE SET NULL,
    used_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Уникальность по (промокод, заказ): один заказ не может списать промокод
    -- дважды, но пользователь вправе применить промокод к разным заказам —
    -- ровно столько раз, сколько задано в promocode.max_uses.
    UNIQUE (promocode_id, order_id)
);

-- User-promo binding: users can "save" promo codes to their profile.
CREATE TABLE IF NOT EXISTS "user_promocode" (
    user_id      BIGINT NOT NULL,
    promocode_id BIGINT NOT NULL REFERENCES "promocode"(id) ON DELETE CASCADE,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, promocode_id)
);

CREATE INDEX IF NOT EXISTS idx_promocode_code ON "promocode"(code);
CREATE INDEX IF NOT EXISTS idx_promocode_usage_client ON "promocode_usage"(client_account_id);
CREATE INDEX IF NOT EXISTS idx_user_promocode_user ON "user_promocode"(user_id);
