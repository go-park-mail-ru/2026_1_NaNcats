INSERT INTO "location" (address_text, coordinate)
VALUES
    ('Москва, ул. Тверская, 15', ST_SetSRID(ST_MakePoint(37.609420, 55.761502), 4326)),       -- ID 1: ресторан Burger House
    ('Москва, ул. Новый Арбат, 21', ST_SetSRID(ST_MakePoint(37.584980, 55.751823), 4326)),    -- ID 2: ресторан Pizza & Pasta
    ('Москва, Ленинградский пр., 39', ST_SetSRID(ST_MakePoint(37.535890, 55.797241), 4326)),  -- ID 3: адрес клиента 1 (Иван)
    ('Москва, пр. Мира, 119', ST_SetSRID(ST_MakePoint(37.628861, 55.830607), 4326));          -- ID 4: адрес клиента 2 (Анна)
    
INSERT INTO "user" (phone, name, email, password_hash, user_role)
VALUES
    ('+7 (900) 111-11-11', 'Иван Иванов', 'ivan.client@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'client'),    		-- ID 1
    ('+7 (900) 222-22-22', 'Анна Смирнова', 'anna.client@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'client'),  		-- ID 2
    ('+7 (900) 333-33-33', 'Петр Быстрый', 'petr.courier@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'courier'), 		-- ID 3
    ('+7 (900) 444-44-44', 'Алексей Медленный', 'alex.courier@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'courier'), -- ID 4
    ('+7 (900) 555-55-55', 'Сергей Бургеров', 'sergey.owner@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'owner'),   	-- ID 5
    ('+7 (900) 666-66-66', 'Мария Пиццева', 'maria.owner@mail.ru', '$2a$10$12345678901234567890123456789012345678901234567890123', 'owner');    	-- ID 6

INSERT INTO "category" (name)
VALUES
    ('Бургеры'),      -- ID 1
    ('Пицца'),        -- ID 2
    ('Напитки'),      -- ID 3
    ('Закуски'),      -- ID 4
    ('Десерты');      -- ID 5

INSERT INTO "client_profile" (account_id, bonus_balance, streak_count)
VALUES
    (1, 1500, 5), -- Иван имеет 1500 бонусов
    (2, 0, 1);    -- а Аня просто плохая

INSERT INTO "courier_profile" (account_id, status)
VALUES
    (3, 'waiting'),    -- Петр ждет заказов
    (4, 'delivering'); -- Алексей уже везет заказ

INSERT INTO "owner_profile" (account_id)
VALUES
    (5), -- Сергей
    (6); -- Мария
    
INSERT INTO "restaurant_brand" (owner_profile_id, name, description, promotion_tier)
VALUES
    (5, 'Burger House', 'Лучшие крафтовые бургеры из мраморной говядины', 2), 	-- ID 1 (Сергей)
    (6, 'Pizza & Pasta', 'Настоящая неаполитанская пицца из дровяной печи', 1); -- ID 2 (Мария)
    
INSERT INTO "restaurant_brand_category" (restaurant_brand_id, category_id)
VALUES
    (1, 1), (1, 3), (1, 4), -- Burger House продает бургеры, напитки и закуски
    (2, 2), (2, 3), (2, 5); -- Pizza & Pasta продает пиццу, напитки и десерты

INSERT INTO "restaurant_branch" (restaurant_brand_id, location_id, open_time, close_time)
VALUES
    (1, 1, '10:00:00', '23:00:00'), -- Burger House на Тверской
    (2, 2, '09:00:00', '22:00:00'); -- Pizza & Pasta на Новом Арбате
    
INSERT INTO "dish" (restaurant_brand_id, name, description, price)
VALUES
    (1, 'Классический Чизбургер', 'Мраморная говядина, чеддер, томаты, фирменный соус', 450000000), -- ID 1 (450 руб)
    (1, 'Картофель Фри', 'Хрустящий картофель с морской солью', 150000000),                       	-- ID 2 (150 руб)
    (1, 'Кока-Кола Добрый', 'Холодная газировка 0.5', 100000000),                                 	-- ID 3 (100 руб)
    (2, 'Маргарита', 'Томатный соус, моцарелла, базилик', 550000000),                             	-- ID 4 (550 руб)
    (2, 'Тирамису', 'Классический итальянский десерт', 350000000);                                	-- ID 5 (350 руб)
    
INSERT INTO "dish_category" (dish_id, category_id)
VALUES
    (1, 1), -- Чизбургер -> Бургеры
    (2, 4), -- Картофель Фри -> Закуски
    (3, 3), -- Кола -> Напитки
    (4, 2), -- Маргарита -> Пицца
    (5, 5); -- Тирамису -> Десерты
    
INSERT INTO "client_address" (location_id, client_account_id, apartment, entrance, floor_level, label)
VALUES
    (3, 1, '142', '3', '12', 'Дом'),    -- Иван заказывает на Ленинградку
    (4, 2, '15', '1', '4', 'Работа');   -- Анна на проспект Мира

INSERT INTO "promocode" (code, discount_percent, discount_amount, is_global, expires_at)
VALUES
    ('SALE20', 20, NULL, TRUE, '2027-12-31 23:59:59+00'),      		-- Скидка 20%
    ('MINUS500', NULL, 500000000, TRUE, '2026-12-31 23:59:59+00'); 	-- Скидка 500 рублей

INSERT INTO "order" (client_account_id, courier_account_id, restaurant_branch_id, client_address_id, total_cost, promocode_id, status)
VALUES
    (1, 3, 1, 1, 1050000000, NULL, 'finished'),   -- ID 1: Иван заказал из Burger House, заказ завершен
    (2, NULL, 2, 2, 900000000, NULL, 'waiting'),  -- ID 2: Анна заказала из Pizza & Pasta, ждет курьера
    (1, 4, 1, 1, 600000000, 1, 'delivering');     -- ID 3: Иван снова заказал из Burger House, в пути (везет Алексей)
    
INSERT INTO "order_dish" (order_id, dish_id, quantity, price)
VALUES
    (1, 1, 2, 450000000), -- Заказ 1: 2 чизбургера (900 руб)
    (1, 2, 1, 150000000), -- Заказ 1: 1 картофель (150 руб) -> Итог: 1050 руб
    (2, 4, 1, 550000000), -- Заказ 2: 1 пицца (550 руб)
    (2, 5, 1, 350000000), -- Заказ 2: 1 тирамису (350 руб) -> Итог: 900 руб
    (3, 1, 1, 450000000), -- Заказ 3: 1 чизбургер (450 руб)
    (3, 2, 1, 150000000); -- Заказ 3: 1 картофель (150 руб) -> Итог: 600 руб
    
INSERT INTO "order_review" (order_id, restaurant_rating, courier_rating, client_comment)
VALUES
    (1, 5, 5, 'Лучшие бургеры в городе, курьер приехал даже раньше времени!');

INSERT INTO "achievement" (code, name, description)
VALUES
    ('FIRST_ORDER', 'Первый укус', 'Оформите свой первый успешный заказ'),
    ('BURGER_LOVER', 'Бургерный Монстр', 'Закажите более 10 бургеров за месяц');

INSERT INTO "user_achievement" (achievement_id, user_id)
VALUES
    (1, 1); -- Иван получил достижение за первый заказ
    
    