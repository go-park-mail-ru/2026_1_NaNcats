-- Откат seed-данных: чистим связи и сбрасываем emoji в дефолт.
-- Сами категории не удаляем — это сделает down 000001.
DELETE FROM "restaurant_brand_category" WHERE TRUE;
UPDATE "category" SET emoji = '';
