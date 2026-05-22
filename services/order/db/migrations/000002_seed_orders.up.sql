-- Наполнение заказами для user_id = 3 (первый клиент).
-- Restaurant brand/branch IDs из сида ресторанов:
--   1=Вкусно-и-точка, 2=Subway, 3=Папа Джонс, 4=Крошка Картошка,
--   5=Братья Караваевы, 7=Technikum, 10=Tutta La Vita, 11=KFC
-- Цены в микро-рублях (1₽ = 1000000).

-- Order 1: Вкусно - и точка, finished, 5 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 1, 1, 'addr-seed-01', 464000000, 'Вкусно - и точка', 'finished',
        NOW() - INTERVAL '5 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(1, 1, 'Гранд', 2, 185000000),
(1, 3, 'Картофель Фри', 1, 95000000);

-- Order 2: Папа Джонс, finished, 8 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 3, 3, 'addr-seed-01', 1698000000, 'Папа Джонс', 'finished',
        NOW() - INTERVAL '8 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(2, 11, 'Пицца Пепперони', 1, 699000000),
(2, 12, 'Пицца Мясная', 1, 899000000),
(2, 13, 'Сырные палочки', 1, 299000000);

-- Order 3: Subway, finished, 12 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 2, 2, 'addr-seed-02', 790000000, 'Subway', 'finished',
        NOW() - INTERVAL '12 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(3, 19, 'БМТ Саб', 2, 350000000),
(3, 21, 'Кукис', 1, 90000000);

-- Order 4: Вкусно - и точка (repeat), finished, 15 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 1, 1, 'addr-seed-01', 369000000, 'Вкусно - и точка', 'finished',
        NOW() - INTERVAL '15 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(4, 2, 'Чизбургер', 1, 75000000),
(4, 5, 'Наггетсы 6 шт', 1, 110000000),
(4, 1, 'Гранд', 1, 185000000);

-- Order 5: KFC, finished, 20 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 11, 11, 'addr-seed-03', 450000000, 'KFC', 'finished',
        NOW() - INTERVAL '20 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(5, 1, 'Острые крылышки', 2, 185000000),
(5, 3, 'Картофель Фри', 1, 95000000);

-- Order 6: Tutta La Vita, finished, 25 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 10, 10, 'addr-seed-02', 890000000, 'Tutta La Vita', 'finished',
        NOW() - INTERVAL '25 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(6, 15, 'Ролл Филадельфия', 1, 550000000),
(6, 16, 'Ролл Калифорния', 1, 490000000);

-- Order 7: Technikum, finished, 30 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 7, 7, 'addr-seed-01', 1200000000, 'Technikum', 'finished',
        NOW() - INTERVAL '30 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(7, 1, 'Паста Карбонара', 4, 185000000),
(7, 4, 'Цезарь с курицей', 3, 89000000);

-- Order 8: Братья Караваевы, delivering (active)
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 5, 5, 'addr-seed-01', 520000000, 'Братья Караваевы', 'delivering',
        NOW() - INTERVAL '35 minutes');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(8, 1, 'Сырник', 2, 185000000),
(8, 8, 'Борщ', 1, 110000000);

-- Order 9: Крошка Картошка, finished, 3 days ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 4, 4, 'addr-seed-03', 380000000, 'Крошка Картошка', 'finished',
        NOW() - INTERVAL '3 days');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(9, 1, 'Картошка с грибами', 1, 185000000),
(9, 3, 'Картофель Фри', 2, 95000000);

-- Order 10: Папа Джонс (repeat), cancelled, 1 day ago
INSERT INTO "order" (admin_account_id, restaurant_branch_id, restaurant_brand_id,
                     client_address_id, total_cost, restaurant_name, status, created_at)
VALUES (3, 3, 3, 'addr-seed-01', 549000000, 'Папа Джонс', 'cancelled',
        NOW() - INTERVAL '1 day');

INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price) VALUES
(10, 14, 'Пицца Маргарита', 1, 549000000);

-- Reviews for finished orders
INSERT INTO "order_review" (order_id, restaurant_rating, courier_rating, client_comment, created_at) VALUES
(1, 5, 5, 'Очень вкусно, как всегда!', NOW() - INTERVAL '5 days'),
(2, 4, 4, 'Пицца была горячая, доставка быстрая', NOW() - INTERVAL '8 days'),
(3, 5, 5, NULL, NOW() - INTERVAL '12 days'),
(4, 4, NULL, NULL, NOW() - INTERVAL '15 days'),
(5, 3, 4, 'Долго ждали', NOW() - INTERVAL '20 days');
