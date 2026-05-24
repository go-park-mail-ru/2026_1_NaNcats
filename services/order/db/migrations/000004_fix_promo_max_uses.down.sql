-- Восстанавливаем исходный max_uses = 1 для VKUSNO20 (как в миграции 000003).
-- current_uses намеренно не откатываем: это фактические данные использований.
UPDATE "promocode"
SET max_uses = 1
WHERE code = 'VKUSNO20';
