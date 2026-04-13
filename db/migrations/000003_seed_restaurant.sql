-- 1. Добавляем блюда во "Вкусно и точка"
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Вкусно - и точка' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url) VALUES
((SELECT id FROM rb), 'Гранд', 'Бифштекс из 100% говядины, сыр Чеддер, лук, маринованные огурчики', 185000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2b5cec7a-be0f-4ec3-81b8-702ee44856c5.webp'),
((SELECT id FROM rb), 'Чизбургер', 'Рубленый бифштекс и кусочек сыра Чеддер на карамелизованной булочке', 75000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8b545e9e-19b8-4cd2-b8d5-f419472760c6.webp'),
((SELECT id FROM rb), 'Картофель Фри', 'Вкусный, обжаренный в растительном фритюре и слегка подсоленный картофель', 95000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6977092f-17d4-4ea7-bb09-428693af4100.webp'),
((SELECT id FROM rb), 'Кола Добрый', 'Освежающий газированный напиток 0.5л', 89000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e63d0a07-976a-421c-b7e6-542975a6dfdc.webp'),
((SELECT id FROM rb), 'Наггетсы 6 шт', 'Сочные кусочки куриного филе в хрустящей панировке', 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d0c0825b-a568-463a-8123-50ca928501f8.webp'),
((SELECT id FROM rb), 'Наггетсы 4 шт', 'Сочные кусочки куриного филе в хрустящей панировке', 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/82874210-ac20-47e6-95b5-70f0a9e27f6c.webp'),
((SELECT id FROM rb), 'Наггетсы 20 шт', 'Сочные кусочки куриного филе в хрустящей панировке', 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f190d34e-c999-432b-82b5-f31cc354f206.webp'),
((SELECT id FROM rb), 'Картофель по-деревенски', 'Крутая картоха', 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a7cf471d-a0e0-4a5b-9059-73053c69f705.webp'),
((SELECT id FROM rb), 'Мороженое', 'Очень вкусно в теплую погоду', 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e5fa9081-88ba-47b3-81de-be59cf70a55a.webp'),
((SELECT id FROM rb), 'Влажая салфетка', 'Не запачкайтесь!', 1100000, 'https://nancats-bucket.storage.yandexcloud.net/foods/85a82632-7f54-4f71-84ce-f87b13329272.webp');

-- 2. Добавляем блюда в "Папа Джонс"
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Папа Джонс' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url) VALUES
((SELECT id FROM rb), 'Пицца Пепперони (30 см)', 'Пикантная пепперони, томатный соус и сыр моцарелла', 699000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bac8f917-2762-4e65-9e61-fbc6a4ef54da.webp'),
((SELECT id FROM rb), 'Пицца Мясная (30 см)', 'Бекон, пепперони, ветчина, томатный соус и моцарелла', 899000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/78d0499a-7c23-49d4-824b-84955f9e6225.webp'),
((SELECT id FROM rb), 'Сырные палочки', 'Хрустящие палочки с сыром и чесночным соусом', 299000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2b77e55a-a7a7-4be3-98b0-578e159f0214.webp'),
((SELECT id FROM rb), 'Пицца Маргарита (30 см)', 'Традиционная пицца с томатным соусом и увеличенной порцией моцареллы', 549000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d7444179-ef0c-4dac-817d-70eeb62e4ca7.webp');

-- 3. Добавляем блюда в "FoodBand"
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'FoodBand' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url) VALUES
((SELECT id FROM rb), 'Ролл Филадельфия', 'Лосось, сливочный сыр, свежий огурец (8 шт)', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7fd35a57-a554-45dd-a6ed-0a204d78fb44.webp'),
((SELECT id FROM rb), 'Ролл Калифорния', 'Снежный краб, авокадо, огурец, икра тобико (8 шт)', 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cb07f9fd-db1f-454f-aafa-43bb63b249aa.webp'),
((SELECT id FROM rb), 'Запеченный ролл с угрем', 'Угорь, сливочный сыр, унаги соус, кунжут', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5fb4fd23-002e-4277-9ae6-1d962c4d2af6.webp'),
((SELECT id FROM rb), 'Сет Студенческий', 'Филадельфия лайт, Калифорния, Каппа маки', 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8996b6ec-57ca-4722-b60f-25ae5b7f799a.webp');

-- 4. Добавляем блюда в "Subway"
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Subway' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url) VALUES
((SELECT id FROM rb), 'БМТ Саб (15 см)', 'Итальянская салями, пепперони, ветчина и свежие овощи', 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bf4a404a-f7f7-4277-b3e5-4d3a7becd569.webp'),
((SELECT id FROM rb), 'Саб Тунец (15 см)', 'Нежный тунец, смешанный с майонезом, и овощи на выбор', 390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/897f11c9-c9f3-4982-b3de-139b1b0113ae.webp'),
((SELECT id FROM rb), 'Кукис', 'Свежеиспеченное печенье с кусочками шоколада', 90000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d8aa921a-5d14-435d-97b7-90153f548fc5.webp');
