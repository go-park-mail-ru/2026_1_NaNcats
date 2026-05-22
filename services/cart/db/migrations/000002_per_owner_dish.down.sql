-- Возврат к ключу (cart_id, dish_id). Если у блюда несколько строк, оставляем
-- строку с наибольшим количеством: иначе старый ключ восстановить нельзя.
DELETE FROM "cart_dish" cd
USING "cart_dish" other
WHERE cd.cart_id = other.cart_id
  AND cd.dish_id = other.dish_id
  AND (cd.quantity < other.quantity
       OR (cd.quantity = other.quantity AND cd.owner_user_id > other.owner_user_id));

ALTER TABLE "cart_dish" DROP CONSTRAINT cart_dish_pkey;
ALTER TABLE "cart_dish" ADD CONSTRAINT cart_dish_pkey PRIMARY KEY (cart_id, dish_id);

ALTER TABLE "cart_dish" ALTER COLUMN owner_user_id DROP NOT NULL;
ALTER TABLE "cart_dish" ALTER COLUMN owner_user_id DROP DEFAULT;
UPDATE "cart_dish" SET owner_user_id = NULL WHERE owner_user_id = 0;
