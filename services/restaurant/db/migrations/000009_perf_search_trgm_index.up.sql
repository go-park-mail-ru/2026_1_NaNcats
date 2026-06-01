-- Оптимизация поиска ресторанов (итерация 1).
-- Запрос SearchRestaurantBrands использует:
--   WHERE name ILIKE '%q%' OR description ILIKE '%q%'
-- что без индекса даёт последовательное сканирование всей таблицы restaurant_brand.
-- Триграммные GIN-индексы (pg_trgm) позволяют планировщику выполнять
-- BitmapOr по двум индексам вместо Seq Scan.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_restaurant_brand_name_trgm
    ON "restaurant_brand" USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_restaurant_brand_description_trgm
    ON "restaurant_brand" USING gin (description gin_trgm_ops);
