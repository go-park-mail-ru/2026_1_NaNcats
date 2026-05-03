-- Добавляем дополнительные категории (часть уже есть из 000002)
INSERT INTO "category" (name) VALUES
('Популярное'),
('Десерты'),
('Аптеки'),
('Цветы'),
('Здоровье'),
('Стейки'),
('Паста'),
('Азиатская кухня'),
('Морепродукты'),
('Бизнес-ланч'),
('Веганское'),
('Доставка 24/7'),
('Блины'),
('Вок'),
('Грузинская кухня'),
('Украинская кухня'),
('Домашняя кухня'),
('Хлеб и выпечка'),
('Торты на заказ'),
('Мороженое'),
('Полезные перекусы'),
('Смузи'),
('Чай'),
('Детское меню'),
('Вечеринка'),
('Кейтеринг'),
('Алкоголь'),
('Пиво'),
('Коктейли'),
('Супы'),
('Салаты')
ON CONFLICT (name) DO NOTHING;

-- Привязываем рестораны к категориям
-- Фастфуд
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Вкусно - и точка', 'KFC', 'Subway', 'Pinsky Go')
  AND c.name = 'Фастфуд'
ON CONFLICT DO NOTHING;

-- Пицца
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Папа Джонс', 'FoodBand', 'Villa Pasta', 'Империя Пиццы', 'Tutta La Vita')
  AND c.name = 'Пицца'
ON CONFLICT DO NOTHING;

-- Суши
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('FoodBand', 'Izumi', 'Раменная Ru-Rik')
  AND c.name = 'Суши'
ON CONFLICT DO NOTHING;

-- Бургеры
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Ketch Up', 'Вкусно - и точка', 'El Chapo Burgers Tacos&Burritos', 'Калифорния Дайнер')
  AND c.name = 'Бургеры'
ON CONFLICT DO NOTHING;

-- Завтраки
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Братья Караваевы', 'Anna')
  AND c.name = 'Завтраки'
ON CONFLICT DO NOTHING;

-- Кофе
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Anna', 'Братья Караваевы')
  AND c.name = 'Кофе'
ON CONFLICT DO NOTHING;

-- Русская кухня
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Крошка Картошка', 'Лепим и Варим', 'Аист')
  AND c.name = 'Русская'
ON CONFLICT DO NOTHING;

-- Китайская кухня
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Китайские новости')
  AND c.name = 'Китайская'
ON CONFLICT DO NOTHING;

-- Азиатская кухня
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Раменная Ru-Rik', 'Izumi', 'Такахули', 'Eshak', 'Китайские новости')
  AND c.name = 'Азиатская кухня'
ON CONFLICT DO NOTHING;

-- Стейки
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Arium Grill', 'Moro')
  AND c.name = 'Стейки'
ON CONFLICT DO NOTHING;

-- Морепродукты
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Честная Рыба', 'Moro')
  AND c.name = 'Морепродукты'
ON CONFLICT DO NOTHING;

-- Грузинская кухня
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('DiDi')
  AND c.name = 'Грузинская кухня'
ON CONFLICT DO NOTHING;

-- Паста
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Villa Pasta', 'Tutta La Vita')
  AND c.name = 'Паста'
ON CONFLICT DO NOTHING;

-- Домашняя кухня
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.name IN ('Братья Караваевы', 'Лепим и Варим', 'Крошка Картошка')
  AND c.name = 'Домашняя кухня'
ON CONFLICT DO NOTHING;

-- Популярное (все рестораны с promotion_tier >= 2)
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
SELECT rb.id, c.id
FROM "restaurant_brand" rb, "category" c
WHERE rb.promotion_tier >= 2
  AND c.name = 'Популярное'
ON CONFLICT DO NOTHING;

-- emoji-колонка теперь часть init_schema (000001). Для совместимости с
-- БД, где init_schema уже была применена без emoji, добавляем колонку
-- идемпотентно — на свежих БД это no-op.
ALTER TABLE "category" ADD COLUMN IF NOT EXISTS emoji TEXT NOT NULL DEFAULT '';

UPDATE "category" SET emoji = '🔥' WHERE name = 'Популярное';
UPDATE "category" SET emoji = '🍕' WHERE name = 'Пицца';
UPDATE "category" SET emoji = '🍣' WHERE name = 'Суши';
UPDATE "category" SET emoji = '🍔' WHERE name = 'Бургеры';
UPDATE "category" SET emoji = '🍰' WHERE name = 'Десерты';
UPDATE "category" SET emoji = '💊' WHERE name = 'Аптеки';
UPDATE "category" SET emoji = '💐' WHERE name = 'Цветы';
UPDATE "category" SET emoji = '🍳' WHERE name = 'Завтраки';
UPDATE "category" SET emoji = '🥦' WHERE name = 'Здоровье';
UPDATE "category" SET emoji = '☕' WHERE name = 'Кофе';
UPDATE "category" SET emoji = '🥩' WHERE name = 'Стейки';
UPDATE "category" SET emoji = '🍝' WHERE name = 'Паста';
UPDATE "category" SET emoji = '🥢' WHERE name = 'Азиатская кухня';
UPDATE "category" SET emoji = '🦞' WHERE name = 'Морепродукты';
UPDATE "category" SET emoji = '🍱' WHERE name = 'Бизнес-ланч';
UPDATE "category" SET emoji = '🌱' WHERE name = 'Веганское';
UPDATE "category" SET emoji = '⏰' WHERE name = 'Доставка 24/7';
UPDATE "category" SET emoji = '🥞' WHERE name = 'Блины';
UPDATE "category" SET emoji = '🥡' WHERE name = 'Вок';
UPDATE "category" SET emoji = '🥠' WHERE name = 'Китайская';
UPDATE "category" SET emoji = '🥙' WHERE name = 'Грузинская кухня';
UPDATE "category" SET emoji = '🍲' WHERE name = 'Украинская кухня';
UPDATE "category" SET emoji = '🏠' WHERE name = 'Домашняя кухня';
UPDATE "category" SET emoji = '🍟' WHERE name = 'Фастфуд';
UPDATE "category" SET emoji = '🥖' WHERE name = 'Хлеб и выпечка';
UPDATE "category" SET emoji = '🎂' WHERE name = 'Торты на заказ';
UPDATE "category" SET emoji = '🍦' WHERE name = 'Мороженое';
UPDATE "category" SET emoji = '🥜' WHERE name = 'Полезные перекусы';
UPDATE "category" SET emoji = '🥤' WHERE name = 'Смузи';
UPDATE "category" SET emoji = '🍵' WHERE name = 'Чай';
UPDATE "category" SET emoji = '🧸' WHERE name = 'Детское меню';
UPDATE "category" SET emoji = '🎉' WHERE name = 'Вечеринка';
UPDATE "category" SET emoji = '🍽️' WHERE name = 'Кейтеринг';
UPDATE "category" SET emoji = '🍷' WHERE name = 'Алкоголь';
UPDATE "category" SET emoji = '🍺' WHERE name = 'Пиво';
UPDATE "category" SET emoji = '🍹' WHERE name = 'Коктейли';
UPDATE "category" SET emoji = '🥣' WHERE name = 'Супы';
UPDATE "category" SET emoji = '🥗' WHERE name = 'Салаты';
UPDATE "category" SET emoji = '🇷🇺' WHERE name = 'Русская';
UPDATE "category" SET emoji = '🇨🇳' WHERE name = 'Китайская';
