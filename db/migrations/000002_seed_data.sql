INSERT INTO "location" (address_text, latitude, longitude)
VALUES
    ('Москва, ул. Тверская, 3', 55.756877, 37.613709), -- ID 1 ресторан 1
    ('Москва, ул. Арбат, 10', 55.751680, 37.596639), -- ID 2 ресторан 2
    ('Москва, Госпитальный пер., 4/6', 55.766155, 37.682650), -- ID 3 адрес клиента 1
    ('Москва, Измайловский пр., 73', 55.793420, 37.795799); -- ID 4 адрес клиента 2

INSERT INTO "user" (phone, name, email, password_hash, user_role)
VALUES
    ('+7 (999) 111-11-11', 'Александр Рева', 'alex.client@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'client'),  -- ID 1
    ('+7 (999) 222-22-22', 'Сикс севен', 'six.client@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'client'),  -- ID 2
    ('+7 (999) 333-33-33', 'Олег Курьер', 'oleg.courier@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'courier'), -- ID 3
    ('+7 (999) 444-44-44', 'СвэгДрип Владелец', 'swagdrip.owner@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'owner'),   -- ID 4
    ('+7 (999) 555-55-55', 'Бензин Владелец', 'benzin.owner@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'owner');   -- ID 5

INSERT INTO "restaurant_brand" (name, description, promotion_tier)
VALUES
    ('Вкусно - и точка', 'Сеть предприятий быстрого обслуживания', 1), -- ID 1
    ('Бургер Кинг', 'Готовим на огне', 2);                             -- ID 2

INSERT INTO "client_profile" (account_id, bonus_balance, streak_count)
VALUES
    (1, 500, 3), -- Александр Рева
    (2, 100, 1); -- сикс севен

INSERT INTO "courier_profile" (account_id, status)
VALUES
    (3, 'waiting'); -- олежа в ожидании заказа

INSERT INTO "owner_profile" (account_id, restaurant_brand_id)
VALUES
    (4, 1), -- свэгдрип владеет вкусно и точка
    (5, 2); -- бензин владеет бургер кингом

INSERT INTO "restaurant_branch" (restaurant_brand_id, location_id, open_time, close_time)
VALUES
    (1, 1, '08:00:00', '22:00:00'), -- Филиал №1 (мак на тверской улице)
    (2, 2, '04:00:00', '23:59:59'); -- Филиал №2 (бк на арбате)

INSERT INTO "category" (name)
VALUES
    ('Бургеры'),      -- ID 1
    ('Напитки'),      -- ID 2
    ('Горячая пища'), -- ID 3
    ('Снэки');        -- ID 4

INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
VALUES
    (1, 1), (1, 2), (1, 3),
    (2, 1), (2, 3);

INSERT INTO "dish" (restaurant_brand_id, name, description, price)
VALUES
    (1, 'Биг Хит', 'Оч вкусный бургер бери его', 349),       -- ID 1
    (1, 'картошка фри', 'Оч вкусная картошка бери её', 119), -- ID 2
    (1, 'Пепси', 'жидкое золото!', 105),                     -- ID 3
    (2, 'Воппер', 'мясное золото!', 339);                    -- ID 4

INSERT INTO "dish_category" (dish_id, category_id)
VALUES
    (1, 1), (2, 3), -- биг хит (бургер); картошка фри (горячая пища)
    (3, 2), (4, 1); -- пепси (напитки) ; воппер (бургер)

INSERT INTO "client_address" (location_id, client_account_id, apartment, entrance, floor_level, label)
VALUES
    (3, 1, '67', '69', '52', 'Общага'),    -- Александр Рева очень интересно живет
    (4, 2, '420', '42', '1337', 'Общага'); -- Сикс Севен живет ещё интереснее

INSERT INTO "promocode" (code, discount_percent, discount_amount, is_global, expires_at)
VALUES
    ('WELCOME50', 50, NULL, TRUE, '2027-01-01 00:00:00+00'), -- скидка в процентах
    ('MINUS100', NULL, 100, TRUE, '2026-12-31 00:00:00+00'); -- скидка в рублях

INSERT INTO "order" (client_account_id, courier_account_id, restaurant_branch_id, client_address_id, promocode_id, status)
VALUES
    (1, 3, 1, 1, 1, 'delivering'),    -- ID 1, Рева заказал - олежа везет
    (2, NULL, 2, 2, NULL, 'waiting'); -- ID 2, Сикс Севен заказал - ждет назначения курьера, промокод не использовал

INSERT INTO "order_dish" (order_id, dish_id, quantity, price)
VALUES
    (1, 1, 2, 349), -- ЗАКАЗ #1: 2 биг хита по 349
    (1, 2, 1, 119), -- ЗАКАЗ #1: 1 картошка за 119
    (1, 3, 1, 105), -- ЗАКАЗ #1: 1 пепси за 105
    (2, 4, 2, 339); -- ЗАКАЗ #2: 2 воппера по 339

INSERT INTO "order_review" (order_id, restaurant_rating, courier_rating, client_comment)
VALUES
    (1, 5, 5, 'Я не верил, когда мне сказали, что мой заказ привезут на мини-погрузчике с отвалом для убора снега, НО ЭТО СЛУЧИЛОСЬ!');

