-- Возврат к ключу (order_id, dish_id). Если у блюда несколько строк,
-- оставляем строку с наибольшим количеством.
DELETE FROM "order_dish" od
USING "order_dish" other
WHERE od.order_id = other.order_id
  AND od.dish_id = other.dish_id
  AND (od.quantity < other.quantity
       OR (od.quantity = other.quantity AND od.owner_user_id > other.owner_user_id));

ALTER TABLE "order_dish" DROP CONSTRAINT order_dish_pkey;
ALTER TABLE "order_dish" ADD CONSTRAINT order_dish_pkey PRIMARY KEY (order_id, dish_id);

ALTER TABLE "order_dish" ALTER COLUMN owner_user_id DROP NOT NULL;
ALTER TABLE "order_dish" ALTER COLUMN owner_user_id DROP DEFAULT;
UPDATE "order_dish" SET owner_user_id = NULL WHERE owner_user_id = 0;
