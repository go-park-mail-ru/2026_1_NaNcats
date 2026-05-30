INSERT INTO "category" (name) VALUES
('Бургеры'), ('Пицца'), ('Суши'), ('Завтраки'), ('Кофе'), ('Русская'), ('Китайская'), ('Фастфуд')
ON CONFLICT (name) DO NOTHING;

INSERT INTO "restaurant_brand" (owner_profile_id, name, description, promotion_tier, logo_url)
VALUES 
(1, 'Вкусно - и точка', 'Тот самый вкус', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/25253e87-0836-5d32-a0d7-a04f17711bda.webp'),
(1, 'Subway', 'Ешь свежее', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/2b4eeeb2-501b-4923-8238-9ab99ee8745f.webp'),
(1, 'Папа Джонс', 'Лучшие ингредиенты. Лучшая пицца.', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/bfb573dd-5ea1-4b52-adaf-17ac5eff5646.webp'),
(1, 'Крошка Картошка', 'Настоящая русская картошка', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/80eb7180-33f9-4841-b314-50372bbb6a9b.webp'),
(1, 'Братья Караваевы', 'Домашняя еда и выпечка', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/d000c551-0d84-4003-be37-3d5b21993ad5.webp'),
(1, 'FoodBand', 'Пицца и роллы 24/7', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/e6b14fa5-72dd-4f2d-b912-551b65ac5815.webp'),
(1, 'Technikum', 'Гастробистро от White Rabbit Family', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/16de116c-62c6-42d9-9e14-d08670a26e6c.webp'),
(1, 'Лепим и Варим', 'Лучшие пельмени в городе', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/280689c2-36da-4398-aba0-5844a348d464.webp'),
(1, 'Китайские новости', 'Настоящая китайская кухня', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/1c2079cd-edb6-4559-9127-437f54ede600.webp'),
(1, 'Tutta La Vita', 'Итальянская еда', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/8c8389dc-48d0-48ea-aa4a-030a920808d5.webp'),
(1, 'KFC', 'Вкусная курица', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/2d94c911-3db4-47a6-bc6b-597dac6a281a.webp'),
(1, 'Аист', 'Классика на Патриарших', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/c5c1ab16-1997-4a1a-afc8-681791eb7e91.webp'),
(1, 'Anna', 'Уютная кофейня', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/145cf370-37e4-445a-aa12-b551af28a3c9.webp'),
(1, 'Torro Grill', 'Стейки и гриль на углях', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/83b25c4a-06f9-512e-93da-6bcb76385a3b.webp'),
(1, 'Честная Рыба', 'Свежие морепродукты', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/f1fed5c4-2554-468f-b698-24ebf9abcb86.webp'),
(1, 'DiDi', 'Грузинская кухня', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/f493338c-e75d-441f-9206-2d4133bf1d9b.webp'),
(1, 'Eshak', 'Восточная кухня', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/9d44017a-feb1-42f0-8996-a035c2534fc8.webp'),
(1, 'Izumi', 'Японский стрит-фуд', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/aa58cd83-7527-475b-8ccf-0efd240481e7.webp'),
(1, 'Ketch Up', 'Авторские бургеры', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/81bb54e5-ad00-4a26-b9da-b3bc2da9db04.webp'),
(1, 'Moro', 'Средиземноморская кухня', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/c4a1e8dc-6a92-47c9-a9c6-83825e17a025.webp'),
(1, 'Pinsky Go', 'Быстро и вкусно', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/71700fe7-c5f5-40ad-903f-ee48d29eef38.webp'),
(1, 'Раменная Ru-Rik', 'Тот самый рамен', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/f5ec34a3-7bb8-4ea5-899f-3769eaf279f5.webp'),
(1, 'Империя Пиццы', 'Очень много пиццы!', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/9f018be0-410a-4f92-b7f3-a8fe6cdd2360.webp'),
(1, 'El Chapo Burgers Tacos&Burritos', 'Бургеры, тако и бурито - прямиком из Мексики', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/b2d14ebd-d49a-4ce8-b3de-2f72e147910d.webp'),
(1, 'Калифорния Дайнер', 'Что-то из Америки', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/321dd89a-04ef-4182-98e7-664d7ceec9dc.webp'),
(1, 'Машрумс', 'какие-то грибы', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/a65167f1-ad83-4a56-8ef2-5c50e696f05f.webp'),
(1, 'Руки ВВерх!', 'Поднимаем-поднимаем', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/94a58017-a27c-41b2-b8cd-78879ee353c0.webp'),
(1, 'Такахули', 'И ведь действительно', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/8cfe2e15-b368-4036-9db1-e5fb19be303f.webp'),
(1, 'Ванлав', 'Зеро френдс', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/0535dca3-d6c9-4e41-8be4-d8f0f58b25d4.webp'),
(1, 'Varvarka III', 'Будто император какой-то', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/fd308474-e54b-4f91-a6b4-e5cb9d5b786d.webp'),
(1, 'Villa Pasta', 'Опять макаронники со своей пиццей', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/9de17c18-38c5-40c4-9731-4d9911470757.webp')
ON CONFLICT (name) DO NOTHING;

INSERT INTO "restaurant_branch" (restaurant_brand_id, location_id, open_time, close_time)
SELECT b.id, 1, '00:00:00', '23:59:59'
FROM restaurant_brand b
WHERE b.id NOT IN (SELECT restaurant_brand_id FROM restaurant_branch)
ON CONFLICT DO NOTHING;

-- Вкусно и точка
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

-- Папа Джонс
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Папа Джонс' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url) VALUES
((SELECT id FROM rb), 'Пицца Пепперони (30 см)', 'Пикантная пепперони, томатный соус и сыр моцарелла', 699000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bac8f917-2762-4e65-9e61-fbc6a4ef54da.webp'),
((SELECT id FROM rb), 'Пицца Мясная (30 см)', 'Бекон, пепперони, ветчина, томатный соус и моцарелла', 899000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/78d0499a-7c23-49d4-824b-84955f9e6225.webp'),
((SELECT id FROM rb), 'Сырные палочки', 'Хрустящие палочки с сыром и чесночным соусом', 299000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2b77e55a-a7a7-4be3-98b0-578e159f0214.webp'),
((SELECT id FROM rb), 'Пицца Маргарита (30 см)', 'Традиционная пицца с томатным соусом и увеличенной порцией моцареллы', 549000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d7444179-ef0c-4dac-817d-70eeb62e4ca7.webp');

-- FoodBand
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'FoodBand' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url) VALUES
((SELECT id FROM rb), 'Ролл Филадельфия', 'Лосось, сливочный сыр, свежий огурец (8 шт)', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7fd35a57-a554-45dd-a6ed-0a204d78fb44.webp'),
((SELECT id FROM rb), 'Ролл Калифорния', 'Снежный краб, авокадо, огурец, икра тобико (8 шт)', 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cb07f9fd-db1f-454f-aafa-43bb63b249aa.webp'),
((SELECT id FROM rb), 'Запеченный ролл с угрем', 'Угорь, сливочный сыр, унаги соус, кунжут', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5fb4fd23-002e-4277-9ae6-1d962c4d2af6.webp'),
((SELECT id FROM rb), 'Сет Студенческий', 'Филадельфия лайт, Калифорния, Каппа маки', 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8996b6ec-57ca-4722-b60f-25ae5b7f799a.webp');

-- Subway
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Subway' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url) VALUES
((SELECT id FROM rb), 'БМТ Саб (15 см)', 'Итальянская салями, пепперони, ветчина и свежие овощи', 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bf4a404a-f7f7-4277-b3e5-4d3a7becd569.webp'),
((SELECT id FROM rb), 'Саб Тунец (15 см)', 'Нежный тунец, смешанный с майонезом, и овощи на выбор', 390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/897f11c9-c9f3-4982-b3de-139b1b0113ae.webp'),
((SELECT id FROM rb), 'Кукис', 'Свежеиспеченное печенье с кусочками шоколада', 90000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d8aa921a-5d14-435d-97b7-90153f548fc5.webp');
