-- Seed promo codes. Prices in micro-rubles (1₽ = 1000000).

INSERT INTO "promocode" (code, title, discount_percent, discount_amount, max_uses, min_order_amount, is_global, expires_at) VALUES
('WELCOME300', '300 ₽ на первый заказ',         NULL, 300000000,  1, 1200000000, TRUE,  '2026-05-31'::TIMESTAMPTZ),
('EPIC30',     'Скидка 30% на пиццу',           30,   NULL,       3, 800000000,  FALSE, '2026-05-20'::TIMESTAMPTZ),
('FREEBURGER', 'Бесплатная доставка',            NULL, 150000000,  1, 0,          FALSE, '2026-05-13'::TIMESTAMPTZ),
('VKUSNO20',   'Скидка 20% на первый заказ',     20,   NULL,       1, 0,          FALSE, '2026-06-30'::TIMESTAMPTZ),
('SUMMER15',   'Летняя скидка 15%',              15,   NULL,       5, 500000000,  TRUE,  '2026-08-31'::TIMESTAMPTZ),
('PIZZA500',   '500 ₽ на пиццу',                 NULL, 500000000,  2, 1500000000, FALSE, '2026-07-15'::TIMESTAMPTZ),
('FRIENDS',    '200 ₽ за приглашение друга',      NULL, 200000000,  10, 0,         TRUE,  '2026-12-31'::TIMESTAMPTZ),
('STUDENT',    'Студенческая скидка 10%',         10,   NULL,       99, 0,         TRUE,  '2026-09-01'::TIMESTAMPTZ),
('NEWUSER',    '150 ₽ новому пользователю',       NULL, 150000000,  1, 600000000,  TRUE,  '2026-06-15'::TIMESTAMPTZ),
('WEEKEND25',  'Скидка 25% в выходные',           25,   NULL,       3, 1000000000, TRUE,  '2026-07-01'::TIMESTAMPTZ)
ON CONFLICT (code) DO NOTHING;

-- Scope EPIC30 to Папа Джонс (brand_id=3)
INSERT INTO "promocode_restaurant_brand" (promocode_id, restaurant_brand_id)
SELECT id, 3 FROM "promocode" WHERE code = 'EPIC30'
ON CONFLICT DO NOTHING;

-- Scope PIZZA500 to Папа Джонс (3) and FoodBand (6)
INSERT INTO "promocode_restaurant_brand" (promocode_id, restaurant_brand_id)
SELECT id, 3 FROM "promocode" WHERE code = 'PIZZA500'
ON CONFLICT DO NOTHING;
INSERT INTO "promocode_restaurant_brand" (promocode_id, restaurant_brand_id)
SELECT id, 6 FROM "promocode" WHERE code = 'PIZZA500'
ON CONFLICT DO NOTHING;

-- Scope VKUSNO20 to Вкусно - и точка (brand_id=1)
INSERT INTO "promocode_restaurant_brand" (promocode_id, restaurant_brand_id)
SELECT id, 1 FROM "promocode" WHERE code = 'VKUSNO20'
ON CONFLICT DO NOTHING;

-- Scope FREEBURGER to brand_id=19 (Ketch Up — бургерная)
INSERT INTO "promocode_restaurant_brand" (promocode_id, restaurant_brand_id)
SELECT id, 19 FROM "promocode" WHERE code = 'FREEBURGER'
ON CONFLICT DO NOTHING;

-- Bind promos to user_id=3 (first client)
INSERT INTO "user_promocode" (user_id, promocode_id)
SELECT 3, id FROM "promocode" WHERE code IN ('WELCOME300', 'EPIC30', 'FREEBURGER')
ON CONFLICT DO NOTHING;
