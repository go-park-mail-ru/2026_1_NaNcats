-- Делаем владельца частью первичного ключа order_dish: одно блюдо у разных
-- участников совместного заказа хранится отдельными строками.
-- owner_user_id = 0 значит, что позиция ничья.

UPDATE "order_dish" SET owner_user_id = 0 WHERE owner_user_id IS NULL;

ALTER TABLE "order_dish" ALTER COLUMN owner_user_id SET DEFAULT 0;
ALTER TABLE "order_dish" ALTER COLUMN owner_user_id SET NOT NULL;

ALTER TABLE "order_dish" DROP CONSTRAINT order_dish_pkey;
ALTER TABLE "order_dish" ADD CONSTRAINT order_dish_pkey PRIMARY KEY (order_id, dish_id, owner_user_id);
