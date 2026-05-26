ALTER TABLE "client_profile"
    ADD COLUMN IF NOT EXISTS paid_orders_count INT NOT NULL DEFAULT 0
        CHECK (paid_orders_count >= 0);

CREATE TABLE "achievement" (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
        CHECK (char_length(code) <= 64),
    title TEXT NOT NULL
        CHECK (char_length(title) >= 1 AND char_length(title) <= 64),
    description TEXT NOT NULL
        CHECK (char_length(description) <= 256),
    icon TEXT NOT NULL
        CHECK (char_length(icon) <= 16),
    sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE "account_achievement" (
    account_id BIGINT NOT NULL,
    achievement_id BIGINT NOT NULL,
    awarded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (account_id, achievement_id),

    CONSTRAINT fk_account_achievement_user
        FOREIGN KEY (account_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_account_achievement_achievement
        FOREIGN KEY (achievement_id)
        REFERENCES "achievement"(id)
        ON DELETE CASCADE
);

CREATE TABLE "account_restaurant" (
    account_id BIGINT NOT NULL,
    restaurant_id BIGINT NOT NULL,
    first_order_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (account_id, restaurant_id),

    CONSTRAINT fk_account_restaurant_user
        FOREIGN KEY (account_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE
);

INSERT INTO "achievement" (code, title, description, icon, sort_order) VALUES
    ('first_order', 'Первый заказ', 'Оплатите первый заказ на FoodCourt', '🍔', 1),
    ('five_orders', 'Постоянный клиент', 'Оплатите 5 заказов', '🏅', 2),
    ('gourmand_three', 'Гурман', 'Попробуйте еду из 3 разных ресторанов', '🍽️', 3),
    ('first_spin', 'Испытатель удачи', 'Запустите Колесо Пиццули в первый раз', '🌀', 4);
    ('lucky_wheel_winner', 'Любимчик Пиццули', 'Выиграйте эксклюзивную награду в Колесе Пиццули', '🎡', 5),
    ('streak_six', 'Постоянство', 'Поддерживайте серию заказов 6 недель подряд', '🔥', 6)
ON CONFLICT (code) DO NOTHING;

UPDATE "client_profile"
SET streak_count = 3,
    last_order_date = date_trunc('week', NOW())
WHERE account_id = (SELECT id FROM "user" WHERE email = 'anna@foodcourt.fun');
