-- Делаем владельца частью первичного ключа cart_dish: одно блюдо у разных
-- участников совместной корзины теперь хранится отдельными строками.
-- owner_user_id = 0 значит, что позиция ничья.

UPDATE "cart_dish" SET owner_user_id = 0 WHERE owner_user_id IS NULL;

ALTER TABLE "cart_dish" ALTER COLUMN owner_user_id SET DEFAULT 0;
ALTER TABLE "cart_dish" ALTER COLUMN owner_user_id SET NOT NULL;

ALTER TABLE "cart_dish" DROP CONSTRAINT cart_dish_pkey;
ALTER TABLE "cart_dish" ADD CONSTRAINT cart_dish_pkey PRIMARY KEY (cart_id, dish_id, owner_user_id);
