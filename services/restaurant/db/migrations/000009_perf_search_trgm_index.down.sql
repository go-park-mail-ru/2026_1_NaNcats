DROP INDEX IF EXISTS idx_restaurant_brand_description_trgm;
DROP INDEX IF EXISTS idx_restaurant_brand_name_trgm;
-- Расширение pg_trgm намеренно не удаляем: оно может использоваться другими индексами.
