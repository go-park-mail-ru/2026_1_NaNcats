-- Удаляем категории, к которым не привязан ни один ресторан и ни одно блюдо.
-- Такие категории на главной — это «пустые» фильтры: нажимаешь, а ресторанов нет.
-- Условие через NOT EXISTS делает миграцию безопасной к FK (ON DELETE RESTRICT)
-- и независимой от окружения: удаляются ровно те категории, что пусты в этой БД.
DELETE FROM "category" c
WHERE NOT EXISTS (
        SELECT 1 FROM "restaurant_brand_category" rbc
        WHERE rbc.category_id = c.id
    )
    AND NOT EXISTS (
        SELECT 1 FROM "dish_category" dc
        WHERE dc.category_id = c.id
    );
