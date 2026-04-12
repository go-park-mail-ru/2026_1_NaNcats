-- 1. Категории
INSERT INTO "category" (name) VALUES
('Бургеры'), ('Пицца'), ('Суши'), ('Завтраки'), ('Кофе'), ('Русская'), ('Китайская'), ('Фастфуд')
ON CONFLICT (name) DO NOTHING;

INSERT INTO "user" (name, email, password_hash, user_role)
VALUES ('Главный Ресторатор', 'owner@foodcourt.fun', '$2a$10$12345678901234567890123456789012345678901234567890123', 'owner')
ON CONFLICT (email) DO NOTHING;

-- Привязываем профиль владельца
INSERT INTO "owner_profile" (account_id)
SELECT id FROM "user" WHERE email = 'owner@foodcourt.fun'
ON CONFLICT DO NOTHING;

-- 2. Локации
INSERT INTO "location" (address_text, coordinate) VALUES
('Москва, ул. Арбат, 1', ST_SetSRID(ST_MakePoint(37.599, 55.752), 4326)),
('Москва, Цветной бульвар, 15', ST_SetSRID(ST_MakePoint(37.622, 55.771), 4326)),
('Москва, ул. Тверская, 7', ST_SetSRID(ST_MakePoint(37.611, 55.758), 4326))
ON CONFLICT DO NOTHING;

-- 3. Бренды
WITH owner AS (SELECT account_id FROM owner_profile LIMIT 1)
INSERT INTO "restaurant_brand" (owner_profile_id, name, description, promotion_tier, logo_url)
VALUES 
((SELECT account_id FROM owner), 'Вкусно - и точка', 'Тот самый вкус', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/e5f0825c-c82d-4f8d-b9e0-d0149a38322f.webp'),
((SELECT account_id FROM owner), 'Subway', 'Ешь свежее', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/2b4eeeb2-501b-4923-8238-9ab99ee8745f.webp'),
((SELECT account_id FROM owner), 'Папа Джонс', 'Лучшие ингредиенты. Лучшая пицца.', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/bfb573dd-5ea1-4b52-adaf-17ac5eff5646.webp'),
((SELECT account_id FROM owner), 'Крошка Картошка', 'Настоящая русская картошка', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/80eb7180-33f9-4841-b314-50372bbb6a9b.webp'),
((SELECT account_id FROM owner), 'Братья Караваевы', 'Домашняя еда и выпечка', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/d000c551-0d84-4003-be37-3d5b21993ad5.webp'),
((SELECT account_id FROM owner), 'FoodBand', 'Пицца и роллы 24/7', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/e6b14fa5-72dd-4f2d-b912-551b65ac5815.webp'),
((SELECT account_id FROM owner), 'Technikum', 'Гастробистро от White Rabbit Family', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/16de116c-62c6-42d9-9e14-d08670a26e6c.webp'),
((SELECT account_id FROM owner), 'Лепим и Варим', 'Лучшие пельмени в городе', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/280689c2-36da-4398-aba0-5844a348d464.webp'),
((SELECT account_id FROM owner), 'Китайские новости', 'Настоящая китайская кухня', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/1c2079cd-edb6-4559-9127-437f54ede600.webp'),
((SELECT account_id FROM owner), 'Tutta La Vita', 'Итальянская еда', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/8c8389dc-48d0-48ea-aa4a-030a920808d5.webp'),
((SELECT account_id FROM owner), 'KFC', 'Вкусная курица', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/2d94c911-3db4-47a6-bc6b-597dac6a281a.webp'),
((SELECT account_id FROM owner), 'Аист', 'Классика на Патриарших', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/c5c1ab16-1997-4a1a-afc8-681791eb7e91.webp'),
((SELECT account_id FROM owner), 'Anna', 'Уютная кофейня', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/145cf370-37e4-445a-aa12-b551af28a3c9.webp'),
((SELECT account_id FROM owner), 'Arium Grill', 'Мясо на углях', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/8d3b2425-1130-4fae-8ac1-2090f27280b1.webp'),
((SELECT account_id FROM owner), 'Честная Рыба', 'Свежие морепродукты', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/f1fed5c4-2554-468f-b698-24ebf9abcb86.webp'),
((SELECT account_id FROM owner), 'DiDi', 'Грузинская кухня', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/f493338c-e75d-441f-9206-2d4133bf1d9b.webp'),
((SELECT account_id FROM owner), 'Eshak', 'Восточная кухня', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/9d44017a-feb1-42f0-8996-a035c2534fc8.webp'),
((SELECT account_id FROM owner), 'Izumi', 'Японский стрит-фуд', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/aa58cd83-7527-475b-8ccf-0efd240481e7.webp'),
((SELECT account_id FROM owner), 'Ketch Up', 'Авторские бургеры', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/81bb54e5-ad00-4a26-b9da-b3bc2da9db04.webp'),
((SELECT account_id FROM owner), 'Moro', 'Средиземноморская кухня', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/c4a1e8dc-6a92-47c9-a9c6-83825e17a025.webp'),
((SELECT account_id FROM owner), 'Pinsky Go', 'Быстро и вкусно', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/71700fe7-c5f5-40ad-903f-ee48d29eef38.webp'),
((SELECT account_id FROM owner), 'Раменная Ru-Rik', 'Тот самый рамен', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/f5ec34a3-7bb8-4ea5-899f-3769eaf279f5.webp'),
((SELECT account_id FROM owner), 'Империя Пиццы', 'Очень много пиццы!', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/9f018be0-410a-4f92-b7f3-a8fe6cdd2360.webp'),
((SELECT account_id FROM owner), 'El Chapo Burgers Tacos&Burritos', 'Бургеры, тако и бурито - прямиком из Мексики', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/b2d14ebd-d49a-4ce8-b3de-2f72e147910d.webp'),
((SELECT account_id FROM owner), 'Калифорния Дайнер', 'Что-то из Америки', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/321dd89a-04ef-4182-98e7-664d7ceec9dc.webp'),
((SELECT account_id FROM owner), 'Машрумс', 'какие-то грибы', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/a65167f1-ad83-4a56-8ef2-5c50e696f05f.webp'),
((SELECT account_id FROM owner), 'Руки ВВерх!', 'Поднимаем-поднимаем', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/94a58017-a27c-41b2-b8cd-78879ee353c0.webp'),
((SELECT account_id FROM owner), 'Такахули', 'И ведь действительно', 3, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/8cfe2e15-b368-4036-9db1-e5fb19be303f.webp'),
((SELECT account_id FROM owner), 'Ванлав', 'Зеро френдс', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/0535dca3-d6c9-4e41-8be4-d8f0f58b25d4.webp'),
((SELECT account_id FROM owner), 'Varvarka III', 'Будто император какой-то', 2, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/fd308474-e54b-4f91-a6b4-e5cb9d5b786d.webp'),
((SELECT account_id FROM owner), 'Villa Pasta', 'Опять макаронники со своей пиццей', 1, 'https://nancats-bucket.storage.yandexcloud.net/restaurants/9de17c18-38c5-40c4-9731-4d9911470757.webp')
ON CONFLICT (name) DO NOTHING;

-- 4. Создаем филиалы (физические точки), чтобы рестораны появились в списке
INSERT INTO "restaurant_branch" (restaurant_brand_id, location_id, open_time, close_time)
SELECT b.id, (SELECT id FROM location LIMIT 1), '00:00:00', '23:59:59'
FROM restaurant_brand b
WHERE b.id NOT IN (SELECT restaurant_brand_id FROM restaurant_branch)
ON CONFLICT DO NOTHING;
