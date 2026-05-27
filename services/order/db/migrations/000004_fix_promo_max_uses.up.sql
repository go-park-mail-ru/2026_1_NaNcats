-- Пересчитываем current_uses по фактическим записям в promocode_usage.
-- Исправляет расхождение при возможном сбое rollback в саге оплаты.
UPDATE "promocode" p
SET current_uses = (
    SELECT COUNT(*) FROM "promocode_usage" pu
    WHERE pu.promocode_id = p.id
);

-- VKUSNO20 — «скидка 20% на первый заказ» каждого пользователя.
-- Исходный max_uses = 1 задаёт глобальный лимит (одна штука на весь сервис),
-- хотя уникальность «первый заказ» на пользователя уже обеспечивается
-- таблицей promocode_usage (CountUserPromocodeUsage в usecase).
-- Снимаем глобальный лимит: NULL = без ограничений по числу пользователей.
UPDATE "promocode"
SET max_uses = NULL
WHERE code = 'VKUSNO20';
