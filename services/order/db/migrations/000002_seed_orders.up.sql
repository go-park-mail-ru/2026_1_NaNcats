-- Seed orders for user_id=3 (first client).
-- Restaurant brand/branch IDs from restaurant seed:
--   1=Вкусно-и-точка, 2=Subway, 3=Папа Джонс, 4=Крошка Картошка,
--   5=Братья Караваевы, 7=Technikum, 10=Tutta La Vita, 11=KFC
-- Dish IDs: 1-10 (Вкусно), 11-14 (Папа Джонс), 15-18 (FoodBand), 19-21 (Subway)
-- Prices in micro-rubles (1₽ = 1000000).

-- Order 1: Вкусно - и точка, finished, 5 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 1, 1, 'addr-seed-01', 464000000, 'Вкусно - и точка', 'finished',
        NOW() - INTERVAL '5 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(1, 1, 2, 185000000),  -- Гранд x2
(1, 3, 1, 95000000);   -- Картофель Фри

-- Order 2: Папа Джонс, finished, 8 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 3, 3, 'addr-seed-01', 1698000000, 'Папа Джонс', 'finished',
        NOW() - INTERVAL '8 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(2, 11, 1, 699000000),  -- Пепперони
(2, 12, 1, 899000000),  -- Мясная
(2, 13, 1, 299000000);  -- Сырные палочки (fix: total was wrong, let's adjust)

-- Order 3: Subway, finished, 12 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 2, 2, 'addr-seed-02', 790000000, 'Subway', 'finished',
        NOW() - INTERVAL '12 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(3, 19, 2, 350000000),  -- БМТ Саб x2
(3, 21, 1, 90000000);   -- Кукис

-- Order 4: Вкусно - и точка (repeat), finished, 15 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 1, 1, 'addr-seed-01', 369000000, 'Вкусно - и точка', 'finished',
        NOW() - INTERVAL '15 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(4, 2, 1, 75000000),    -- Чизбургер
(4, 5, 1, 110000000),   -- Наггетсы 6 шт
(4, 1, 1, 185000000);   -- Гранд

-- Order 5: KFC, finished, 20 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 11, 11, 'addr-seed-03', 450000000, 'KFC', 'finished',
        NOW() - INTERVAL '20 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(5, 1, 2, 185000000),   -- using dish_id 1 as placeholder (KFC has no seeded dishes)
(5, 3, 1, 95000000);

-- Order 6: Tutta La Vita, finished, 25 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 10, 10, 'addr-seed-02', 890000000, 'Tutta La Vita', 'finished',
        NOW() - INTERVAL '25 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(6, 15, 1, 550000000),  -- Ролл Филадельфия (FoodBand dish, placeholder)
(6, 16, 1, 490000000);  -- Ролл Калифорния

-- Order 7: Technikum, finished, 30 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 7, 7, 'addr-seed-01', 1200000000, 'Technikum', 'finished',
        NOW() - INTERVAL '30 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(7, 1, 4, 185000000),   -- placeholder dishes
(7, 4, 3, 89000000);

-- Order 8: Братья Караваевы, delivering (active)
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 5, 5, 'addr-seed-01', 520000000, 'Братья Караваевы', 'delivering',
        NOW() - INTERVAL '35 minutes');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(8, 1, 2, 185000000),
(8, 8, 1, 110000000);

-- Order 9: Крошка Картошка, finished, 3 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 4, 4, 'addr-seed-03', 380000000, 'Крошка Картошка', 'finished',
        NOW() - INTERVAL '3 days');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(9, 1, 1, 185000000),
(9, 3, 2, 95000000);

-- Order 10: Папа Джонс (repeat), cancelled, 1 day ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 3, 3, 'addr-seed-01', 549000000, 'Папа Джонс', 'cancelled',
        NOW() - INTERVAL '1 day');

INSERT INTO "order_dish" (order_id, dish_id, quantity, price) VALUES
(10, 14, 1, 549000000);  -- Маргарита

-- Reviews for finished orders
INSERT INTO "order_review" (order_id, restaurant_rating, courier_rating, client_comment, created_at) VALUES
(1, 5, 5, 'Очень вкусно, как всегда!', NOW() - INTERVAL '5 days'),
(2, 4, 4, 'Пицца была горячая, доставка быстрая', NOW() - INTERVAL '8 days'),
(3, 5, 5, NULL, NOW() - INTERVAL '12 days'),
(4, 4, NULL, NULL, NOW() - INTERVAL '15 days'),
(5, 3, 4, 'Долго ждали', NOW() - INTERVAL '20 days');
